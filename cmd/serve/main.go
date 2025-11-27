package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aplulu/hakoniwa/internal/config"
	"github.com/aplulu/hakoniwa/internal/server"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := config.LoadConf(); err != nil {
		log.Error(fmt.Sprintf("failed to load config: %v", err))
		os.Exit(1)
	}

	quitCh := make(chan os.Signal, 1)
	signal.Notify(
		quitCh,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		<-quitCh
		log.Info("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.StopServer(shutdownCtx); err != nil {
			log.Error(fmt.Sprintf("failed to stop server: %v", err))
			os.Exit(1)
		}
	}()

	staticDir := "ui/dist"
	if err := server.StartServer(log, staticDir); err != nil {
		log.Error(fmt.Sprintf("failed to start server: %v", err))
		os.Exit(1)
	}
}