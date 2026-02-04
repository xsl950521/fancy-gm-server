package middleware

import (
	"redis_data/pkg/errors"
	"redis_data/pkg/logger"
	"redis_data/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			
			// 记录错误日志
			if appErr, ok := errors.AsAppError(err); ok {
				logger.Logger.Error("Request error",
					zap.Int("code", int(appErr.Code)),
					zap.String("message", appErr.Message),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Error(appErr.Err),
				)
			} else {
				logger.Logger.Error("Request error",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Error(err),
				)
			}

			// 返回错误响应
			response.Error(c, err)
			c.Abort()
		}
	}
}

// Recovery 恢复中间件（捕获panic）
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Stack("stack"),
				)

				appErr := errors.New(errors.ErrCodeInternal, "内部服务器错误")
				response.Error(c, appErr)
				c.Abort()
			}
		}()
		c.Next()
	}
}
