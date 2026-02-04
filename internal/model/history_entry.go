package model

import (
	"time"
)

// HistoryEntry 历史记录条目（用于Repository层）
type HistoryEntry struct {
	ID            string        `json:"id"`
	Mode          string       `json:"mode"`
	Archive       bool         `json:"archive"`
	Files         []FileInfo   `json:"files"`
	TotalFiles    int          `json:"totalFiles"`
	ProcessedFiles int         `json:"processedFiles"`
	ErrorFiles    int          `json:"errorFiles"`
	Results       []SheetResult `json:"results"`
	Status        string       `json:"status"`
	Error         string       `json:"error,omitempty"`
	CreatedAt     time.Time    `json:"createdAt"`
	CompletedAt   *time.Time   `json:"completedAt,omitempty"`
	ResultPath    string       `json:"resultPath"`
}
