package bazi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ResolveCityLongitude resolves a city's longitude (degrees East) based on the provided
// human-readable province/city names.
//
// Implementation strategy:
//   - Use AMap (高德) geocoding only. If it fails, the caller should fall back to
//     Beijing time (no longitude correction) and surface the AMap error.
func ResolveCityLongitude(ctx context.Context, provinceName, cityName string) (float64, error) {
	provinceName = normalizeAdminName(provinceName)
	origCityName := strings.TrimSpace(cityName)
	cityName = normalizeAdminName(cityName)
	if cityName == "" && origCityName != "" {
		cityName = origCityName
	}
	if strings.TrimSpace(cityName) == "" {
		return 0, fmt.Errorf("city name is empty")
	}

	cacheKey := provinceName + "|" + cityName
	if v, ok := geoCache.Load(cacheKey); ok {
		return v.(float64), nil
	}

	key := strings.TrimSpace(os.Getenv("AMAP_KEY"))
	// Be tolerant of quotes if users set AMAP_KEY in an env file like AMAP_KEY="xxx".
	key = strings.Trim(key, "\"'")
	if key == "" {
		return 0, fmt.Errorf("amap_key_missing")
	}
	lon, err := resolveCityLongitudeAMap(ctx, key, provinceName, cityName)
	if err != nil {
		return 0, fmt.Errorf("amap_api_failed(%s): %w", keyHint(key), err)
	}
	geoCache.Store(cacheKey, lon)
	return lon, nil
}

var geoCache sync.Map // key -> float64 longitude

func keyHint(key string) string {
	// Do not leak the full key. Provide a small fingerprint for debugging.
	key = strings.TrimSpace(key)
	n := len(key)
	last4 := key
	if n > 4 {
		last4 = key[n-4:]
	}
	return fmt.Sprintf("key_len=%d,last4=%s", n, last4)
}

// ----------------------------
// AMap geocoding (CN, requires key)
// ----------------------------

type amapGeoResp struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Geocodes []struct {
		Location string `json:"location"` // "lng,lat"
	} `json:"geocodes"`
}

func resolveCityLongitudeAMap(ctx context.Context, key, provinceName, cityName string) (float64, error) {
	base := "https://restapi.amap.com/v3/geocode/geo"
	q := url.Values{}
	q.Set("key", key)
	q.Set("address", cityName)
	if provinceName != "" {
		q.Set("city", provinceName)
	}
	u := base + "?" + q.Encode()

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          20,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 6 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return 0, fmt.Errorf("amap http %d", resp.StatusCode)
	}

	var out amapGeoResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, err
	}
	if out.Status != "1" {
		return 0, fmt.Errorf("amap error: %s (%s)", out.Info, out.Infocode)
	}
	if len(out.Geocodes) == 0 || strings.TrimSpace(out.Geocodes[0].Location) == "" {
		return 0, fmt.Errorf("amap empty")
	}
	parts := strings.Split(strings.TrimSpace(out.Geocodes[0].Location), ",")
	if len(parts) < 1 {
		return 0, fmt.Errorf("amap invalid location")
	}
	lon, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, fmt.Errorf("amap invalid longitude")
	}
	return lon, nil
}

func normalizeAdminName(s string) string {
	s = strings.TrimSpace(s)
	// Remove common Chinese administrative suffixes (longer ones first).
	s = strings.TrimSuffix(s, "特别行政区")
	s = strings.TrimSuffix(s, "维吾尔自治区")
	s = strings.TrimSuffix(s, "壮族自治区")
	s = strings.TrimSuffix(s, "回族自治区")
	s = strings.TrimSuffix(s, "自治区")
	s = strings.TrimSuffix(s, "自治州")
	s = strings.TrimSuffix(s, "地区")
	s = strings.TrimSuffix(s, "盟")
	s = strings.TrimSuffix(s, "州")
	s = strings.TrimSuffix(s, "省")
	s = strings.TrimSuffix(s, "市")
	return strings.TrimSpace(s)
}
