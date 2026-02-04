package models

import (
	"sync"
	"time"
)

// JobManager 任务管理器
type JobManager struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

// NewJobManager 创建任务管理器
func NewJobManager() *JobManager {
	jm := &JobManager{
		jobs: make(map[string]*Job),
	}
	// 启动清理协程，定期清理过期任务（24小时）
	go jm.cleanupExpiredJobs()
	return jm
}

// CreateJob 创建新任务
func (jm *JobManager) CreateJob(id, mode string, archive bool, files []FileInfo) *Job {
	jm.mu.Lock()
	defer jm.mu.Unlock()

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

	jm.jobs[id] = job
	return job
}

// GetJob 获取任务
func (jm *JobManager) GetJob(id string) (*Job, bool) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	job, exists := jm.jobs[id]
	return job, exists
}

// DeleteJob 删除任务
func (jm *JobManager) DeleteJob(id string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	delete(jm.jobs, id)
}

// cleanupExpiredJobs 清理过期任务（24小时）
func (jm *JobManager) cleanupExpiredJobs() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		jm.mu.Lock()
		now := time.Now()
		for id, job := range jm.jobs {
			// 删除24小时前完成的任务
			if job.CompletedAt != nil && now.Sub(*job.CompletedAt) > 24*time.Hour {
				delete(jm.jobs, id)
			}
		}
		jm.mu.Unlock()
	}
}
