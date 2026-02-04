package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"redis_data/api/models"
	"redis_data/pkg/errors"
	"redis_data/pkg/response"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	JobManager *models.JobManager
	UploadDir  string
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(jm *models.JobManager, uploadDir string) *UploadHandler {
	// 确保上传目录存在
	os.MkdirAll(uploadDir, 0755)
	return &UploadHandler{
		JobManager: jm,
		UploadDir:  uploadDir,
	}
}

// UploadFiles 处理文件上传
func (h *UploadHandler) UploadFiles(c *gin.Context) {
	// 解析表单
	form, err := c.MultipartForm()
	if err != nil {
		response.Error(c, errors.Wrap(err, errors.ErrCodeInvalidParam, "解析表单失败"))
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.BadRequest(c, "未上传文件")
		return
	}

	// 生成任务ID
	jobID := generateJobID()

	// 保存文件并收集文件信息
	fileInfos := make([]models.FileInfo, 0, len(files))
	for _, fileHeader := range files {
		// 验证文件类型
		if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".txt") {
			continue
		}

		// 保存文件
		savedPath, err := h.saveUploadedFile(fileHeader, jobID)
		if err != nil {
			continue
		}

		// 检测文件类型（根据文件名或内容）
		fileType := detectFileType(fileHeader.Filename)

		fileInfos = append(fileInfos, models.FileInfo{
			Filename: fileHeader.Filename,
			Size:     fileHeader.Size,
			Type:     fileType,
			Path:     savedPath,
		})
	}

	if len(fileInfos) == 0 {
		response.BadRequest(c, "没有有效的文件")
		return
	}

	// 获取模式参数
	mode := c.PostForm("mode")
	if mode == "" {
		mode = "all"
	}

	archive := c.PostForm("archive") == "true"

	// 创建任务
	job, err := h.JobManager.CreateJob(jobID, mode, archive, fileInfos)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"jobId":  job.ID,
		"status": job.Status,
		"files":  fileInfos,
	})
}

// saveUploadedFile 保存上传的文件
func (h *UploadHandler) saveUploadedFile(fileHeader *multipart.FileHeader, jobID string) (string, error) {
	// 打开上传的文件
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 创建目标目录
	targetDir := filepath.Join(h.UploadDir, jobID)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	// 创建目标文件
	dstPath := filepath.Join(targetDir, fileHeader.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return dstPath, nil
}

// detectFileType 根据文件名检测文件类型
func detectFileType(filename string) string {
	filename = strings.ToLower(filename)
	if strings.Contains(filename, "total_rank") || strings.Contains(filename, "totalrank") {
		return "total"
	}
	if strings.Contains(filename, "rank") && !strings.Contains(filename, "total") {
		return "daily"
	}
	if strings.Contains(filename, "reward") {
		return "reward"
	}
	if strings.Contains(filename, "mail") {
		return "mail"
	}
	if strings.Contains(filename, "month_rank") || strings.Contains(filename, "monthrank") {
		return "month"
	}
	if strings.Contains(filename, "club_pid") || strings.Contains(filename, "clubpid") {
		return "clubpid"
	}
	return "unknown"
}

// generateJobID 生成任务ID
func generateJobID() string {
	return fmt.Sprintf("job_%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000000)
}
