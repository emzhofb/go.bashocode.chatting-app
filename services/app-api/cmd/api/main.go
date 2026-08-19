package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ikhda/tinode-chat/services/app-api/internal/auth"
	"github.com/ikhda/tinode-chat/services/app-api/internal/config"
	"github.com/ikhda/tinode-chat/services/app-api/internal/health"
	"github.com/ikhda/tinode-chat/services/app-api/internal/httpapi"
	"github.com/ikhda/tinode-chat/services/app-api/internal/storage"
	"github.com/ikhda/tinode-chat/services/app-api/internal/tinodeauth"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	startupCtx, startupCancel := context.WithTimeout(context.Background(), cfg.ReadinessTimeout)
	defer startupCancel()
	if err := db.PingContext(startupCtx); err != nil {
		logger.Error("database is not reachable", "error", err)
		os.Exit(1)
	}
	if err := (storage.MigrationRunner{DB: db, LockTable: cfg.MigrationLockTable}).Run(startupCtx); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	healthHandler := health.Handler{DB: db, Timeout: cfg.ReadinessTimeout}
	authHandler := httpapi.AuthHandler{Service: auth.Service{DB: db, BcryptCost: cfg.BcryptCost, AccessTokenTTL: cfg.AccessTokenTTL, RefreshTokenTTL: cfg.RefreshTokenTTL}, Limiter: httpapi.NewRateLimiter(20, time.Minute)}
	tinodeAuthHandler := tinodeauth.Handler{Service: authHandler.Service}
	publicMux := http.NewServeMux()
	publicMux.HandleFunc("GET /health/live", healthHandler.Live)
	publicMux.HandleFunc("GET /health/ready", healthHandler.Ready)
	publicMux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	publicMux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	publicMux.HandleFunc("POST /v1/auth/refresh", authHandler.Refresh)
	publicMux.Handle("POST /v1/auth/logout", authHandler.RequireAuth(http.HandlerFunc(authHandler.Logout)))
	publicMux.Handle("GET /v1/me", authHandler.RequireAuth(http.HandlerFunc(authHandler.Me)))
	internalMux := http.NewServeMux()
	internalMux.HandleFunc("GET /health/live", healthHandler.Live)
	internalMux.HandleFunc("GET /health/ready", healthHandler.Ready)
	internalMux.Handle("POST /internal/tinode/auth/{endpoint}", tinodeAuthHandler)

	publicServer := &http.Server{Addr: cfg.PublicAddr, Handler: publicMux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	internalServer := &http.Server{Addr: cfg.InternalAddr, Handler: internalMux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}

	serverErrors := make(chan error, 2)
	go func() { serverErrors <- publicServer.ListenAndServe() }()
	go func() { serverErrors <- internalServer.ListenAndServe() }()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	case <-shutdownCtx.Done():
		logger.Info("shutdown requested")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := publicServer.Shutdown(ctx); err != nil {
		logger.Error("shutdown public server", "error", err)
	}
	if err := internalServer.Shutdown(ctx); err != nil {
		logger.Error("shutdown internal server", "error", err)
	}
	logger.Info("app-api stopped")
}
