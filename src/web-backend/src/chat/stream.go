package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type streamRequest struct {
	Prompt string `json:"prompt"`
}

type dashscopeChatCompletionsRequest struct {
	Model         string                   `json:"model"`
	Messages      []dashscopeChatMessage    `json:"messages"`
	Stream        bool                     `json:"stream"`
	StreamOptions *dashscopeStreamOptions   `json:"stream_options,omitempty"`
}

type dashscopeStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type dashscopeChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI-compatible streaming chunk (DashScope compatible-mode).
// We only care about delta.content.
type dashscopeStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// StreamHandler is a standard HTTP handler (http_no_protocol style) that streams
// response chunks to the browser via chunked transfer encoding.
//
// Frontend reads it with: res.body.getReader() + TextDecoder.
func StreamHandler(w http.ResponseWriter, r *http.Request) error {
	// Preflight support (CORS headers are set by the server filter in main.go).
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("method not allowed"))
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	var req streamRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid json"))
		return nil
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("prompt is empty"))
		return nil
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("streaming unsupported: ResponseWriter is not a Flusher")
	}

	// Make sure proxies (e.g. nginx) don't buffer the response.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// IMPORTANT: do NOT hard-code API keys here. Read from environment.
	apiKey := strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	if apiKey == "" {
		_, _ = w.Write([]byte("后端未配置 LLM_API_KEY（请先在运行环境设置该环境变量）\n"))
		flusher.Flush()
		return nil
	}

	apiURL := strings.TrimSpace(os.Getenv("LLM_API_URL"))
	if apiURL == "" {
		_, _ = w.Write([]byte("后端未配置 LLM_API_URL（例如 DashScope compatible-mode 的 chat/completions 地址）\n"))
		flusher.Flush()
		return nil
	}

	return proxyDashScopeStream(r.Context(), w, flusher, apiKey, apiURL, prompt)
}

func proxyDashScopeStream(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, apiKey, apiURL, prompt string) error {
	// Build upstream request body.
	upReqBody := dashscopeChatCompletionsRequest{
		Model: "qwen-plus",
		Messages: []dashscopeChatMessage{
			{Role: "user", Content: prompt},
		},
		Stream:        true,
		StreamOptions: &dashscopeStreamOptions{IncludeUsage: true},
	}
	b, err := json.Marshal(upReqBody)
	if err != nil {
		return err
	}

	upReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apiURL,
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}
	upReq.Header.Set("Authorization", "Bearer "+apiKey)
	upReq.Header.Set("Content-Type", "application/json")
	upReq.Header.Set("Accept", "text/event-stream")

	// Use a transport with sane timeouts; keep response streaming.
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   8 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   8 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		// Do not set Timeout here; it would cancel long streaming responses.
	}

	upResp, err := client.Do(upReq)
	if err != nil {
		_, _ = w.Write([]byte("上游请求失败：无法连接到 DashScope\n"))
		flusher.Flush()
		return nil
	}
	defer upResp.Body.Close()

	if upResp.StatusCode/100 != 2 {
		errBody, _ := io.ReadAll(io.LimitReader(upResp.Body, 8<<10))
		_, _ = w.Write([]byte(fmt.Sprintf("上游返回错误：HTTP %d\n%s\n", upResp.StatusCode, string(errBody))))
		flusher.Flush()
		return nil
	}

	// Parse SSE from upstream. For each event, extract delta.content and stream it to client.
	br := bufio.NewReaderSize(upResp.Body, 64<<10)
	var dataLines []string
	flushEvent := func() {
		if len(dataLines) == 0 {
			return
		}
		data := strings.Join(dataLines, "\n")
		dataLines = dataLines[:0]
		data = strings.TrimSpace(data)
		if data == "" {
			return
		}
		if data == "[DONE]" {
			return
		}
		var chunk dashscopeStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			// If the upstream sends something unexpected, ignore it instead of breaking the stream.
			return
		}
		if len(chunk.Choices) == 0 {
			return
		}
		t := chunk.Choices[0].Delta.Content
		if t == "" {
			return
		}
		_, _ = w.Write([]byte(t))
		flusher.Flush()
	}

	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			s := strings.TrimRight(line, "\r\n")
			if s == "" {
				flushEvent()
			} else if strings.HasPrefix(s, "data:") {
				dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(s, "data:")))
			}
			// Ignore: "event:", "id:", "retry:" etc.
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				flushEvent()
				return nil
			}
			return nil
		}
	}
}
