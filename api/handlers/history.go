package handlers

import (
	"fmt"
	"redis_data/service/history"

	"github.com/gin-gonic/gin"
)

// HistoryHandler 历史记录处理器
type HistoryHandler struct {
	HistoryManager *history.HistoryManager
}

// NewHistoryHandler 创建历史记录处理器
func NewHistoryHandler(hm *history.HistoryManager) *HistoryHandler {
	return &HistoryHandler{
		HistoryManager: hm,
	}
}

// GetHistories 获取历史记录列表
func (h *HistoryHandler) GetHistories(c *gin.Context) {
	limit := 100 // 默认返回100条
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := parseInt(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	histories, err := h.HistoryManager.GetHistories(limit)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to load histories: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"histories": histories,
		"total":     len(histories),
	})
}

// GetHistory 获取单个历史记录详情
func (h *HistoryHandler) GetHistory(c *gin.Context) {
	id := c.Param("id")

	entry, err := h.HistoryManager.GetHistory(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "History not found"})
		return
	}

	c.JSON(200, entry)
}

// DeleteHistory 删除历史记录
func (h *HistoryHandler) DeleteHistory(c *gin.Context) {
	id := c.Param("id")

	if err := h.HistoryManager.DeleteHistory(id); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete history: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "History deleted successfully"})
}

// parseInt 解析整数（辅助函数）
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
