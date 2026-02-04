package middleware

import (
	"redis_data/pkg/audit"
	"redis_data/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditMiddleware 审计日志中间件
func AuditMiddleware(auditService *audit.AuditService, action, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理请求
		c.Next()

		// 获取用户ID（如果已认证）
		var userID *uint
		if uid, exists := c.Get(UserIDKey); exists {
			if id, ok := uid.(uint); ok {
				userID = &id
			}
		}

		// 获取请求信息
		ipAddress := c.ClientIP()
		userAgent := c.Request.UserAgent()
		requestID, _ := c.Get(RequestIDKey)
		requestIDStr := ""
		if id, ok := requestID.(string); ok {
			requestIDStr = id
		}

		// 获取资源ID（从路径参数或查询参数）
		resourceID := c.Param("id")
		if resourceID == "" {
			resourceID = c.Param("jobId")
		}

		// 确定状态
		status := "success"
		if c.Writer.Status() >= 400 {
			status = "failed"
		}

		// 记录审计日志（异步，不阻塞请求）
		go func() {
			if err := auditService.LogAction(userID, action, resource, resourceID, ipAddress, userAgent, requestIDStr, nil, status); err != nil {
				logger.Logger.Error("Failed to log audit",
					zap.String("action", action),
					zap.String("resource", resource),
					zap.Error(err),
				)
			}
		}()
	}
}
