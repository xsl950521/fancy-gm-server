package history

import (
	"redis_data/api/models"
	"redis_data/internal/model"
	"redis_data/internal/repository"
	"redis_data/pkg/errors"
	"redis_data/pkg/logger"

	"go.uber.org/zap"
)

// HistoryEntry 历史记录条目（保持向后兼容，实际使用model.HistoryEntry）
type HistoryEntry = model.HistoryEntry

// HistoryManager 历史记录管理器
type HistoryManager struct {
	repo repository.HistoryRepository
}

// NewHistoryManager 创建历史记录管理器
func NewHistoryManager(repo repository.HistoryRepository) *HistoryManager {
	return &HistoryManager{
		repo: repo,
	}
}

// SaveHistory 保存历史记录
func (hm *HistoryManager) SaveHistory(job *models.Job, resultPath string) error {
	// 转换FileInfo
	files := make([]model.FileInfo, len(job.Files))
	for i, f := range job.Files {
		files[i] = model.FileInfo{
			Filename: f.Filename,
			Size:     f.Size,
			Type:     f.Type,
			Path:     f.Path,
		}
	}

	// 转换SheetResult
	results := make([]model.SheetResult, len(job.Results))
	for i, r := range job.Results {
		results[i] = model.SheetResult{
			SheetName: r.SheetName,
			RowCount:  r.RowCount,
			Columns:   r.Columns,
			Data:      r.Data,
		}
	}

	// 创建新的历史记录条目
	entry := model.HistoryEntry{
		ID:            job.ID,
		Mode:          job.Mode,
		Archive:       job.Archive,
		Files:         files,
		TotalFiles:    job.TotalFiles,
		ProcessedFiles: job.ProcessedFiles,
		ErrorFiles:    job.ErrorFiles,
		Results:       results,
		Status:        string(job.GetStatus()),
		CreatedAt:     job.CreatedAt,
		CompletedAt:   job.CompletedAt,
		ResultPath:    resultPath,
	}

	if job.Error != "" {
		entry.Error = job.Error
	}

	if err := hm.repo.Create(&entry); err != nil {
		logger.Logger.Error("Failed to save history",
			zap.String("job_id", job.ID),
			zap.Error(err),
		)
		return errors.Wrap(err, errors.ErrCodeInternal, "保存历史记录失败")
	}

	return nil
}

// GetHistories 获取历史记录列表（分页）
func (hm *HistoryManager) GetHistories(page, pageSize int) ([]HistoryEntry, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100 // 默认每页100条
	}
	if pageSize > 1000 {
		pageSize = 1000 // 最大每页1000条
	}
	
	offset := (page - 1) * pageSize
	entries, total, err := hm.repo.GetAll(pageSize, offset)
	if err != nil {
		logger.Logger.Error("Failed to get histories", zap.Error(err))
		return nil, 0, errors.Wrap(err, errors.ErrCodeInternal, "获取历史记录失败")
	}

	result := make([]HistoryEntry, len(entries))
	for i, entry := range entries {
		result[i] = *entry
	}

	return result, total, nil
}

// GetHistory 获取单个历史记录
func (hm *HistoryManager) GetHistory(id string) (*HistoryEntry, error) {
	entry, err := hm.repo.GetByID(id)
	if err != nil {
		logger.Logger.Error("Failed to get history",
			zap.String("id", id),
			zap.Error(err),
		)
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "获取历史记录失败")
	}

	if entry == nil {
		return nil, errors.New(errors.ErrCodeNotFound, "历史记录不存在")
	}

	return entry, nil
}

// DeleteHistory 删除历史记录
func (hm *HistoryManager) DeleteHistory(id string) error {
	if err := hm.repo.Delete(id); err != nil {
		logger.Logger.Error("Failed to delete history",
			zap.String("id", id),
			zap.Error(err),
		)
		return errors.Wrap(err, errors.ErrCodeInternal, "删除历史记录失败")
	}
	return nil
}
