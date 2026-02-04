package middleware

import (
	"bytes"
	"io"
	"redis_data/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger 请求日志中间件
// 记录请求参数、响应时间、响应状态
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()

		// 获取请求ID（如果已设置）
		requestID, _ := c.Get(RequestIDKey)
		requestIDStr := ""
		if id, ok := requestID.(string); ok {
			requestIDStr = id
		}

		// 记录请求体（仅对非文件上传的POST/PUT请求）
		var requestBody string
		if method == "POST" || method == "PUT" || method == "PATCH" {
			// 检查是否是multipart/form-data（文件上传）
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && len(contentType) >= 19 && contentType[:19] == "multipart/form-data" {
				requestBody = "[文件上传，跳过请求体记录]"
			} else {
				// 读取请求体
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err == nil && len(bodyBytes) > 0 {
					// 限制请求体日志长度（避免日志过大）
					maxBodyLen := 1024
					if len(bodyBytes) > maxBodyLen {
						requestBody = string(bodyBytes[:maxBodyLen]) + "...[截断]"
					} else {
						requestBody = string(bodyBytes)
					}
					// 恢复请求体，供后续处理使用
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}
		}

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 构建日志字段
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.String("ip", clientIP),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// 添加请求ID
		if requestIDStr != "" {
			fields = append(fields, zap.String("request_id", requestIDStr))
		}

		// 添加请求体（如果存在且不是文件上传）
		if requestBody != "" {
			fields = append(fields, zap.String("request_body", requestBody))
		}

		// 添加错误信息（如果有）
		if len(c.Errors) > 0 {
			fields = append(fields, zap.Strings("errors", c.Errors.Errors()))
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			logger.Logger.Error("HTTP Request", fields...)
		} else if statusCode >= 400 {
			logger.Logger.Warn("HTTP Request", fields...)
		} else {
			logger.Logger.Info("HTTP Request", fields...)
		}
	}
}
