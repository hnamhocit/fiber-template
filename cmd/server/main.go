package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/hnamhocit/fiber-template/internal/bootstrap"
	"github.com/hnamhocit/fiber-template/internal/config"
	"github.com/hnamhocit/fiber-template/internal/logger"
	"github.com/hnamhocit/fiber-template/internal/server"
)

// @title			Fiber Template API
// @version		1.0
// @description	Production-ready Go Fiber template: ftpl migrations, sqlc, live API docs.
// @host			localhost:8080
// @BasePath		/v1
func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. Config + logger.
	cfg := config.Load()
	slog.SetDefault(logger.New(cfg.LogLevel, cfg.IsProduction()))

	// 2. Bootstrap infra (DB + optional Redis/Minio).
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	infra, err := bootstrap.Init(ctx, cfg, config.LoadServices())
	if err != nil {
		return err
	}
	defer infra.Close()

	// 3. Build Deps container + app.
	deps := server.Deps{Cfg: cfg, DB: infra.DB, Redis: infra.Redis, Minio: infra.Minio}
	hc := bootstrap.Health(infra)
	app := server.New(deps, hc, features(deps)...) // ← ĐÚNG: truyền deps, không phải cfg

	// 4. Serve.
	return serve(app, cfg)
}

func serve(app *fiber.App, cfg *config.Config) error {
	errCh := make(chan error, 1)
	go func() {
		slog.Info("server listening", "port", cfg.Port, "env", cfg.AppEnv)
		errCh <- app.Listen(cfg.Port, fiber.ListenConfig{
			DisableStartupMessage: true,
		})
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		slog.Info("shutting down", "signal", sig.String())
	case err := <-errCh:
		if err != nil {
			return err
		}
		slog.Info("server stopped unexpectedly")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		return err
	}

	slog.Info("server stopped")
	return nil
}
