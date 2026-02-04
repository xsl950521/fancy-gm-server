package audit

import (
	"encoding/json"
	"redis_data/internal/model"
	"redis_data/internal/repository"
	"time"
)

// AuditService 审计日志服务
type AuditService struct {
	auditRepo repository.AuditRepository
}

// NewAuditService 创建审计日志服务
func NewAuditService(auditRepo repository.AuditRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// LogAction 记录操作
func (s *AuditService) LogAction(userID *uint, action, resource, resourceID, ipAddress, userAgent, requestID string, details interface{}, status string) error {
	var detailsJSON string
	if details != nil {
		detailsBytes, err := json.Marshal(details)
		if err == nil {
			detailsJSON = string(detailsBytes)
		}
	}

	log := &model.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		RequestID:  requestID,
		Details:    detailsJSON,
		Status:     status,
		CreatedAt:  time.Now(),
	}

	return s.auditRepo.Create(log)
}

// LogSuccess 记录成功操作
func (s *AuditService) LogSuccess(userID *uint, action, resource, resourceID, ipAddress, userAgent, requestID string, details interface{}) error {
	return s.LogAction(userID, action, resource, resourceID, ipAddress, userAgent, requestID, details, "success")
}

// LogFailed 记录失败操作
func (s *AuditService) LogFailed(userID *uint, action, resource, resourceID, ipAddress, userAgent, requestID string, details interface{}) error {
	return s.LogAction(userID, action, resource, resourceID, ipAddress, userAgent, requestID, details, "failed")
}

// GetLogs 获取审计日志
func (s *AuditService) GetLogs(userID *uint, action string, startTime, endTime *time.Time, offset, limit int) ([]*model.AuditLog, int64, error) {
	return s.auditRepo.List(userID, action, startTime, endTime, offset, limit)
}
