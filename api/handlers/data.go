package handlers

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"redis_data/api/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// DataHandler 数据查询和下载处理器
type DataHandler struct {
	JobManager *models.JobManager
	ResultDir  string
}

// NewDataHandler 创建数据处理器
func NewDataHandler(jm *models.JobManager, resultDir string) *DataHandler {
	return &DataHandler{
		JobManager: jm,
		ResultDir:  resultDir,
	}
}

// GetData 获取数据（分页）
func (h *DataHandler) GetData(c *gin.Context) {
	jobID := c.Query("jobId")
	sheetName := c.Query("sheet")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	job, exists := h.JobManager.GetJob(jobID)
	if !exists {
		c.JSON(404, gin.H{"error": "Job not found"})
		return
	}

	if job.GetStatus() != models.JobStatusCompleted {
		c.JSON(400, gin.H{"error": "Job not completed"})
		return
	}

	// 读取Excel文件
	excelPath := filepath.Join(h.ResultDir, fmt.Sprintf("%s.xlsx", jobID))
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to open excel file: " + err.Error()})
		return
	}
	defer f.Close()

	// 如果没有指定工作表，使用第一个
	if sheetName == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			c.JSON(404, gin.H{"error": "No sheets found"})
			return
		}
		sheetName = sheets[0]
	}

	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read sheet: " + err.Error()})
		return
	}

	if len(rows) == 0 {
		c.JSON(200, gin.H{
			"total":    0,
			"page":     page,
			"pageSize": pageSize,
			"data":     []interface{}{},
			"columns":  []string{},
		})
		return
	}

	// 第一行是表头
	columns := rows[0]
	dataRows := rows[1:]

	// 分页
	total := len(dataRows)
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	if start >= total {
		c.JSON(200, gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
			"data":     []interface{}{},
			"columns":  columns,
		})
		return
	}

	// 转换为接口数组
	data := make([]interface{}, 0, end-start)
	for i := start; i < end; i++ {
		row := dataRows[i]
		rowData := make(map[string]interface{})
		for j, col := range columns {
			if j < len(row) {
				rowData[col] = row[j]
			} else {
				rowData[col] = ""
			}
		}
		data = append(data, rowData)
	}

	c.JSON(200, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     data,
		"columns":  columns,
	})
}

// DownloadFile 下载文件
func (h *DataHandler) DownloadFile(c *gin.Context) {
	jobID := c.Param("jobId")
	format := c.DefaultQuery("format", "excel")
	sheetName := c.Query("sheet")

	// 先尝试从内存中获取任务
	job, exists := h.JobManager.GetJob(jobID)
	var excelPath string

	if exists {
		// 任务在内存中
		if job.GetStatus() != models.JobStatusCompleted {
			c.JSON(400, gin.H{"error": "Job not completed"})
			return
		}
		excelPath = filepath.Join(h.ResultDir, fmt.Sprintf("%s.xlsx", jobID))
	} else {
		// 任务不在内存中，尝试从历史记录中查找
		// 直接从结果目录查找文件（历史记录已保存文件路径）
		excelPath = filepath.Join(h.ResultDir, fmt.Sprintf("%s.xlsx", jobID))
		
		// 检查文件是否存在
		if _, err := os.Stat(excelPath); os.IsNotExist(err) {
			c.JSON(404, gin.H{"error": "File not found"})
			return
		}
	}

	switch format {
	case "csv":
		h.downloadCSV(c, excelPath, sheetName)
	case "excel", "xlsx":
		h.downloadExcel(c, excelPath)
	default:
		c.JSON(400, gin.H{"error": "Unsupported format"})
	}
}

// downloadExcel 下载Excel文件
func (h *DataHandler) downloadExcel(c *gin.Context, excelPath string) {
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(excelPath)))
	c.File(excelPath)
}

// downloadCSV 下载CSV文件
func (h *DataHandler) downloadCSV(c *gin.Context, excelPath, sheetName string) {
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to open excel file"})
		return
	}
	defer f.Close()

	// 如果没有指定工作表，使用第一个
	if sheetName == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			c.JSON(404, gin.H{"error": "No sheets found"})
			return
		}
		sheetName = sheets[0]
	}

	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read sheet"})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", strings.ReplaceAll(sheetName, " ", "_")))

	// 写入CSV
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 写入BOM（支持Excel正确显示中文）
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	for _, row := range rows {
		writer.Write(row)
	}
}
