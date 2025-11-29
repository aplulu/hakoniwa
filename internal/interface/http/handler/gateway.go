package handler

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aplulu/hakoniwa/internal/domain/model"
	"github.com/aplulu/hakoniwa/internal/interface/http/middleware"
	"github.com/aplulu/hakoniwa/internal/usecase"
)

type GatewayHandler struct {
	authUsecase     usecase.Auth
	instanceUsecase usecase.InstanceManagement
	apiServer       http.Handler
	proxyHandler    *ProxyHandler
	staticHandler   http.Handler
	logger          *slog.Logger
}

func NewGatewayHandler(
	authUsecase usecase.Auth,
	instanceUsecase usecase.InstanceManagement,
	apiServer http.Handler,
	proxyHandler *ProxyHandler,
	staticDir string,
	logger *slog.Logger,
) *GatewayHandler {
	return &GatewayHandler{
		authUsecase:     authUsecase,
		instanceUsecase: instanceUsecase,
		apiServer:       apiServer,
		proxyHandler:    proxyHandler,
		staticHandler:   http.FileServer(http.Dir(staticDir)),
		logger:          logger,
	}
}

func (h *GatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 1. API Handling
	if strings.HasPrefix(path, "/_hakoniwa/api") {
		// Wrap with CookieSetter for LoginAnonymous
		ctx := WithCookieSetter(r.Context(), func(token string) {
			http.SetCookie(w, &http.Cookie{
				Name:     "hakoniwa_session",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Expires:  time.Now().Add(24 * time.Hour), // TODO: Configurable
			})
		})

		http.StripPrefix("/_hakoniwa/api", h.apiServer).ServeHTTP(w, r.WithContext(ctx))
		return
	}

	// 2. Static Assets Handling (always allow access to assets)
	if strings.HasPrefix(path, "/_hakoniwa/") {
		// Serve static files (assets)
		// Strip prefix so that /_hakoniwa/assets/foo.js -> /assets/foo.js in file server
		// Note: Vite base is /_hakoniwa/, so requests will come as /_hakoniwa/assets/...
		// If we mount ui/dist at /, then we need to strip /_hakoniwa/
		http.StripPrefix("/_hakoniwa/", h.staticHandler).ServeHTTP(w, r)
		return
	}

	// 3. Check Authentication
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		// Not authenticated -> Serve Login Screen (React index.html)
		// React router should handle /login route if exists, but here we serve the main entry point
		h.serveReactApp(w, r)
		return
	}

	// 4. Pod Routing Logic
	// Authenticated -> Check Instance Status
	instance, err := h.instanceUsecase.GetInstanceStatus(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("Failed to get instance status", "user_id", user.ID, "error", err)
		h.serveReactApp(w, r) // Show error/loading in React
		return
	}

	if instance != nil && instance.Status == model.InstanceStatusRunning && instance.PodIP != "" {
		// Proxy to Pod
		h.proxyHandler.Proxy(user.ID, instance.PodIP, w, r)
		return
	}

	// Instance not ready -> Serve Loading Screen (React index.html)
	h.serveReactApp(w, r)
}

func (h *GatewayHandler) serveReactApp(w http.ResponseWriter, r *http.Request) {
	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Serve index.html for SPA
	// We can reuse staticHandler but we need to make sure it serves index.html for unknown paths?
	// Or explicitly serve index.html
	// Since staticHandler is a FileServer, it serves index.html if path is /
	// But here we might be at any path.
	// Ideally, we read index.html content and serve it.
	// For simplicity, let's assume staticHandler points to ui/dist.
	// We need to serve ui/dist/index.html.

	// Re-constructing request to serve /index.html from static handler
	r.URL.Path = "/"
	h.staticHandler.ServeHTTP(w, r)
}
