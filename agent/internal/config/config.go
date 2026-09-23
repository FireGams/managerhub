// Package config loads agent configuration from the environment.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds agent settings.
type Config struct {
	ControllerURL     string
	EnrollToken       string
	Name              string
	StatePath         string
	HeartbeatInterval time.Duration
	MetricsInterval   time.Duration
	Shell             string
}

// Load reads MH_* environment variables.
func Load() (Config, error) {
	cfg := Config{
		ControllerURL:     getenv("MH_CONTROLLER_URL", "http://localhost:8080"),
		EnrollToken:       os.Getenv("MH_ENROLL_TOKEN"),
		Name:              os.Getenv("MH_AGENT_NAME"),
		StatePath:         getenv("MH_AGENT_STATE", "agent.state.json"),
		HeartbeatInterval: duration("MH_HEARTBEAT_INTERVAL", 10*time.Second),
		MetricsInterval:   duration("MH_METRICS_INTERVAL", 15*time.Second),
		Shell:             os.Getenv("MH_AGENT_SHELL"),
	}
	if cfg.ControllerURL == "" {
		return cfg, fmt.Errorf("config: MH_CONTROLLER_URL is required")
	}
	if cfg.Name == "" {
		h, _ := os.Hostname()
		cfg.Name = h
	}
	return cfg, nil
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func duration(k string, d time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return d
}
