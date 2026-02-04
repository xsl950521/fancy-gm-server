package handlers

import (
	"redis_data/api/validators"
	"redis_data/pkg/response"
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

// GetHistories 获取历史记录列表（分页）
func (h *HistoryHandler) GetHistories(c *gin.Context) {
	var req validators.GetHistoryRequest
	// 设置默认值
	req.Page = 1
	req.PageSize = 100
	if !validators.ValidateQuery(c, &req) {
		return
	}

	histories, total, err := h.HistoryManager.GetHistories(req.Page, req.PageSize)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.SuccessPage(c, histories, total, req.Page, req.PageSize)
}

// GetHistory 获取单个历史记录详情
func (h *HistoryHandler) GetHistory(c *gin.Context) {
	var req validators.GetHistoryDetailRequest
	if !validators.ValidateURI(c, &req) {
		return
	}

	entry, err := h.HistoryManager.GetHistory(req.ID)
	if err != nil {
		response.NotFound(c, "历史记录不存在")
		return
	}

	response.Success(c, entry)
}

// DeleteHistory 删除历史记录
func (h *HistoryHandler) DeleteHistory(c *gin.Context) {
	var req validators.DeleteHistoryRequest
	if !validators.ValidateURI(c, &req) {
		return
	}

	if err := h.HistoryManager.DeleteHistory(req.ID); err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, "历史记录删除成功")
}
