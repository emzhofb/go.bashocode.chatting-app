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

	"github.com/ikhda/openim-chat/services/app-api/internal/auth"
	"github.com/ikhda/openim-chat/services/app-api/internal/config"
	"github.com/ikhda/openim-chat/services/app-api/internal/health"
	"github.com/ikhda/openim-chat/services/app-api/internal/httpapi"
	"github.com/ikhda/openim-chat/services/app-api/internal/openim"
	"github.com/ikhda/openim-chat/services/app-api/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel(cfg.LogLevel)})))

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)

	startupCtx, startupCancel := context.WithTimeout(context.Background(), cfg.ReadinessTimeout)
	defer startupCancel()
	if err := db.PingContext(startupCtx); err != nil {
		slog.Error("database is not reachable", "error", err)
		os.Exit(1)
	}
	if err := (storage.MigrationRunner{DB: db, LockTable: cfg.MigrationLockTable}).Run(startupCtx); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	openIMClient := &openim.Client{BaseURL: cfg.OpenIMAPIURL, AdminUser: cfg.OpenIMAdminUserID, AdminSecret: cfg.OpenIMAdminSecret, HTTP: &http.Client{Timeout: cfg.OpenIMHTTPTimeout}}
	authService := auth.Service{DB: db, Provisioner: openIMClient, BcryptCost: cfg.BcryptCost, AccessTokenTTL: cfg.AccessTokenTTL, RefreshTokenTTL: cfg.RefreshTokenTTL}
	authHandler := httpapi.AuthHandler{Service: authService, OpenIM: openIMClient, PublicAPIURL: cfg.OpenIMPublicAPIURL, PublicWSURL: cfg.OpenIMPublicWSURL, PlatformID: cfg.OpenIMPlatformID, Limiter: httpapi.NewRateLimiter(20, time.Minute)}
	healthHandler := health.Handler{DB: db, Timeout: cfg.ReadinessTimeout}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", healthHandler.Live)
	mux.HandleFunc("GET /health/ready", healthHandler.Ready)
	mux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /v1/auth/refresh", authHandler.Refresh)
	mux.Handle("POST /v1/auth/logout", authHandler.RequireAuth(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /v1/me", authHandler.RequireAuth(http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /v1/openim/session", authHandler.RequireAuth(http.HandlerFunc(authHandler.OpenIMSession)))
	mux.Handle("GET /v1/users/search", authHandler.RequireAuth(http.HandlerFunc(authHandler.SearchUsers)))
	server := &http.Server{Addr: cfg.PublicAddr, Handler: httpapi.CORSHandler(mux, cfg.AllowedOrigins), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}

	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	case <-shutdownCtx.Done():
		slog.Info("shutdown requested")
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("shutdown http server", "error", err)
	}
	slog.Info("app-api stopped")
}

func logLevel(value string) slog.Level {
	switch value {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
