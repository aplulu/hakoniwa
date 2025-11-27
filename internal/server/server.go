package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"

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
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	// Background Workers
	cleaner := background.NewInactivityCleaner(
		instanceRepository,
		k8sClient,
		log,
		config.InstanceInactivityTimeout(),
	)
	ctx, cancel := context.WithCancel(context.Background())
	cleanerCancel = cancel
	go cleaner.Start(ctx)

	// Usecase
	authUsecase := usecase.NewAuthInteractor()
	instanceUsecase := usecase.NewInstanceInteractor(instanceRepository, k8sClient)

	// Handlers
	apiHandler := handler.NewAPIHandler(authUsecase, instanceUsecase)
	apiServer, err := hakoniwa.NewServer(apiHandler)
	if err != nil {
		return fmt.Errorf("failed to create api server: %w", err)
	}

	proxyHandler := handler.NewProxyHandler(log)

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
	handler := authMiddleware.Handle(gatewayHandler)

	server = &http.Server{
		Addr: net.JoinHostPort(config.Listen(), config.Port()),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/_/docs/") {
				http.StripPrefix("/_/docs", newDocsHandler()).ServeHTTP(w, r)
				return
			}

			handler.ServeHTTP(w, r)
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
