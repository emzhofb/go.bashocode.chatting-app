package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env                string
	PublicAddr         string
	InternalAddr       string
	DatabaseURL        string
	PublicBaseURL      string
	AllowedOrigins     string
	ReadinessTimeout   time.Duration
	ShutdownTimeout    time.Duration
	LogLevel           string
	MigrationLockTable string
	BcryptCost         int
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
}

func Load() (Config, error) {
	publicAddr, err := required("APP_HTTP_ADDR")
	if err != nil {
		return Config{}, err
	}
	internalAddr, err := required("APP_INTERNAL_HTTP_ADDR")
	if err != nil {
		return Config{}, err
	}
	databaseURL, err := required("APP_DATABASE_URL")
	if err != nil {
		return Config{}, err
	}

	readinessTimeout, err := durationEnv("APP_READINESS_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := durationEnv("APP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	bcryptCost, err := intEnv("APP_BCRYPT_COST", 12)
	if err != nil || bcryptCost < 10 || bcryptCost > 31 {
		return Config{}, fmt.Errorf("APP_BCRYPT_COST must be an integer between 10 and 31")
	}
	accessTTL, err := durationEnv("APP_ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	refreshTTL, err := durationEnv("APP_REFRESH_TOKEN_TTL", 30*24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Env:                envOr("APP_ENV", "development"),
		PublicAddr:         publicAddr,
		InternalAddr:       internalAddr,
		DatabaseURL:        databaseURL,
		PublicBaseURL:      envOr("APP_PUBLIC_BASE_URL", "http://localhost:8080"),
		AllowedOrigins:     envOr("APP_ALLOWED_ORIGINS", "http://localhost:3000"),
		ReadinessTimeout:   readinessTimeout,
		ShutdownTimeout:    shutdownTimeout,
		LogLevel:           envOr("APP_LOG_LEVEL", "info"),
		MigrationLockTable: envOr("APP_MIGRATION_LOCK_TABLE", "schema_migrations"),
		BcryptCost:         bcryptCost,
		AccessTokenTTL:     accessTTL,
		RefreshTokenTTL:    refreshTTL,
	}, nil
}

func intEnv(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return parsed, nil
}

func required(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func durationEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", name, err)
	}
	if parsed <= 0 {
		return 0, errors.New(name + " must be greater than zero")
	}
	return parsed, nil
}

func BoolEnv(name string, fallback bool) bool {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
