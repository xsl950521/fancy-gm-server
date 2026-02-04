package model

import (
	"time"
)

// Job 任务模型
type Job struct {
	ID            string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Status        string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Mode          string    `gorm:"type:varchar(50)" json:"mode"`
	Archive       bool      `gorm:"default:false" json:"archive"`
	Progress      int       `gorm:"default:0" json:"progress"`
	TotalFiles    int       `gorm:"default:0" json:"totalFiles"`
	ProcessedFiles int      `gorm:"default:0" json:"processedFiles"`
	ErrorFiles    int       `gorm:"default:0" json:"errorFiles"`
	Error         string    `gorm:"type:text" json:"error,omitempty"`
	ResultPath    string    `gorm:"type:varchar(500)" json:"resultPath,omitempty"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`

	// 关联
	Files   []JobFile   `gorm:"foreignKey:JobID;constraint:OnDelete:CASCADE" json:"files"`
	Results []JobResult `gorm:"foreignKey:JobID;constraint:OnDelete:CASCADE" json:"results"`
}

// JobFile 任务文件
type JobFile struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	JobID    string `gorm:"type:varchar(64);not null;index" json:"jobId"`
	Filename string `gorm:"type:varchar(255);not null" json:"filename"`
	Size     int64  `gorm:"not null" json:"size"`
	Type     string `gorm:"type:varchar(50)" json:"type"`
	Path     string `gorm:"type:varchar(500)" json:"path"`
}

// JobResult 任务结果
type JobResult struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	JobID     string `gorm:"type:varchar(64);not null;index" json:"jobId"`
	SheetName string `gorm:"type:varchar(255);not null" json:"sheetName"`
	RowCount  int    `gorm:"default:0" json:"rowCount"`
	Columns   string `gorm:"type:text" json:"columns"` // JSON格式存储
	Data      string `gorm:"type:text" json:"data,omitempty"` // JSON格式存储，可选
}

// TableName 指定表名
func (Job) TableName() string {
	return "jobs"
}

func (JobFile) TableName() string {
	return "job_files"
}

func (JobResult) TableName() string {
	return "job_results"
}
