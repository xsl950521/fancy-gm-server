package model

import (
	"time"
)

// History 历史记录模型
type History struct {
	ID            string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	JobID         string    `gorm:"type:varchar(64);not null;index" json:"jobId"`
	Mode          string    `gorm:"type:varchar(50)" json:"mode"`
	Archive       bool      `gorm:"default:false" json:"archive"`
	TotalFiles    int       `gorm:"default:0" json:"totalFiles"`
	ProcessedFiles int      `gorm:"default:0" json:"processedFiles"`
	ErrorFiles    int       `gorm:"default:0" json:"errorFiles"`
	Status        string    `gorm:"type:varchar(20)" json:"status"`
	Error         string    `gorm:"type:text" json:"error,omitempty"`
	ResultPath    string    `gorm:"type:varchar(500)" json:"resultPath"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index" json:"createdAt"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`

	// 关联
	Files   []HistoryFile   `gorm:"foreignKey:HistoryID;constraint:OnDelete:CASCADE" json:"files"`
	Results []HistoryResult `gorm:"foreignKey:HistoryID;constraint:OnDelete:CASCADE" json:"results"`
}

// HistoryFile 历史记录文件
type HistoryFile struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	HistoryID string `gorm:"type:varchar(64);not null;index" json:"historyId"`
	Filename  string `gorm:"type:varchar(255);not null" json:"filename"`
	Size      int64  `gorm:"not null" json:"size"`
	Type      string `gorm:"type:varchar(50)" json:"type"`
	Path      string `gorm:"type:varchar(500)" json:"path"`
}

// HistoryResult 历史记录结果
type HistoryResult struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	HistoryID string `gorm:"type:varchar(64);not null;index" json:"historyId"`
	SheetName string `gorm:"type:varchar(255);not null" json:"sheetName"`
	RowCount  int    `gorm:"default:0" json:"rowCount"`
	Columns   string `gorm:"type:text" json:"columns"` // JSON格式存储
	Data      string `gorm:"type:text" json:"data,omitempty"` // JSON格式存储，可选
}

// TableName 指定表名
func (History) TableName() string {
	return "histories"
}

func (HistoryFile) TableName() string {
	return "history_files"
}

func (HistoryResult) TableName() string {
	return "history_results"
}
