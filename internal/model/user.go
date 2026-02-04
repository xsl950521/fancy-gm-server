package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Password  string         `gorm:"not null;size:255" json:"-"` // 不序列化密码
	Email     string         `gorm:"size:100" json:"email"`
	RoleID    uint           `gorm:"not null;default:3" json:"role_id"` // 默认角色：操作员
	Role      Role           `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Status    string         `gorm:"default:active;size:20" json:"status"` // active, inactive, locked
	LastLogin *time.Time     `json:"last_login,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Role 角色模型
type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null;size:50" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	Permissions []Permission   `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Permission 权限模型
type Permission struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Code        string         `gorm:"uniqueIndex;not null;size:100" json:"code"` // 权限代码，如：file.upload
	Description string         `gorm:"size:255" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// AuditLog 操作审计日志
type AuditLog struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      *uint          `json:"user_id,omitempty"`
	User        *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action      string         `gorm:"not null;size:100" json:"action"`      // 操作类型
	Resource    string         `gorm:"size:100" json:"resource"`             // 资源类型
	ResourceID  string         `gorm:"size:100" json:"resource_id,omitempty"` // 资源ID
	IPAddress   string         `gorm:"size:45" json:"ip_address"`
	UserAgent   string         `gorm:"size:500" json:"user_agent"`
	RequestID   string         `gorm:"size:100" json:"request_id"`
	Details     string         `gorm:"type:text" json:"details,omitempty"` // JSON格式的详细信息
	Status      string         `gorm:"size:20" json:"status"`              // success, failed
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
