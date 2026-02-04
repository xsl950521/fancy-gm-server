package models

import (
	"redis_data/pkg/errors"
	"redis_data/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// JobManager 任务管理器
type JobManager struct {
	repo JobRepository
}

// NewJobManager 创建任务管理器
func NewJobManager(repo JobRepository) *JobManager {
	jm := &JobManager{
		repo: repo,
	}
	// 启动清理协程，定期清理过期任务（24小时）
	go jm.cleanupExpiredJobs()
	return jm
}

// CreateJob 创建新任务
func (jm *JobManager) CreateJob(id, mode string, archive bool, files []FileInfo) (*Job, error) {
	job := &Job{
		ID:            id,
		Status:        JobStatusPending,
		Mode:          mode,
		Archive:       archive,
		Files:         files,
		Progress:      0,
		TotalFiles:    len(files),
		ProcessedFiles: 0,
		ErrorFiles:    0,
		Results:       []SheetResult{},
		CreatedAt:     time.Now(),
	}

	if err := jm.repo.Create(job); err != nil {
		logger.Logger.Error("Failed to create job",
			zap.String("job_id", id),
			zap.Error(err),
		)
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "创建任务失败")
	}

	return job, nil
}

// GetJob 获取任务
func (jm *JobManager) GetJob(id string) (*Job, bool) {
	job, err := jm.repo.GetByID(id)
	if err != nil {
		logger.Logger.Error("Failed to get job",
			zap.String("job_id", id),
			zap.Error(err),
		)
		return nil, false
	}
	if job == nil {
		return nil, false
	}
	return job, true
}

// UpdateJob 更新任务
func (jm *JobManager) UpdateJob(job *Job) error {
	if err := jm.repo.Update(job); err != nil {
		logger.Logger.Error("Failed to update job",
			zap.String("job_id", job.ID),
			zap.Error(err),
		)
		return errors.Wrap(err, errors.ErrCodeInternal, "更新任务失败")
	}
	return nil
}

// DeleteJob 删除任务
func (jm *JobManager) DeleteJob(id string) error {
	if err := jm.repo.Delete(id); err != nil {
		logger.Logger.Error("Failed to delete job",
			zap.String("job_id", id),
			zap.Error(err),
		)
		return errors.Wrap(err, errors.ErrCodeInternal, "删除任务失败")
	}
	return nil
}

// cleanupExpiredJobs 清理过期任务（24小时）
func (jm *JobManager) cleanupExpiredJobs() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := jm.repo.CleanupExpired(24 * time.Hour); err != nil {
			logger.Logger.Error("Failed to cleanup expired jobs", zap.Error(err))
		}
	}
}
