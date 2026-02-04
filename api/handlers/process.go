package handlers

import (
	"fmt"
	"path/filepath"
	"redis_data/api/models"
	"redis_data/service/history"
	"redis_data/service/parser"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
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
	jobID := c.Param("jobId")

	job, exists := h.JobManager.GetJob(jobID)
	if !exists {
		c.JSON(404, gin.H{"error": "Job not found"})
		return
	}

	// 如果任务已经在处理中，返回当前状态
	if job.GetStatus() == models.JobStatusProcessing {
		c.JSON(200, gin.H{
			"jobId":  job.ID,
			"status": job.Status,
		})
		return
	}

	// 异步处理任务
	go h.processJobAsync(job)

	c.JSON(200, gin.H{
		"jobId":  job.ID,
		"status": models.JobStatusProcessing,
	})
}

// processJobAsync 异步处理任务
func (h *ProcessHandler) processJobAsync(job *models.Job) {
	job.SetStatus(models.JobStatusProcessing)

	// 创建Excel文件
	excelFile := excelize.NewFile()
	defer excelFile.Close()

	// 创建解析器
	p, err := parser.NewParser(job.Mode)
	if err != nil {
		job.SetError(err)
		return
	}

	// 处理文件
	if err := p.ProcessFiles(job.Files, excelFile, job); err != nil {
		job.SetError(err)
		return
	}

	// 保存Excel文件
	resultPath := filepath.Join(h.ResultDir, fmt.Sprintf("%s.xlsx", job.ID))
	if err := parser.SaveExcelFile(excelFile, resultPath); err != nil {
		job.SetError(err)
		return
	}

	// 提取工作表信息
	results, err := parser.ExtractSheetResults(resultPath)
	if err != nil {
		job.SetError(err)
		return
	}

	// 更新结果（需要添加一个方法来安全地更新Results）
	job.SetResults(results)

	job.SetStatus(models.JobStatusCompleted)

	// 保存到历史记录
	if h.HistoryManager != nil {
		h.HistoryManager.SaveHistory(job, resultPath)
	}
}

// GetJobStatus 获取任务状态
func (h *ProcessHandler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")

	job, exists := h.JobManager.GetJob(jobID)
	if !exists {
		c.JSON(404, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(200, job)
}
