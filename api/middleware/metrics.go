package middleware

import (
	"redis_data/pkg/logger"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Metrics 性能监控数据
type Metrics struct {
	mu                sync.RWMutex
	requestCounts     map[string]int64      // 接口调用次数
	totalLatency      map[string]time.Duration // 总耗时
	maxLatency        map[string]time.Duration // 最大耗时
	minLatency        map[string]time.Duration // 最小耗时
	slowRequests      []SlowRequest          // 慢请求列表
	maxSlowRequests   int                    // 最大慢请求记录数
	slowThreshold     time.Duration          // 慢请求阈值
}

// SlowRequest 慢请求记录
type SlowRequest struct {
	Path      string
	Method    string
	Latency   time.Duration
	Status    int
	Timestamp time.Time
	RequestID string
}

var globalMetrics *Metrics

func init() {
	globalMetrics = &Metrics{
		requestCounts:   make(map[string]int64),
		totalLatency:    make(map[string]time.Duration),
		maxLatency:      make(map[string]time.Duration),
		minLatency:      make(map[string]time.Duration),
		slowRequests:    make([]SlowRequest, 0),
		maxSlowRequests: 100, // 最多记录100个慢请求
		slowThreshold:   1 * time.Second, // 默认1秒为慢请求
	}
}

// PerformanceMonitor 性能监控中间件
// 记录接口耗时、慢请求、统计调用次数
func PerformanceMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		routeKey := method + " " + path

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 获取请求ID
		requestID, _ := c.Get(RequestIDKey)
		requestIDStr := ""
		if id, ok := requestID.(string); ok {
			requestIDStr = id
		}

		// 更新统计信息
		globalMetrics.mu.Lock()
		globalMetrics.requestCounts[routeKey]++
		globalMetrics.totalLatency[routeKey] += latency

		// 更新最大耗时
		if maxLat, exists := globalMetrics.maxLatency[routeKey]; !exists || latency > maxLat {
			globalMetrics.maxLatency[routeKey] = latency
		}

		// 更新最小耗时
		if minLat, exists := globalMetrics.minLatency[routeKey]; !exists || latency < minLat {
			globalMetrics.minLatency[routeKey] = latency
		}

		// 记录慢请求
		if latency >= globalMetrics.slowThreshold {
			slowReq := SlowRequest{
				Path:      path,
				Method:    method,
				Latency:   latency,
				Status:    statusCode,
				Timestamp: time.Now(),
				RequestID: requestIDStr,
			}

			// 添加到慢请求列表（保持最多maxSlowRequests个）
			globalMetrics.slowRequests = append(globalMetrics.slowRequests, slowReq)
			if len(globalMetrics.slowRequests) > globalMetrics.maxSlowRequests {
				// 移除最旧的记录
				globalMetrics.slowRequests = globalMetrics.slowRequests[1:]
			}

			// 记录慢请求日志
			logger.Logger.Warn("Slow request detected",
				zap.String("method", method),
				zap.String("path", path),
				zap.Duration("latency", latency),
				zap.Int("status", statusCode),
				zap.String("request_id", requestIDStr),
			)
		}
		globalMetrics.mu.Unlock()
	}
}

// GetMetrics 获取性能统计信息
func GetMetrics() map[string]interface{} {
	globalMetrics.mu.RLock()
	defer globalMetrics.mu.RUnlock()

	stats := make(map[string]interface{})
	routeStats := make(map[string]map[string]interface{})

	for route, count := range globalMetrics.requestCounts {
		totalLat := globalMetrics.totalLatency[route]
		avgLat := totalLat / time.Duration(count)
		maxLat := globalMetrics.maxLatency[route]
		minLat := globalMetrics.minLatency[route]

		routeStats[route] = map[string]interface{}{
			"count":        count,
			"avg_latency":  avgLat.String(),
			"max_latency":  maxLat.String(),
			"min_latency":  minLat.String(),
			"total_latency": totalLat.String(),
		}
	}

	stats["routes"] = routeStats
	stats["slow_requests_count"] = len(globalMetrics.slowRequests)
	stats["slow_requests"] = globalMetrics.slowRequests

	return stats
}

// ResetMetrics 重置统计信息
func ResetMetrics() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.requestCounts = make(map[string]int64)
	globalMetrics.totalLatency = make(map[string]time.Duration)
	globalMetrics.maxLatency = make(map[string]time.Duration)
	globalMetrics.minLatency = make(map[string]time.Duration)
	globalMetrics.slowRequests = make([]SlowRequest, 0)
}

// SetSlowThreshold 设置慢请求阈值
func SetSlowThreshold(threshold time.Duration) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.slowThreshold = threshold
}

// SetMaxSlowRequests 设置最大慢请求记录数
func SetMaxSlowRequests(max int) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.maxSlowRequests = max
}
