<script setup>
import { nextTick, onBeforeUnmount, onMounted, reactive, ref } from "vue";
import MarkdownIt from "markdown-it";
import DOMPurify from "dompurify";

const md = new MarkdownIt({
  linkify: true,
  breaks: true,
  // Keep HTML off; we only accept Markdown and sanitize the rendered output anyway.
  html: false,
});

const draft = ref("");
const sending = ref(false);
const messages = ref([
  {
    id: 1,
    role: "assistant",
    text: "你好，我是 LLYB AI（Demo）。你可以先随便问我一句。",
    html: "", // filled below
    streaming: false,
    ts: Date.now(),
  },
]);

let nextID = 2;
const listEl = ref(null);
const inputEl = ref(null);

// If user scrolls up to read history, don't force-scroll them to the bottom.
const stickyToBottom = ref(true);
const updateSticky = () => {
  const el = listEl.value;
  if (!el) return;
  const thresholdPx = 32;
  const distance = el.scrollHeight - el.scrollTop - el.clientHeight;
  stickyToBottom.value = distance <= thresholdPx;
};

const apiBase = (() => {
  const v = typeof __BACKEND_HOST__ === "string" ? __BACKEND_HOST__.trim() : "";
  return v ? v.replace(/\/+$/, "") : "";
})();

const renderMarkdown = (text) => {
  const raw = md.render(String(text ?? ""));
  // Defense-in-depth: sanitize rendered HTML to avoid XSS.
  return DOMPurify.sanitize(raw);
};

// Fill initial message HTML once (avoid re-rendering Markdown on every chunk).
messages.value[0].html = renderMarkdown(messages.value[0].text);

const scrollToBottom = async () => {
  await nextTick();
  const el = listEl.value;
  if (!el) return;
  // Only autoscroll if user hasn't intentionally scrolled up.
  if (stickyToBottom.value) el.scrollTop = el.scrollHeight;
};

const focusInput = async () => {
  await nextTick();
  // Don't steal focus from other inputs/modals if they appear later.
  inputEl.value?.focus?.();
  // Put caret at end.
  const el = inputEl.value;
  if (el && typeof el.setSelectionRange === "function") {
    const len = String(el.value ?? "").length;
    el.setSelectionRange(len, len);
  }
};

// Detect ``` fences even if they are split across chunk boundaries.
// We keep a 2-char tail buffer and scan (tail + appended) for "```".
// Returns whether we closed a fence during this update.
const updateFenceState = (state, appended) => {
  const wasOpen = !!state.fenceOpen;
  const s = (state.tail || "") + String(appended || "");
  let i = 0;
  while (true) {
    const idx = s.indexOf("```", i);
    if (idx === -1) break;
    state.fenceOpen = !state.fenceOpen;
    i = idx + 3;
  }
  state.tail = s.slice(-2);
  return wasOpen && !state.fenceOpen;
};

// "Paragraph boundary" = a blank line ("\n\n") outside code fences.
// Keep a 1-char tail buffer so boundaries spanning chunks are detected.
const hitParagraphBoundary = (state, appended) => {
  const s = (state.paraTail || "") + String(appended || "");
  state.paraTail = s.slice(-1);
  return s.includes("\n\n");
};

const send = async () => {
  const text = draft.value.trim();
  if (!text || sending.value) return;
  if (!apiBase) {
    const msg = "未配置 BACKEND_HOST，无法请求后端流式接口。";
    messages.value.push({
      id: nextID++,
      role: "assistant",
      text: msg,
      html: renderMarkdown(msg),
      streaming: false,
      ts: Date.now(),
    });
    await scrollToBottom();
    return;
  }
  sending.value = true;

  messages.value.push({
    id: nextID++,
    role: "user",
    text,
    html: renderMarkdown(text),
    streaming: false,
    ts: Date.now(),
  });
  draft.value = "";
  await scrollToBottom();

  // Stream from backend: POST /ai/chat/stream, read chunks via ReadableStream.
  // Use a reactive object so mutations (assistant.text += ...) trigger UI updates.
  const assistant = reactive({
    id: nextID++,
    role: "assistant",
    text: "",
    html: "",
    streaming: true,
    ts: Date.now(),
    // internal stream rendering state
    _md: { fenceOpen: false, tail: "", paraTail: "", lastRender: 0, renderedLen: 0 },
  });
  messages.value.push(assistant);
  // Force-scroll for a new assistant message start.
  stickyToBottom.value = true;
  await scrollToBottom();

  const controller = new AbortController();
  activeController.value = controller;
  // Streaming responses can be long; keep this generously high.
  // If you want a "Stop" button later, we can expose controller.abort() to UI.
  const timer = setTimeout(() => controller.abort(), 5 * 60_000);
  try {
    const res = await fetch(`${apiBase}/ai/chat/stream`, {
      method: "POST",
      headers: { "Content-Type": "text/plain" },
      body: JSON.stringify({ prompt: text }),
      signal: controller.signal,
    });
    if (!res.ok) throw new Error(`stream http ${res.status}`);
    if (!res.body) throw new Error("浏览器不支持流式响应（ReadableStream 不可用）");

    const reader = res.body.getReader();
    const decoder = new TextDecoder("utf-8");
    // Keep reading responsive; always append plain text quickly.
    // Only re-render Markdown at paragraph boundaries and on code-fence closure.
    let pending = "";
    let lastFlush = 0;
    const flushEveryMs = 50;
    const flush = (forceRender = false) => {
      if (!pending) return;
      const appended = pending;
      assistant.text += appended;
      pending = "";

      const closedFence = updateFenceState(assistant._md, appended);
      const boundaryHit = !assistant._md.fenceOpen && hitParagraphBoundary(assistant._md, appended);

      const now = performance.now();
      let renderLen = 0;
      if (forceRender || closedFence) {
        renderLen = assistant.text.length;
      } else if (boundaryHit) {
        // Render only up to the last paragraph boundary, so we never split a paragraph
        // between Markdown (m.html) and the streaming tail (plain text).
        const idx = assistant.text.lastIndexOf("\n\n");
        if (idx !== -1) renderLen = idx + 2;
      }

      if (renderLen > assistant._md.renderedLen) {
        assistant.html = renderMarkdown(assistant.text.slice(0, renderLen));
        assistant._md.lastRender = now;
        assistant._md.renderedLen = renderLen;
      }

      // Don't await; keep the read loop responsive.
      scrollToBottom();
    };
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      if (value) {
        pending += decoder.decode(value, { stream: true });
        const now = performance.now();
        if (now - lastFlush >= flushEveryMs) {
          lastFlush = now;
          flush();
        }
      }
    }
    // Flush remaining decoded text (and any internal decoder state).
    pending += decoder.decode();
    flush(true);
    assistant.streaming = false;
    assistant.html = renderMarkdown(assistant.text);
    stickyToBottom.value = true;
    await scrollToBottom();
  } catch (e) {
    const aborted = e?.name === "AbortError";
    assistant.text = assistant.text || (aborted ? "已中止。" : `流式请求失败：${e?.message || "unknown error"}`);
    assistant.streaming = false;
    assistant.html = renderMarkdown(assistant.text);
  } finally {
    clearTimeout(timer);
    if (activeController.value === controller) activeController.value = null;
    sending.value = false;
    // Put focus back to input for quick follow-ups.
    await nextTick();
    inputEl.value?.focus?.();
  }
};

const activeController = ref(null);
const stop = () => {
  const c = activeController.value;
  if (c) c.abort();
};

// Allow typing immediately without clicking into the textarea.
// If the user is not focused on an editable element, we capture printable keys
// and append them into draft, then focus the input.
const onGlobalKeydown = (e) => {
  if (e.defaultPrevented) return;
  if (e.ctrlKey || e.metaKey || e.altKey) return;
  if (e.key === "Escape" || e.key === "Tab") return;

  const ae = document.activeElement;
  const isEditable =
    ae &&
    (ae === inputEl.value ||
      ae.tagName === "INPUT" ||
      ae.tagName === "TEXTAREA" ||
      ae.isContentEditable === true);
  if (isEditable) return;

  // Only hijack "typing" keys.
  if (e.key === "Backspace") {
    e.preventDefault();
    draft.value = draft.value.slice(0, -1);
    focusInput();
    return;
  }
  if (e.key.length === 1) {
    e.preventDefault();
    draft.value += e.key;
    focusInput();
  }
};

const onKeydown = (e) => {
  // Enter to send; Shift+Enter for newline
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    if (sending.value) stop();
    else send();
  }
};

onMounted(() => {
  const el = listEl.value;
  if (!el) return;
  el.addEventListener("scroll", updateSticky, { passive: true });
  updateSticky();
  window.addEventListener("keydown", onGlobalKeydown, true);
  focusInput();
});

onBeforeUnmount(() => {
  const el = listEl.value;
  if (!el) return;
  el.removeEventListener("scroll", updateSticky);
  window.removeEventListener("keydown", onGlobalKeydown, true);
});
</script>

<template>
  <div class="chat">
    <header class="chat-header">
      <div class="chat-title">AI 推理</div>
      <div class="chat-sub">流式输出：已接入后端 /ai/chat/stream</div>
    </header>

    <div ref="listEl" class="chat-list" role="log" aria-live="polite">
      <div
        v-for="m in messages"
        :key="m.id"
        class="msg"
        :class="m.role === 'user' ? 'msg-user' : 'msg-assistant'"
      >
        <div class="bubble">
          <!-- Streaming: show Markdown for completed paragraphs + plain text tail for the current paragraph. -->
          <template v-if="m.streaming">
            <div v-if="m.html" class="markdown" v-html="m.html"></div>
            <pre
              v-if="m.text && m._md && m._md.renderedLen < m.text.length"
              class="plain tail"
              v-text="m.text.slice(m._md.renderedLen)"
            ></pre>
            <pre v-else-if="!m.html" class="plain" v-text="m.text"></pre>
          </template>
          <div v-else class="markdown" v-html="m.html"></div>
        </div>
      </div>
    </div>

    <footer class="chat-input">
      <textarea
        v-model="draft"
        ref="inputEl"
        class="input"
        rows="1"
        placeholder="输入内容，Enter 发送，Shift+Enter 换行"
        @keydown="onKeydown"
      ></textarea>
      <button
        class="send"
        type="button"
        :class="{ stopping: sending }"
        :disabled="(!sending && !draft.trim())"
        :aria-label="sending ? '中止' : '发送'"
        @click="sending ? stop() : send()"
      >
        <svg v-if="!sending" class="send-icon" viewBox="0 0 24 24" aria-hidden="true">
          <path
            d="M12 4l5.5 6.2a1 1 0 0 1-.75 1.67H14v7a1 1 0 0 1-2 0v-7H7.25A1 1 0 0 1 6.5 10.2L12 4z"
            fill="currentColor"
          />
        </svg>
        <svg v-else class="send-icon" viewBox="0 0 24 24" aria-hidden="true">
          <rect x="7" y="7" width="10" height="10" rx="1.5" fill="currentColor" />
        </svg>
      </button>
    </footer>
  </div>
</template>

<style scoped>
.chat {
  height: calc(100vh - 56px - 36px); /* topbar + main padding */
  min-height: 520px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  /* Center the whole chat UI and make it wider within the right content area. */
  width: min(70%, 1120px);
  margin: 0 auto;
}

.chat-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding: 12px 14px;
  border: 1px solid rgba(255, 255, 255, 0.10);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.03);
}

.chat-title {
  font-weight: 750;
  letter-spacing: 0.01em;
}

.chat-sub {
  font-size: 12px;
  opacity: 0.7;
}

.chat-list {
  flex: 1;
  overflow: auto;
  padding: 14px;
  border: 1px solid rgba(255, 255, 255, 0.10);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.03);
}

.msg {
  display: flex;
  margin: 10px 0;
}

.msg-user {
  justify-content: flex-end;
}

.msg-assistant {
  justify-content: flex-start;
}

.bubble {
  max-width: 100%;
  padding: 10px 12px;
  border-radius: 14px;
  line-height: 1.45;
  word-break: break-word;
  border: 1px solid rgba(255, 255, 255, 0.10);
}

.msg-user .bubble {
  max-width: min(680px, 82%);
}

.msg-assistant .bubble {
  max-width: 100%;
}

.plain {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font: inherit;
}

.tail {
  margin: 0;
  opacity: 0.92;
}

.msg-user .bubble {
  background: rgba(77, 215, 200, 0.14);
  border-color: rgba(77, 215, 200, 0.28);
}

.msg-assistant .bubble {
  background: rgba(255, 255, 255, 0.04);
}

@media (max-width: 980px) {
  .chat {
    width: 100%;
  }
}

.chat-input {
  display: flex;
  gap: 10px;
  padding: 12px;
  border: 1px solid rgba(255, 255, 255, 0.10);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.03);
}

.input {
  flex: 1;
  resize: none;
  border: 1px solid rgba(255, 255, 255, 0.10);
  background: rgba(0, 0, 0, 0.25);
  color: rgba(255, 255, 255, 0.9);
  border-radius: 12px;
  padding: 10px 12px;
  outline: none;
  min-height: 40px;
  max-height: 160px;
}

.input:disabled {
  opacity: 0.7;
}

.send {
  width: 44px;
  height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  border: 1px solid rgba(77, 215, 200, 0.40);
  background: rgba(77, 215, 200, 0.18);
  color: rgba(255, 255, 255, 0.92);
  cursor: pointer;
}

.send.stopping {
  border-color: rgba(255, 255, 255, 0.28);
  background: rgba(255, 255, 255, 0.10);
}

.send:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

.send-icon {
  width: 20px;
  height: 20px;
}

/* Markdown styling (scoped -> use :deep) */
.markdown :deep(p) {
  margin: 0;
}

.markdown :deep(p + p) {
  margin-top: 10px;
}

.markdown :deep(a) {
  color: rgba(77, 215, 200, 0.95);
  text-decoration: underline;
  word-break: break-all;
}

.markdown :deep(code) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 0.92em;
  background: rgba(0, 0, 0, 0.28);
  border: 1px solid rgba(255, 255, 255, 0.10);
  padding: 2px 6px;
  border-radius: 8px;
}

.markdown :deep(pre) {
  margin: 10px 0 0;
  padding: 12px;
  overflow: auto;
  border-radius: 12px;
  background: rgba(0, 0, 0, 0.35);
  border: 1px solid rgba(255, 255, 255, 0.10);
}

.markdown :deep(pre code) {
  display: block;
  padding: 0;
  border: 0;
  background: transparent;
  white-space: pre;
}

.markdown :deep(ul),
.markdown :deep(ol) {
  margin: 10px 0 0;
  padding-left: 22px;
}

.markdown :deep(li) {
  margin: 6px 0;
}
</style>
