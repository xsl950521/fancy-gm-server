package handlers

import (
	"redis_data/api/middleware"
	"redis_data/pkg/response"

	"github.com/gin-gonic/gin"
)

// MetricsHandler 性能监控处理器
type MetricsHandler struct{}

// NewMetricsHandler 创建性能监控处理器
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// GetMetrics 获取性能统计信息
// GET /api/metrics
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	metrics := middleware.GetMetrics()
	response.Success(c, metrics)
}

// ResetMetrics 重置性能统计信息
// POST /api/metrics/reset
func (h *MetricsHandler) ResetMetrics(c *gin.Context) {
	middleware.ResetMetrics()
	response.Success(c, "性能统计已重置")
}
