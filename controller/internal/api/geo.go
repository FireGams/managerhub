package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

// GeoInfo holds approximate geolocation of an IP.
type GeoInfo struct {
	City    string
	Country string
}

// lookupGeo resolves an IP to city/country using ip-api.com (from MeshService).
func lookupGeo(ctx context.Context, ip string) GeoInfo {
	if ip == "" || ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return GeoInfo{City: "Local", Country: "Local"}
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET",
		"http://ip-api.com/json/"+ip+"?fields=status,country,city", nil)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return GeoInfo{City: "Unknown", Country: "Unknown"}
	}
	defer func() { _ = resp.Body.Close() }()
	var d struct {
		Status  string `json:"status"`
		Country string `json:"country"`
		City    string `json:"city"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil || d.Status != "success" {
		return GeoInfo{City: "Unknown", Country: "Unknown"}
	}
	return GeoInfo{City: d.City, Country: d.Country}
}

// readClientIP extracts the real client IP.
func readClientIP(r *http.Request) string {
	if f := r.Header.Get("X-Forwarded-For"); f != "" {
		return strings.TrimSpace(strings.Split(f, ",")[0])
	}
	h, _, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		return r.RemoteAddr
	}
	return h
}
