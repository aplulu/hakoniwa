package handler

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/aplulu/hakoniwa/internal/usecase"
)

type ProxyHandler struct {
	instanceUsecase usecase.InstanceManagement
	logger          *slog.Logger
}

func NewProxyHandler(instanceUsecase usecase.InstanceManagement, logger *slog.Logger) *ProxyHandler {
	return &ProxyHandler{
		instanceUsecase: instanceUsecase,
		logger:          logger,
	}
}

func (h *ProxyHandler) Proxy(instanceID, targetURL string, w http.ResponseWriter, r *http.Request) {
	url, err := url.Parse(targetURL)
	if err != nil {
		h.logger.Error("Failed to parse proxy target URL", "url", targetURL, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Activity Tracker Setup
	var lastUpdate time.Time
	var mu sync.Mutex

	tracker := func() {
		mu.Lock()
		now := time.Now()
		if now.Sub(lastUpdate) < 10*time.Second {
			mu.Unlock()
			return
		}
		lastUpdate = now
		mu.Unlock()

		// Update LastActiveAt asynchronously
		go func() {
			if err := h.instanceUsecase.UpdateLastActive(context.Background(), instanceID); err != nil {
				h.logger.Error("Failed to update instance activity", "instance_id", instanceID, "error", err)
			}
		}()
	}

	// Wrap ResponseWriter
	wrappedWriter := &ActivityTrackingResponseWriter{
		ResponseWriter: w,
		tracker:        tracker,
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

	proxy.ServeHTTP(wrappedWriter, r)
}

// --- Activity Tracking Wrappers ---

type ActivityTracker func()

type ActivityTrackingResponseWriter struct {
	http.ResponseWriter
	tracker ActivityTracker
}

func (w *ActivityTrackingResponseWriter) Write(b []byte) (int, error) {
	w.tracker()
	return w.ResponseWriter.Write(b)
}

func (w *ActivityTrackingResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *ActivityTrackingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	conn, rw, err := h.Hijack()
	if err != nil {
		return nil, nil, err
	}

	trackedConn := &ActivityTrackingConn{
		Conn:    conn,
		tracker: w.tracker,
	}

	return trackedConn, rw, nil
}

type ActivityTrackingConn struct {
	net.Conn
	tracker ActivityTracker
}

func (c *ActivityTrackingConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 {
		c.tracker()
	}
	return
}

func (c *ActivityTrackingConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if n > 0 {
		c.tracker()
	}
	return
}
