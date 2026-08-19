package config

import "testing"

func TestLoadFailsFastWhenRequiredValueMissing(t *testing.T) {
	t.Setenv("APP_HTTP_ADDR", ":8080")
	t.Setenv("APP_INTERNAL_HTTP_ADDR", ":8081")
	t.Setenv("APP_DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing APP_DATABASE_URL error")
	}
}

func TestLoadUsesSafeDefaults(t *testing.T) {
	t.Setenv("APP_HTTP_ADDR", ":8080")
	t.Setenv("APP_INTERNAL_HTTP_ADDR", ":8081")
	t.Setenv("APP_DATABASE_URL", "postgresql://app_chat:password@localhost:5432/app_chat?sslmode=disable")
	t.Setenv("APP_ENV", "")
	t.Setenv("APP_READINESS_TIMEOUT", "")
	t.Setenv("APP_SHUTDOWN_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Env != "development" || cfg.ReadinessTimeout <= 0 || cfg.ShutdownTimeout <= 0 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}
