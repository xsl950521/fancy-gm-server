package repository

import (
	"encoding/json"
	"redis_data/api/models"
	"redis_data/internal/model"
	"time"

	"gorm.io/gorm"
)

// 注意：JobRepository 接口定义在 api/models 包中，避免循环依赖
// 此包实现该接口

type jobRepository struct {
	db *gorm.DB
}

// NewJobRepository 创建任务仓储
// 返回 api/models.JobRepository 接口
func NewJobRepository(db *gorm.DB) models.JobRepository {
	return &jobRepository{db: db}
}

// Create 创建任务
func (r *jobRepository) Create(job *models.Job) error {
	dbJob := r.toDBModel(job)
	if err := r.db.Create(dbJob).Error; err != nil {
		return err
	}
	return nil
}

// GetByID 根据ID获取任务
func (r *jobRepository) GetByID(id string) (*models.Job, error) {
	var dbJob model.Job
	if err := r.db.Preload("Files").Preload("Results").First(&dbJob, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toModel(&dbJob), nil
}

// Update 更新任务
func (r *jobRepository) Update(job *models.Job) error {
	dbJob := r.toDBModel(job)
	
	// 使用事务更新关联数据
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 更新主表
		if err := tx.Model(&model.Job{}).Where("id = ?", job.ID).Updates(dbJob).Error; err != nil {
			return err
		}

		// 删除旧的关联数据
		if err := tx.Where("job_id = ?", job.ID).Delete(&model.JobFile{}).Error; err != nil {
			return err
		}
		if err := tx.Where("job_id = ?", job.ID).Delete(&model.JobResult{}).Error; err != nil {
			return err
		}

		// 插入新的关联数据
		if len(dbJob.Files) > 0 {
			if err := tx.Create(&dbJob.Files).Error; err != nil {
				return err
			}
		}
		if len(dbJob.Results) > 0 {
			if err := tx.Create(&dbJob.Results).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete 删除任务
func (r *jobRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Job{}).Error
}

// GetAll 获取所有任务
func (r *jobRepository) GetAll(limit, offset int) ([]*models.Job, int64, error) {
	var dbJobs []model.Job
	var total int64

	query := r.db.Model(&model.Job{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Files").Preload("Results").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&dbJobs).Error; err != nil {
		return nil, 0, err
	}

	jobs := make([]*models.Job, len(dbJobs))
	for i := range dbJobs {
		jobs[i] = r.toModel(&dbJobs[i])
	}

	return jobs, total, nil
}

// CleanupExpired 清理过期任务
func (r *jobRepository) CleanupExpired(duration time.Duration) error {
	cutoff := time.Now().Add(-duration)
	return r.db.Where("completed_at IS NOT NULL AND completed_at < ?", cutoff).
		Delete(&model.Job{}).Error
}

// toDBModel 转换为数据库模型
func (r *jobRepository) toDBModel(job *models.Job) *model.Job {
	dbJob := &model.Job{
		ID:            job.ID,
		Status:        string(job.Status),
		Mode:          job.Mode,
		Archive:       job.Archive,
		Progress:      job.Progress,
		TotalFiles:    job.TotalFiles,
		ProcessedFiles: job.ProcessedFiles,
		ErrorFiles:    job.ErrorFiles,
		Error:         job.Error,
		CreatedAt:     job.CreatedAt,
		CompletedAt:   job.CompletedAt,
	}

	// 转换文件
	for _, f := range job.Files {
		dbJob.Files = append(dbJob.Files, model.JobFile{
			JobID:    job.ID,
			Filename: f.Filename,
			Size:     f.Size,
			Type:     f.Type,
			Path:     f.Path,
		})
	}

	// 转换结果
	for _, res := range job.Results {
		columnsJSON, _ := json.Marshal(res.Columns)
		var dataJSON []byte
		if res.Data != nil {
			dataJSON, _ = json.Marshal(res.Data)
		}
		dbJob.Results = append(dbJob.Results, model.JobResult{
			JobID:     job.ID,
			SheetName: res.SheetName,
			RowCount:  res.RowCount,
			Columns:   string(columnsJSON),
			Data:      string(dataJSON),
		})
	}

	return dbJob
}

// toModel 转换为业务模型
func (r *jobRepository) toModel(dbJob *model.Job) *models.Job {
	job := &models.Job{
		ID:            dbJob.ID,
		Status:        models.JobStatus(dbJob.Status),
		Mode:          dbJob.Mode,
		Archive:       dbJob.Archive,
		Progress:      dbJob.Progress,
		TotalFiles:    dbJob.TotalFiles,
		ProcessedFiles: dbJob.ProcessedFiles,
		ErrorFiles:    dbJob.ErrorFiles,
		Error:         dbJob.Error,
		CreatedAt:     dbJob.CreatedAt,
		CompletedAt:   dbJob.CompletedAt,
	}

	// 转换文件
	for _, f := range dbJob.Files {
		job.Files = append(job.Files, models.FileInfo{
			Filename: f.Filename,
			Size:     f.Size,
			Type:     f.Type,
			Path:     f.Path,
		})
	}

	// 转换结果
	for _, res := range dbJob.Results {
		var columns []string
		_ = json.Unmarshal([]byte(res.Columns), &columns)
		var data [][]interface{}
		if res.Data != "" {
			_ = json.Unmarshal([]byte(res.Data), &data)
		}
		job.Results = append(job.Results, models.SheetResult{
			SheetName: res.SheetName,
			RowCount:  res.RowCount,
			Columns:   columns,
			Data:      data,
		})
	}

	return job
}
