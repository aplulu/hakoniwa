package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ProxyHandler struct {
	logger *slog.Logger
}

func NewProxyHandler(logger *slog.Logger) *ProxyHandler {
	return &ProxyHandler{
		logger: logger,
	}
}

func (h *ProxyHandler) Proxy(targetIP string, w http.ResponseWriter, r *http.Request) {
	targetURL := fmt.Sprintf("http://%s:3000", targetIP)
	url, err := url.Parse(targetURL)
	if err != nil {
		h.logger.Error("Failed to parse proxy target URL", "url", targetURL, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Webtop/KasmVNC requires Upgrade header for WebSocket
		if r.Header.Get("Upgrade") != "" {
			req.Header.Set("Upgrade", r.Header.Get("Upgrade"))
			req.Header.Set("Connection", r.Header.Get("Connection"))
		}
		// Ensure Host header matches target or is handled correctly
		req.Host = url.Host
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		h.logger.Error("Proxy error", "error", err, "target", targetURL)
		// Gateway logic should handle 502/reloads, but here we just return error
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
}
