package models

import (
	"time"
)

// JobRepository 任务仓储接口（定义在api/models包中，避免循环依赖）
type JobRepository interface {
	Create(job *Job) error
	GetByID(id string) (*Job, error)
	Update(job *Job) error
	Delete(id string) error
	GetAll(limit, offset int) ([]*Job, int64, error)
	CleanupExpired(duration time.Duration) error
}
