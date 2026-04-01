package main

// @title           Modami Core Service API
// @version         1.0
// @description     Core service: catalog, orders, master data, and seller APIs for the Modami marketplace.
// @host            localhost:8087
// @BasePath        /v1/core-services
// @schemes         http https
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modami/core-service/config"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	serviceName := cfg.Observability.ServiceName
	if serviceName == "" {
		serviceName = cfg.App.Name
	}
	serviceVersion := cfg.Observability.ServiceVersion
	if serviceVersion == "" {
		serviceVersion = cfg.App.Version
	}
	environment := cfg.Observability.Environment
	if environment == "" {
		environment = cfg.App.Environment
	}

	if err := logger.Init(logging.Config{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		Environment:    environment,
		Level:          cfg.Observability.LogLevel,
		OTLPEndpoint:   cfg.Observability.OTLPEndpoint,
		Insecure:       cfg.Observability.OTLPInsecure,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	defer logger.Shutdown(ctx)

	conns, err := newConnections(ctx, cfg)
	if err != nil {
		logger.Error(ctx, "failed to establish connections", err)
		os.Exit(1)
	}
	defer conns.Disconnect(ctx)

	app, err := newApplication(ctx, cfg, conns)
	if err != nil {
		logger.Error(ctx, "failed to build application", err)
		os.Exit(1)
	}

	go func() {
		logger.Info(ctx, "HTTP server listening", logging.String("addr", cfg.App.ListenAddr()), logging.String("port", cfg.App.PortString()))
		if err := app.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "http server", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "shutting down...")

	shutdownTimeout := cfg.App.ShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = 30 * time.Second
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := app.HTTPServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "HTTP shutdown error", err)
	}
	logger.Info(ctx, "server exited")
}
