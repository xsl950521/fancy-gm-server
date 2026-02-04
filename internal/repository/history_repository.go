package repository

import (
	"encoding/json"
	"redis_data/internal/model"

	"gorm.io/gorm"
)

// HistoryRepository 历史记录仓储接口
type HistoryRepository interface {
	Create(entry *model.HistoryEntry) error
	GetByID(id string) (*model.HistoryEntry, error)
	GetAll(limit, offset int) ([]*model.HistoryEntry, int64, error)
	Delete(id string) error
}

type historyRepository struct {
	db *gorm.DB
}

// NewHistoryRepository 创建历史记录仓储
func NewHistoryRepository(db *gorm.DB) HistoryRepository {
	return &historyRepository{db: db}
}

// Create 创建历史记录
func (r *historyRepository) Create(entry *model.HistoryEntry) error {
	dbHistory := r.toDBModel(entry)
	if err := r.db.Create(dbHistory).Error; err != nil {
		return err
	}
	return nil
}

// GetByID 根据ID获取历史记录
func (r *historyRepository) GetByID(id string) (*model.HistoryEntry, error) {
	var dbHistory model.History
	if err := r.db.Preload("Files").Preload("Results").
		First(&dbHistory, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toModel(&dbHistory), nil
}

// GetAll 获取所有历史记录
func (r *historyRepository) GetAll(limit, offset int) ([]*model.HistoryEntry, int64, error) {
	var dbHistories []model.History
	var total int64

	query := r.db.Model(&model.History{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Files").Preload("Results").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&dbHistories).Error; err != nil {
		return nil, 0, err
	}

	entries := make([]*model.HistoryEntry, len(dbHistories))
	for i := range dbHistories {
		entries[i] = r.toModel(&dbHistories[i])
	}

	return entries, total, nil
}

// Delete 删除历史记录
func (r *historyRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.History{}).Error
}

// toDBModel 转换为数据库模型
func (r *historyRepository) toDBModel(entry *model.HistoryEntry) *model.History {
	dbHistory := &model.History{
		ID:            entry.ID,
		JobID:         entry.ID, // 使用ID作为JobID
		Mode:          entry.Mode,
		Archive:       entry.Archive,
		TotalFiles:    entry.TotalFiles,
		ProcessedFiles: entry.ProcessedFiles,
		ErrorFiles:    entry.ErrorFiles,
		Status:        entry.Status,
		Error:         entry.Error,
		ResultPath:    entry.ResultPath,
		CreatedAt:     entry.CreatedAt,
		CompletedAt:   entry.CompletedAt,
	}

	// 转换文件
	for _, f := range entry.Files {
		dbHistory.Files = append(dbHistory.Files, model.HistoryFile{
			HistoryID: entry.ID,
			Filename:  f.Filename,
			Size:      f.Size,
			Type:      f.Type,
			Path:      f.Path,
		})
	}

	// 转换结果
	for _, res := range entry.Results {
		columnsJSON, _ := json.Marshal(res.Columns)
		var dataJSON []byte
		if res.Data != nil {
			dataJSON, _ = json.Marshal(res.Data)
		}
		dbHistory.Results = append(dbHistory.Results, model.HistoryResult{
			HistoryID: entry.ID,
			SheetName: res.SheetName,
			RowCount:  res.RowCount,
			Columns:   string(columnsJSON),
			Data:      string(dataJSON),
		})
	}

	return dbHistory
}

// toModel 转换为业务模型
func (r *historyRepository) toModel(dbHistory *model.History) *model.HistoryEntry {
	entry := &model.HistoryEntry{
		ID:            dbHistory.ID,
		Mode:          dbHistory.Mode,
		Archive:       dbHistory.Archive,
		TotalFiles:    dbHistory.TotalFiles,
		ProcessedFiles: dbHistory.ProcessedFiles,
		ErrorFiles:    dbHistory.ErrorFiles,
		Status:        dbHistory.Status,
		Error:         dbHistory.Error,
		ResultPath:    dbHistory.ResultPath,
		CreatedAt:     dbHistory.CreatedAt,
		CompletedAt:   dbHistory.CompletedAt,
	}

	// 转换文件
	for _, f := range dbHistory.Files {
		entry.Files = append(entry.Files, model.FileInfo{
			Filename: f.Filename,
			Size:     f.Size,
			Type:     f.Type,
			Path:     f.Path,
		})
	}

	// 转换结果
	for _, res := range dbHistory.Results {
		var columns []string
		_ = json.Unmarshal([]byte(res.Columns), &columns)
		var data [][]interface{}
		if res.Data != "" {
			_ = json.Unmarshal([]byte(res.Data), &data)
		}
		entry.Results = append(entry.Results, model.SheetResult{
			SheetName: res.SheetName,
			RowCount:  res.RowCount,
			Columns:   columns,
			Data:      data,
		})
	}

	return entry
}
