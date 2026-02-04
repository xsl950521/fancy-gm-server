package config

import (
	"redis_data/pkg/database"
	"redis_data/pkg/logger"
	"time"
)

// AppConfig 应用配置
type AppConfig struct {
	Server     ServerConfig     `yaml:"server"`
	Database   database.Config  `yaml:"database"`
	Log        logger.Config    `yaml:"log"`
	Web        WebConfig        `yaml:"web"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Auth       AuthConfig       `yaml:"auth"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	// 请求日志配置
	RequestLogger struct {
		Enabled bool `yaml:"enabled"` // 是否启用请求日志
	} `yaml:"request_logger"`

	// 性能监控配置
	PerformanceMonitor struct {
		Enabled       bool          `yaml:"enabled"`        // 是否启用性能监控
		SlowThreshold time.Duration `yaml:"slow_threshold"` // 慢请求阈值（如：1s）
		MaxSlowRequests int         `yaml:"max_slow_requests"` // 最大慢请求记录数
	} `yaml:"performance_monitor"`

	// 安全配置
	Security struct {
		MaxRequestBodySize int64 `yaml:"max_request_body_size"` // 最大请求体大小（字节），0表示不限制
		MaxHeaderSize      int   `yaml:"max_header_size"`       // 最大请求头大小（字节），0表示不限制
	} `yaml:"security"`
}

// DefaultAppConfig 返回默认配置
func DefaultAppConfig() AppConfig {
	return AppConfig{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: database.DefaultConfig(),
		Log:      logger.DefaultConfig(),
		Web:      DefaultWebConfig(),
		Middleware: MiddlewareConfig{
			RequestLogger: struct {
				Enabled bool `yaml:"enabled"`
			}{Enabled: true},
			PerformanceMonitor: struct {
				Enabled         bool          `yaml:"enabled"`
				SlowThreshold  time.Duration `yaml:"slow_threshold"`
				MaxSlowRequests int          `yaml:"max_slow_requests"`
			}{
				Enabled:         true,
				SlowThreshold:  1 * time.Second,
				MaxSlowRequests: 100,
			},
			Security: struct {
				MaxRequestBodySize int64 `yaml:"max_request_body_size"`
				MaxHeaderSize      int   `yaml:"max_header_size"`
			}{
				MaxRequestBodySize: 10 * 1024 * 1024, // 10MB
				MaxHeaderSize:      8192,              // 8KB
			},
		},
		Auth: DefaultAuthConfig(),
	}
}
