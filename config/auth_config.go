package config

import (
	"time"
)

// AuthConfig 认证配置
type AuthConfig struct {
	JWT struct {
		SecretKey  string        `yaml:"secret_key"`  // JWT密钥
		Expiration time.Duration `yaml:"expiration"`  // Token过期时间
		RefreshExp time.Duration `yaml:"refresh_exp"` // 刷新Token过期时间
		Issuer     string        `yaml:"issuer"`      // 签发者
	} `yaml:"jwt"`
	Password struct {
		MinLength int `yaml:"min_length"` // 密码最小长度
		RequireUpper bool `yaml:"require_upper"` // 要求大写字母
		RequireLower bool `yaml:"require_lower"` // 要求小写字母
		RequireNumber bool `yaml:"require_number"` // 要求数字
		RequireSpecial bool `yaml:"require_special"` // 要求特殊字符
	} `yaml:"password"`
	Session struct {
		MaxLoginAttempts int           `yaml:"max_login_attempts"` // 最大登录尝试次数
		LockoutDuration  time.Duration `yaml:"lockout_duration"`   // 锁定持续时间
	} `yaml:"session"`
}

// DefaultAuthConfig 返回默认认证配置
func DefaultAuthConfig() AuthConfig {
	cfg := AuthConfig{}
	cfg.JWT.SecretKey = "change-me-in-production" // 生产环境必须修改
	cfg.JWT.Expiration = 24 * time.Hour
	cfg.JWT.RefreshExp = 7 * 24 * time.Hour
	cfg.JWT.Issuer = "gm-server"
	cfg.Password.MinLength = 8
	cfg.Password.RequireUpper = false
	cfg.Password.RequireLower = true
	cfg.Password.RequireNumber = false
	cfg.Password.RequireSpecial = false
	cfg.Session.MaxLoginAttempts = 5
	cfg.Session.LockoutDuration = 30 * time.Minute
	return cfg
}
