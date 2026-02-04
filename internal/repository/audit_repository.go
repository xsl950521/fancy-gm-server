package repository

import (
	"redis_data/internal/model"
	"time"

	"gorm.io/gorm"
)

// AuditRepository 审计日志仓储接口
type AuditRepository interface {
	Create(log *model.AuditLog) error
	GetByID(id uint) (*model.AuditLog, error)
	List(userID *uint, action string, startTime, endTime *time.Time, offset, limit int) ([]*model.AuditLog, int64, error)
	Delete(id uint) error
	DeleteOld(before time.Time) error
}

// auditRepository 审计日志仓储实现
type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository 创建审计日志仓储
func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

// Create 创建审计日志
func (r *auditRepository) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

// GetByID 根据ID获取审计日志
func (r *auditRepository) GetByID(id uint) (*model.AuditLog, error) {
	var log model.AuditLog
	err := r.db.Preload("User").First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// List 列出审计日志
func (r *auditRepository) List(userID *uint, action string, startTime, endTime *time.Time, offset, limit int) ([]*model.AuditLog, int64, error) {
	var logs []*model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

// Delete 删除审计日志
func (r *auditRepository) Delete(id uint) error {
	return r.db.Delete(&model.AuditLog{}, id).Error
}

// DeleteOld 删除旧日志
func (r *auditRepository) DeleteOld(before time.Time) error {
	return r.db.Where("created_at < ?", before).Delete(&model.AuditLog{}).Error
}
