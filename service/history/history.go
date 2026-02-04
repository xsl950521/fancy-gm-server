package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"redis_data/api/models"
	"sort"
	"sync"
	"time"
)

// HistoryManager 历史记录管理器
type HistoryManager struct {
	historyFile string
	mu          sync.RWMutex
}

// HistoryEntry 历史记录条目
type HistoryEntry struct {
	ID            string                 `json:"id"`
	Mode          string                 `json:"mode"`
	Archive       bool                   `json:"archive"`
	Files         []models.FileInfo      `json:"files"`
	TotalFiles    int                    `json:"totalFiles"`
	ProcessedFiles int                   `json:"processedFiles"`
	ErrorFiles    int                    `json:"errorFiles"`
	Results       []models.SheetResult  `json:"results"`
	Status        string                 `json:"status"`
	Error         string                 `json:"error,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	CompletedAt   *time.Time             `json:"completedAt,omitempty"`
	ResultPath    string                 `json:"resultPath"` // Excel文件路径
}

// NewHistoryManager 创建历史记录管理器
func NewHistoryManager(historyDir string) (*HistoryManager, error) {
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	historyFile := filepath.Join(historyDir, "history.json")
	return &HistoryManager{
		historyFile: historyFile,
	}, nil
}

// SaveHistory 保存历史记录
func (hm *HistoryManager) SaveHistory(job *models.Job, resultPath string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// 读取现有历史记录
	histories, err := hm.loadHistories()
	if err != nil {
		histories = []HistoryEntry{}
	}

	// 创建新的历史记录条目
	entry := HistoryEntry{
		ID:            job.ID,
		Mode:          job.Mode,
		Archive:       job.Archive,
		Files:         job.Files,
		TotalFiles:    job.TotalFiles,
		ProcessedFiles: job.ProcessedFiles,
		ErrorFiles:    job.ErrorFiles,
		Results:       job.Results,
		Status:        string(job.GetStatus()),
		CreatedAt:     job.CreatedAt,
		CompletedAt:   job.CompletedAt,
		ResultPath:    resultPath,
	}

	if job.Error != "" {
		entry.Error = job.Error
	}

	// 添加到历史记录列表（按时间倒序）
	histories = append([]HistoryEntry{entry}, histories...)

	// 只保留最近1000条记录
	if len(histories) > 1000 {
		histories = histories[:1000]
	}

	// 保存到文件
	return hm.saveHistories(histories)
}

// GetHistories 获取历史记录列表
func (hm *HistoryManager) GetHistories(limit int) ([]HistoryEntry, error) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	histories, err := hm.loadHistories()
	if err != nil {
		return []HistoryEntry{}, nil
	}

	// 按创建时间倒序排序
	sort.Slice(histories, func(i, j int) bool {
		return histories[i].CreatedAt.After(histories[j].CreatedAt)
	})

	// 限制返回数量
	if limit > 0 && limit < len(histories) {
		histories = histories[:limit]
	}

	return histories, nil
}

// GetHistory 获取单个历史记录
func (hm *HistoryManager) GetHistory(id string) (*HistoryEntry, error) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	histories, err := hm.loadHistories()
	if err != nil {
		return nil, err
	}

	for _, entry := range histories {
		if entry.ID == id {
			return &entry, nil
		}
	}

	return nil, fmt.Errorf("history not found: %s", id)
}

// DeleteHistory 删除历史记录
func (hm *HistoryManager) DeleteHistory(id string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	histories, err := hm.loadHistories()
	if err != nil {
		return err
	}

	// 删除指定记录
	newHistories := []HistoryEntry{}
	for _, entry := range histories {
		if entry.ID != id {
			newHistories = append(newHistories, entry)
		}
	}

	return hm.saveHistories(newHistories)
}

// loadHistories 从文件加载历史记录
func (hm *HistoryManager) loadHistories() ([]HistoryEntry, error) {
	data, err := os.ReadFile(hm.historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []HistoryEntry{}, nil
		}
		return nil, err
	}

	var histories []HistoryEntry
	if err := json.Unmarshal(data, &histories); err != nil {
		return nil, err
	}

	return histories, nil
}

// saveHistories 保存历史记录到文件
func (hm *HistoryManager) saveHistories(histories []HistoryEntry) error {
	data, err := json.MarshalIndent(histories, "", "  ")
	if err != nil {
		return err
	}

	// 写入临时文件，然后重命名（原子操作）
	tmpFile := hm.historyFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, hm.historyFile)
}
