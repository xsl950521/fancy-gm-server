package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// LoadAppConfig 使用Viper加载应用配置（支持JSON、YAML、TOML等格式）
func LoadAppConfig(configPath string) (*AppConfig, error) {
	cfg := DefaultAppConfig()

	// 如果指定了配置文件路径
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// 默认查找配置文件
		viper.SetConfigName("config") // 配置文件名称（不含扩展名）
		viper.SetConfigType("yaml")   // 默认类型
		viper.AddConfigPath(".")      // 当前目录
		viper.AddConfigPath("./config") // config目录
		viper.AddConfigPath("/etc/gm-server/") // 系统配置目录
	}

	// 设置环境变量前缀
	viper.SetEnvPrefix("GM")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // 自动读取环境变量

	// 读取配置文件（如果存在）
	if err := viper.ReadInConfig(); err != nil {
		// 配置文件不存在时使用默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 绑定环境变量
	bindEnvVars()

	// 解析配置
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 应用默认值
	applyDefaults(&cfg)

	// 应用环境变量覆盖
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

// bindEnvVars 绑定环境变量
func bindEnvVars() {
	// Server配置
	viper.BindEnv("server.host", "GM_SERVER_HOST", "SERVER_HOST")
	viper.BindEnv("server.port", "GM_SERVER_PORT", "SERVER_PORT", "PORT")

	// Database配置
	viper.BindEnv("database.host", "GM_DB_HOST", "DB_HOST")
	viper.BindEnv("database.port", "GM_DB_PORT", "DB_PORT")
	viper.BindEnv("database.user", "GM_DB_USER", "DB_USER")
	viper.BindEnv("database.password", "GM_DB_PASSWORD", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "GM_DB_NAME", "DB_NAME")
	viper.BindEnv("database.sslmode", "GM_DB_SSLMODE", "DB_SSLMODE")
	viper.BindEnv("database.max_open_conns", "GM_DB_MAX_OPEN_CONNS", "DB_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_idle_conns", "GM_DB_MAX_IDLE_CONNS", "DB_MAX_IDLE_CONNS")

	// Log配置
	viper.BindEnv("log.level", "GM_LOG_LEVEL", "LOG_LEVEL")
	viper.BindEnv("log.encoding", "GM_LOG_ENCODING", "LOG_ENCODING")
	viper.BindEnv("log.output_path", "GM_LOG_OUTPUT_PATH", "LOG_OUTPUT_PATH")
	viper.BindEnv("log.error_path", "GM_LOG_ERROR_PATH", "LOG_ERROR_PATH")

	// Web配置
	viper.BindEnv("web.host", "GM_WEB_HOST", "WEB_HOST")
	viper.BindEnv("web.port", "GM_WEB_PORT", "WEB_PORT")
	viper.BindEnv("web.resend_batch_size", "GM_RESEND_BATCH_SIZE", "RESEND_BATCH_SIZE")
	viper.BindEnv("web.resend_batch_workers", "GM_RESEND_BATCH_WORKERS", "RESEND_BATCH_WORKERS")
	viper.BindEnv("web.resend_rank_workers", "GM_RESEND_RANK_WORKERS", "RESEND_RANK_WORKERS")
	viper.BindEnv("web.resend_http_timeout_sec", "GM_RESEND_HTTP_TIMEOUT_SEC", "RESEND_HTTP_TIMEOUT_SEC")

	// Auth配置
	viper.BindEnv("auth.jwt.secret_key", "GM_JWT_SECRET_KEY", "JWT_SECRET_KEY")
	viper.BindEnv("auth.jwt.expiration", "GM_JWT_EXPIRATION", "JWT_EXPIRATION")
	viper.BindEnv("auth.jwt.refresh_exp", "GM_JWT_REFRESH_EXP", "JWT_REFRESH_EXP")
}

// applyDefaults 应用默认值
func applyDefaults(cfg *AppConfig) {
	if strings.TrimSpace(cfg.Server.Host) == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port <= 0 {
		cfg.Server.Port = 8080
	}

	// Database默认值
	if strings.TrimSpace(cfg.Database.Host) == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port <= 0 {
		cfg.Database.Port = 5432
	}
	if strings.TrimSpace(cfg.Database.User) == "" {
		cfg.Database.User = "postgres"
	}
	if strings.TrimSpace(cfg.Database.DBName) == "" {
		cfg.Database.DBName = "gm_server"
	}
	if strings.TrimSpace(cfg.Database.SSLMode) == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Database.MaxOpenConns <= 0 {
		cfg.Database.MaxOpenConns = 100
	}
	if cfg.Database.MaxIdleConns <= 0 {
		cfg.Database.MaxIdleConns = 10
	}

	// Log默认值
	if strings.TrimSpace(cfg.Log.Level) == "" {
		cfg.Log.Level = "info"
	}
	if strings.TrimSpace(cfg.Log.Encoding) == "" {
		cfg.Log.Encoding = "json"
	}

	// Web默认值
	if strings.TrimSpace(cfg.Web.Host) == "" {
		cfg.Web.Host = "0.0.0.0"
	}
	if cfg.Web.Port <= 0 {
		cfg.Web.Port = 8080
	}
	if cfg.Web.ResendBatchSize <= 0 {
		cfg.Web.ResendBatchSize = 100
	}
	if cfg.Web.ResendBatchWorkers <= 0 {
		cfg.Web.ResendBatchWorkers = 2
	}
	if cfg.Web.ResendRankWorkers <= 0 {
		cfg.Web.ResendRankWorkers = 4
	}
	if cfg.Web.ResendHTTPTimeoutSec <= 0 {
		cfg.Web.ResendHTTPTimeoutSec = 10
	}

	// Auth默认值
	if cfg.Auth.JWT.SecretKey == "" {
		cfg.Auth.JWT.SecretKey = "change-me-in-production"
	}
	if cfg.Auth.JWT.Expiration == 0 {
		cfg.Auth.JWT.Expiration = 24 * time.Hour
	}
	if cfg.Auth.JWT.RefreshExp == 0 {
		cfg.Auth.JWT.RefreshExp = 7 * 24 * time.Hour
	}
	if cfg.Auth.JWT.Issuer == "" {
		cfg.Auth.JWT.Issuer = "gm-server"
	}
	if cfg.Auth.Password.MinLength <= 0 {
		cfg.Auth.Password.MinLength = 8
	}
	if cfg.Auth.Session.MaxLoginAttempts <= 0 {
		cfg.Auth.Session.MaxLoginAttempts = 5
	}
	if cfg.Auth.Session.LockoutDuration == 0 {
		cfg.Auth.Session.LockoutDuration = 30 * time.Minute
	}
}

// applyEnvOverrides 应用环境变量覆盖（兼容旧的环境变量格式）
func applyEnvOverrides(cfg *AppConfig) {
	// 兼容旧的环境变量格式
	if host := os.Getenv("WEB_HOST"); host != "" {
		cfg.Web.Host = host
	}
	if port := getEnvIntFromString("WEB_PORT"); port != nil {
		cfg.Web.Port = *port
	} else if port := getEnvIntFromString("PORT"); port != nil {
		cfg.Web.Port = *port
		cfg.Server.Port = *port
	}
}

// getEnvIntFromString 从环境变量获取整数（内部使用，避免与web_config.go冲突）
func getEnvIntFromString(name string) *int {
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

// SaveConfig 保存配置到文件
func SaveConfig(cfg *AppConfig, path string) error {
	viper.Set("server", cfg.Server)
	viper.Set("database", cfg.Database)
	viper.Set("log", cfg.Log)
	viper.Set("web", cfg.Web)
	return viper.WriteConfigAs(path)
}

// WatchAppConfig 监听配置文件变化（热重载）
func WatchAppConfig(callback func(*AppConfig)) error {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		cfg, err := LoadAppConfig("")
		if err != nil {
			return
		}
		callback(cfg)
	})
	return nil
}
