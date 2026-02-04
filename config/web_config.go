// Package config provides startup configuration for the web service.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// WebConfig defines startup settings for the web server.
type WebConfig struct {
	Host                 string `json:"host"`
	Port                 int    `json:"port"`
	ResendBatchSize      int    `json:"resend_batch_size"`
	ResendBatchWorkers   int    `json:"resend_batch_workers"`
	ResendRankWorkers    int    `json:"resend_rank_workers"`
	ResendHTTPTimeoutSec int    `json:"resend_http_timeout_sec"`
}

// DefaultWebConfig returns default settings.
func DefaultWebConfig() WebConfig {
	return WebConfig{
		Host:                 "0.0.0.0",
		Port:                 8080,
		ResendBatchSize:      100,
		ResendBatchWorkers:   2,
		ResendRankWorkers:    4,
		ResendHTTPTimeoutSec: 10,
	}
}

// LoadWebConfig loads config from file and applies defaults and env overrides.
func LoadWebConfig(path string) (*WebConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file failed: %w", err)
	}

	cfg := DefaultWebConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config json failed: %w", err)
	}

	applyWebDefaults(&cfg)
	ApplyWebEnvOverrides(&cfg)

	return &cfg, nil
}

// ApplyWebEnvOverrides overrides config with environment variables.
func ApplyWebEnvOverrides(cfg *WebConfig) {
	if cfg == nil {
		return
	}

	if host := strings.TrimSpace(os.Getenv("WEB_HOST")); host != "" {
		cfg.Host = host
	}

	if port := getEnvInt("WEB_PORT"); port != nil {
		cfg.Port = *port
	} else if port := getEnvInt("PORT"); port != nil {
		cfg.Port = *port
	}

	if v := getEnvInt("RESEND_BATCH_SIZE"); v != nil {
		cfg.ResendBatchSize = *v
	}
	if v := getEnvInt("RESEND_BATCH_WORKERS"); v != nil {
		cfg.ResendBatchWorkers = *v
	}
	if v := getEnvInt("RESEND_RANK_WORKERS"); v != nil {
		cfg.ResendRankWorkers = *v
	}
	if v := getEnvInt("RESEND_HTTP_TIMEOUT_SEC"); v != nil {
		cfg.ResendHTTPTimeoutSec = *v
	}

	applyWebDefaults(cfg)
}

func applyWebDefaults(cfg *WebConfig) {
	if strings.TrimSpace(cfg.Host) == "" {
		cfg.Host = "0.0.0.0"
	}
	if cfg.Port <= 0 {
		cfg.Port = 8080
	}
	if cfg.ResendBatchSize <= 0 {
		cfg.ResendBatchSize = 100
	}
	if cfg.ResendBatchWorkers <= 0 {
		cfg.ResendBatchWorkers = 2
	}
	if cfg.ResendRankWorkers <= 0 {
		cfg.ResendRankWorkers = 4
	}
	if cfg.ResendHTTPTimeoutSec <= 0 {
		cfg.ResendHTTPTimeoutSec = 10
	}
}

func getEnvInt(name string) *int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return nil
	}
	return &val
}
