package handlers

import (
	"fmt"
	"path/filepath"
	"redis_data/api/models"
	"redis_data/service/analytics"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// AnalyticsHandler 数据分析处理器
type AnalyticsHandler struct {
	JobManager *models.JobManager
	ResultDir  string
}

// NewAnalyticsHandler 创建数据分析处理器
func NewAnalyticsHandler(jm *models.JobManager, resultDir string) *AnalyticsHandler {
	return &AnalyticsHandler{
		JobManager: jm,
		ResultDir:  resultDir,
	}
}

// GetStatistics 获取统计数据
func (h *AnalyticsHandler) GetStatistics(c *gin.Context) {
	jobID := c.Query("jobId")
	sheetName := c.Query("sheet")

	if jobID == "" {
		c.JSON(400, gin.H{"error": "jobId is required"})
		return
	}

	// 尝试从内存中获取任务
	job, exists := h.JobManager.GetJob(jobID)
	var excelPath string

	if exists {
		if job.GetStatus() != models.JobStatusCompleted {
			c.JSON(400, gin.H{"error": "Job not completed"})
			return
		}
		excelPath = filepath.Join(h.ResultDir, fmt.Sprintf("%s.xlsx", jobID))
	} else {
		// 从历史记录或文件系统查找
		excelPath = filepath.Join(h.ResultDir, fmt.Sprintf("%s.xlsx", jobID))
	}

	// 分析Excel文件
	stats, err := analytics.AnalyzeExcelFile(excelPath, sheetName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to analyze file: " + err.Error()})
		return
	}

	c.JSON(200, stats)
}

// GetSheetList 获取工作表列表
func (h *AnalyticsHandler) GetSheetList(c *gin.Context) {
	jobID := c.Query("jobId")

	if jobID == "" {
		c.JSON(400, gin.H{"error": "jobId is required"})
		return
	}

	excelPath := filepath.Join(h.ResultDir, fmt.Sprintf("%s.xlsx", jobID))

	// 打开Excel文件获取工作表列表
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		c.JSON(404, gin.H{"error": "File not found"})
		return
	}
	defer f.Close()

	sheets := f.GetSheetList()
	sheetInfo := make([]map[string]interface{}, 0, len(sheets))

	for _, sheetName := range sheets {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			continue
		}

		rowCount := len(rows)
		if rowCount > 0 {
			rowCount-- // 减去表头
		}

		sheetInfo = append(sheetInfo, map[string]interface{}{
			"sheetName": sheetName,
			"rowCount":  rowCount,
		})
	}

	c.JSON(200, gin.H{
		"sheets": sheetInfo,
	})
}
