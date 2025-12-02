package handler

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aplulu/hakoniwa/internal/config"
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
		// Wrap with CookieSetter/Clearer
		ctx := WithCookieSetter(r.Context(), func(token string) {
			http.SetCookie(w, &http.Cookie{
				Name:     "hakoniwa_session",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Now().Add(config.SessionExpiration()),
			})
		})
		ctx = WithCookieClearer(ctx, func() {
			// Clear Session
			http.SetCookie(w, &http.Cookie{
				Name:     "hakoniwa_session",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Unix(0, 0),
			})
			// Clear Instance Cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "hakoniwa_instance_id",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
			})
		})

		http.StripPrefix("/_hakoniwa/api", h.apiServer).ServeHTTP(w, r.WithContext(ctx))
		return
	}

	// 2. Static Assets Handling (always allow access to assets)
	if strings.HasPrefix(path, "/_hakoniwa/") {
		// Serve static files (assets)
		http.StripPrefix("/_hakoniwa/", h.staticHandler).ServeHTTP(w, r)
		return
	}

	// 3. Check Authentication
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.redirectToDashboard(w, r)
		return
	}

	// 4. Instance Routing Logic (Cookie-based)
	
	// Explicit Dashboard Access -> Clear Instance Cookie
	if path == "/_hakoniwa/dashboard" {
		http.SetCookie(w, &http.Cookie{
			Name:     "hakoniwa_instance_id",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			HttpOnly: false, // Allow JS to read if needed? Or keep secure. JS needs to Set it, so not HttpOnly? 
			// Actually, if we set it via JS in frontend, this backend handler might overwrite it?
			// Let's assume Frontend sets it to "activate". Backend clears it to "deactivate".
		})
		h.redirectToDashboard(w, r)
		return
	}

	// Check for Active Instance Cookie
	cookie, err := r.Cookie("hakoniwa_instance_id")
	if err == nil && cookie.Value != "" {
		instanceID := cookie.Value
		instance, err := h.instanceUsecase.GetInstance(r.Context(), instanceID)
		if err != nil {
			h.logger.Error("Failed to get instance from cookie", "id", instanceID, "error", err)
			// Fallthrough to dashboard
		} else if instance != nil && instance.UserID == user.ID {
			if instance.Status == model.InstanceStatusRunning && instance.PodIP != "" {
				// Get Port
				it, ok := config.GetInstanceType(instance.Type)
				port := "3000"
				if ok && it.TargetPort != "" {
					port = it.TargetPort
				}
				// Proxy to Pod
				targetURL := "http://" + instance.PodIP + ":" + port
				h.proxyHandler.Proxy(instance.InstanceID, targetURL, w, r)
				return
			}
		}
	}

	// 5. Dashboard (Default)
	h.redirectToDashboard(w, r)
}

func (h *GatewayHandler) redirectToDashboard(w http.ResponseWriter, r *http.Request) {
	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.Redirect(w, r, "/_hakoniwa/", http.StatusFound)
}
