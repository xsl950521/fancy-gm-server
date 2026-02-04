package handlers

import (
	"fmt"
	"path/filepath"
	"redis_data/api/models"
	"redis_data/api/validators"
	"redis_data/pkg/logger"
	"redis_data/pkg/response"
	"redis_data/service/history"
	"redis_data/service/parser"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

// ProcessHandler 处理解析任务
type ProcessHandler struct {
	JobManager     *models.JobManager
	HistoryManager *history.HistoryManager
	ResultDir      string
}

// NewProcessHandler 创建处理处理器
func NewProcessHandler(jm *models.JobManager, hm *history.HistoryManager, resultDir string) *ProcessHandler {
	return &ProcessHandler{
		JobManager:     jm,
		HistoryManager: hm,
		ResultDir:      resultDir,
	}
}

// ProcessJob 处理解析任务
func (h *ProcessHandler) ProcessJob(c *gin.Context) {
	var req validators.ProcessJobRequest
	if !validators.ValidateURI(c, &req) {
		return
	}

	job, exists := h.JobManager.GetJob(req.JobID)
	if !exists {
		response.NotFound(c, "任务不存在")
		return
	}

	// 如果任务已经在处理中，返回当前状态
	if job.GetStatus() == models.JobStatusProcessing {
		response.Success(c, gin.H{
			"jobId":  job.ID,
			"status": job.Status,
		})
		return
	}

	// 异步处理任务
	go h.processJobAsync(job)

	response.Success(c, gin.H{
		"jobId":  job.ID,
		"status": models.JobStatusProcessing,
	})
}

// processJobAsync 异步处理任务
func (h *ProcessHandler) processJobAsync(job *models.Job) {
	job.SetStatus(models.JobStatusProcessing)
	if err := h.JobManager.UpdateJob(job); err != nil {
		logger.Logger.Error("Failed to update job status", zap.String("job_id", job.ID), zap.Error(err))
	}

	// 创建Excel文件
	excelFile := excelize.NewFile()
	defer excelFile.Close()

	// 创建解析器
	p, err := parser.NewParser(job.Mode)
	if err != nil {
		job.SetError(err)
		_ = h.JobManager.UpdateJob(job)
		return
	}

	// 处理文件
	if err := p.ProcessFiles(job.Files, excelFile, job); err != nil {
		job.SetError(err)
		_ = h.JobManager.UpdateJob(job)
		return
	}

	// 保存Excel文件
	resultPath := filepath.Join(h.ResultDir, fmt.Sprintf("%s.xlsx", job.ID))
	if err := parser.SaveExcelFile(excelFile, resultPath); err != nil {
		job.SetError(err)
		_ = h.JobManager.UpdateJob(job)
		return
	}

	// 提取工作表信息
	results, err := parser.ExtractSheetResults(resultPath)
	if err != nil {
		job.SetError(err)
		_ = h.JobManager.UpdateJob(job)
		return
	}

	// 更新结果
	job.SetResults(results)
	job.SetStatus(models.JobStatusCompleted)
	if err := h.JobManager.UpdateJob(job); err != nil {
		logger.Logger.Error("Failed to update job", zap.String("job_id", job.ID), zap.Error(err))
	}

	// 保存到历史记录
	if h.HistoryManager != nil {
		if err := h.HistoryManager.SaveHistory(job, resultPath); err != nil {
			logger.Logger.Error("Failed to save history", zap.String("job_id", job.ID), zap.Error(err))
		}
	}
}

// GetJobStatus 获取任务状态
func (h *ProcessHandler) GetJobStatus(c *gin.Context) {
	var req validators.GetJobStatusRequest
	if !validators.ValidateURI(c, &req) {
		return
	}

	job, exists := h.JobManager.GetJob(req.JobID)
	if !exists {
		response.NotFound(c, "任务不存在")
		return
	}

	response.Success(c, job)
}
