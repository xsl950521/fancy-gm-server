package models

import (
	"sync"
	"time"
)

// JobStatus 任务状态
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// Job 解析任务
type Job struct {
	ID            string                 `json:"id"`
	Status        JobStatus              `json:"status"`
	Mode          string                 `json:"mode"`
	Archive       bool                   `json:"archive"`
	Files         []FileInfo             `json:"files"`
	Progress      int                    `json:"progress"`
	TotalFiles    int                    `json:"totalFiles"`
	ProcessedFiles int                   `json:"processedFiles"`
	ErrorFiles    int                    `json:"errorFiles"`
	Results       []SheetResult          `json:"results"`
	Error         string                 `json:"error,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	CompletedAt   *time.Time             `json:"completedAt,omitempty"`
	mu            sync.RWMutex            // 保护并发访问
}

// FileInfo 文件信息
type FileInfo struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Type     string `json:"type"` // daily, total, reward, etc.
	Path     string `json:"path"` // 服务器端路径
}

// SheetResult 工作表结果
type SheetResult struct {
	SheetName string   `json:"sheetName"`
	RowCount  int      `json:"rowCount"`
	Columns   []string `json:"columns"`
	Data      [][]interface{} `json:"data,omitempty"` // 可选：实际数据
}

// UpdateProgress 更新进度
func (j *Job) UpdateProgress(processed, total int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.ProcessedFiles = processed
	j.TotalFiles = total
	if total > 0 {
		j.Progress = (processed * 100) / total
	}
}

// SetStatus 设置状态
func (j *Job) SetStatus(status JobStatus) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = status
	if status == JobStatusCompleted || status == JobStatusFailed {
		now := time.Now()
		j.CompletedAt = &now
	}
}

// SetError 设置错误
func (j *Job) SetError(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = JobStatusFailed
	if err != nil {
		j.Error = err.Error()
	}
	now := time.Now()
	j.CompletedAt = &now
}

// GetStatus 获取状态（线程安全）
func (j *Job) GetStatus() JobStatus {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Status
}

// SetResults 设置结果（线程安全）
func (j *Job) SetResults(results []SheetResult) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Results = results
}
