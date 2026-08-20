package config

import "testing"

func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("APP_HTTP_ADDR", ":8080")
	t.Setenv("APP_DATABASE_URL", "postgresql://app_chat:password@localhost:5432/app_chat?sslmode=disable")
	t.Setenv("OPENIM_API_URL", "http://openim:10002")
	t.Setenv("OPENIM_ADMIN_SECRET", "development-only-secret")
}

func TestLoadFailsFastWhenRequiredValueMissing(t *testing.T) {
	setRequired(t)
	t.Setenv("APP_DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing APP_DATABASE_URL error")
	}
}

func TestLoadUsesSafeDefaults(t *testing.T) {
	setRequired(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Env != "development" || cfg.BcryptCost != 12 || cfg.OpenIMPlatformID <= 0 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}
