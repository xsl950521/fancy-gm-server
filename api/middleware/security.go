package middleware

import (
	"redis_data/pkg/errors"
	"redis_data/pkg/logger"
	"redis_data/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SecurityConfig 安全中间件配置
type SecurityConfig struct {
	MaxRequestBodySize int64  // 最大请求体大小（字节），0表示不限制
	MaxHeaderSize      int    // 最大请求头大小（字节），0表示不限制
	EnableRateLimit    bool   // 是否启用限流（暂未实现）
}

// DefaultSecurityConfig 返回默认安全配置
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		MaxRequestBodySize: 10 * 1024 * 1024, // 默认10MB
		MaxHeaderSize:      8192,              // 默认8KB
		EnableRateLimit:    false,
	}
}

// Security 安全中间件
// 提供请求体大小限制、基础安全防护
func Security(config SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求体大小
		if config.MaxRequestBodySize > 0 {
			if c.Request.ContentLength > config.MaxRequestBodySize {
				logger.Logger.Warn("Request body too large",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Int64("content_length", c.Request.ContentLength),
					zap.Int64("max_size", config.MaxRequestBodySize),
				)

				appErr := errors.New(
					errors.ErrCodeInvalidParam,
					"请求体过大，最大允许大小: "+formatBytes(config.MaxRequestBodySize),
				)
				response.BadRequest(c, appErr.Message)
				c.Abort()
				return
			}
		}

		// 检查请求头大小
		if config.MaxHeaderSize > 0 {
			headerSize := 0
			for key, values := range c.Request.Header {
				headerSize += len(key)
				for _, value := range values {
					headerSize += len(value)
				}
			}

			if headerSize > config.MaxHeaderSize {
				logger.Logger.Warn("Request header too large",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Int("header_size", headerSize),
					zap.Int("max_size", config.MaxHeaderSize),
				)

				appErr := errors.New(
					errors.ErrCodeInvalidParam,
					"请求头过大，最大允许大小: "+formatBytes(int64(config.MaxHeaderSize)),
				)
				response.BadRequest(c, appErr.Message)
				c.Abort()
				return
			}
		}

		// 设置安全响应头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		c.Next()
	}
}

// formatBytes 格式化字节大小
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatInt(bytes/div, 10) + " " + string([]byte{'K', 'M', 'G', 'T', 'P', 'E'}[exp]) + "B"
}
