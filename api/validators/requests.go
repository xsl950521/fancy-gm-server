package validators

// 定义所有请求参数结构体，统一管理

// ProcessJobRequest 处理任务请求
type ProcessJobRequest struct {
	JobID string `uri:"jobId" binding:"required"`
}

// GetJobStatusRequest 获取任务状态请求
type GetJobStatusRequest struct {
	JobID string `uri:"jobId" binding:"required"`
}

// GetDataRequest 获取数据请求
type GetDataRequest struct {
	JobID    string `form:"jobId" binding:"required"`
	Sheet    string `form:"sheet"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"pageSize" binding:"min=1,max=1000"`
}

// DownloadFileRequest 下载文件请求
type DownloadFileRequest struct {
	JobID    string `uri:"jobId" binding:"required"`
	Format   string `form:"format" binding:"oneof=excel xlsx csv"`
	Sheet    string `form:"sheet"`
}

// GetHistoryRequest 获取历史记录请求
type GetHistoryRequest struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"pageSize" binding:"min=1,max=1000"`
}

// GetHistoryDetailRequest 获取历史记录详情请求
type GetHistoryDetailRequest struct {
	ID string `uri:"id" binding:"required"`
}

// DeleteHistoryRequest 删除历史记录请求
type DeleteHistoryRequest struct {
	ID string `uri:"id" binding:"required"`
}

// GetStatisticsRequest 获取统计数据请求
type GetStatisticsRequest struct {
	JobID string `form:"jobId"`
}

// GetSheetListRequest 获取工作表列表请求
type GetSheetListRequest struct {
	JobID string `form:"jobId" binding:"required"`
}
