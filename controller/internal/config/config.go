// Package config loads controller configuration from the environment.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds every tunable of the controller process.
type Config struct {
	HTTPAddr         string
	PublicURL        string
	DatabaseURL      string
	JWTSecret        string
	EnrollTokenTTL   time.Duration
	MetricsRetention time.Duration
	JobLogRetention  time.Duration
	LogLevel         string
}

// Load reads configuration from MH_* environment variables.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:         getenv("MH_HTTP_ADDR", ":8080"),
		PublicURL:        getenv("MH_PUBLIC_URL", "http://localhost:8080"),
		DatabaseURL:      os.Getenv("MH_DATABASE_URL"),
		JWTSecret:        os.Getenv("MH_JWT_SECRET"),
		EnrollTokenTTL:   duration("MH_ENROLL_TOKEN_TTL", 24*time.Hour),
		MetricsRetention: duration("MH_METRICS_RETENTION", 168*time.Hour),
		JobLogRetention:  duration("MH_JOBLOG_RETENTION", 720*time.Hour),
		LogLevel:         getenv("MH_LOG_LEVEL", "info"),
	}
	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("config: MH_DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 16 {
		return cfg, fmt.Errorf("config: MH_JWT_SECRET must be at least 16 characters")
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func duration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
