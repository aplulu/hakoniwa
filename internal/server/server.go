package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aplulu/hakoniwa/internal/api/hakoniwa"
	"github.com/aplulu/hakoniwa/internal/config"
	"github.com/aplulu/hakoniwa/internal/infrastructure/kubernetes"
	"github.com/aplulu/hakoniwa/internal/infrastructure/memory"
	"github.com/aplulu/hakoniwa/internal/interface/background"
	"github.com/aplulu/hakoniwa/internal/interface/http/handler"
	"github.com/aplulu/hakoniwa/internal/interface/http/middleware"
	"github.com/aplulu/hakoniwa/internal/usecase"
)

var (
	server        *http.Server
	cleanerCancel context.CancelFunc
)

func StartServer(log *slog.Logger, staticDir string) error {
	// Infrastructure
	instanceRepository := memory.NewInstanceRepository()
	k8sClient, err := kubernetes.NewClient(log)
	if err != nil {
		log.Error("failed to create k8s client", "error", err)
		return fmt.Errorf("server.StartServer: failed to create k8s client: %w", err)
	}

	// Background Workers
	ctx, cancel := context.WithCancel(context.Background())
	cleanerCancel = cancel

	cleaner := background.NewInactivityCleaner(
		instanceRepository,
		k8sClient,
		log,
		config.InstanceInactivityTimeout(),
	)
	go cleaner.Start(ctx)

	syncer := background.NewInstanceSyncer(
		instanceRepository,
		k8sClient,
		log,
	)
	go syncer.Start(ctx, 5*time.Second)

	// Usecase
	authUsecase, err := usecase.NewAuthInteractor()
	if err != nil {
		// Log error but continue? Or fail?
		// If OIDC is enabled but fails, we should probably fail or warn.
		// Prompt says "settings are loaded from config".
		// If OIDC is misconfigured, it's better to know early.
		// However, if user wants to run without OIDC but config is messy?
		// Let's log error and if OIDC was strictly required (enabled=true), maybe exit?
		// But NewAuthInteractor only returns error if OIDCEnabled and NewProvider fails.
		// So we should treat it as fatal if we intend to support OIDC.
		if config.OIDCEnabled() {
			return fmt.Errorf("server.StartServer: failed to initialize auth usecase: %w", err)
		}
		// If not enabled, maybe it returns nil error? Yes.
		// But NewAuthInteractor only returns error if OIDCEnabled is true.
		// So safe to return error.
		return fmt.Errorf("server.StartServer: failed to initialize auth usecase: %w", err)
	}
	instanceUsecase := usecase.NewInstanceInteractor(instanceRepository, k8sClient)

	// Handlers
	apiHandler := handler.NewAPIHandler(authUsecase, instanceUsecase)
	apiServer, err := hakoniwa.NewServer(apiHandler)
	if err != nil {
		return fmt.Errorf("server.StartServer: failed to create api server: %w", err)
	}

	proxyHandler := handler.NewProxyHandler(instanceUsecase, log)

	gatewayHandler := handler.NewGatewayHandler(
		authUsecase,
		instanceUsecase,
		apiServer,
		proxyHandler,
		staticDir,
		log,
	)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(authUsecase)
	h := authMiddleware.Handle(gatewayHandler)

	server = &http.Server{
		Addr: net.JoinHostPort(config.Listen(), config.Port()),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
				return
			}

			if strings.HasPrefix(r.URL.Path, "/_hakoniwa/docs/") {
				http.StripPrefix("/_hakoniwa/docs", newDocsHandler()).ServeHTTP(w, r)
				return
			}

			h.ServeHTTP(w, r)
		}),
	}

	listenHost := config.Listen()
	if listenHost == "" {
		listenHost = "localhost"
	}

	log.Info(fmt.Sprintf("Server started at http://%s:%s", listenHost, config.Port()))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func StopServer(ctx context.Context) error {
	if cleanerCancel != nil {
		cleanerCancel()
	}
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server.StopServer: failed to stop server: %w", err)
	}
	return nil
}
