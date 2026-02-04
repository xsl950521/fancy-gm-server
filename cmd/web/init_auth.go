package main

import (
	"redis_data/config"
	"redis_data/internal/model"
	"redis_data/pkg/auth"
	"redis_data/pkg/logger"

	"gorm.io/gorm"
	"go.uber.org/zap"
)

// initDefaultRolesAndPermissions 初始化默认角色和权限
func initDefaultRolesAndPermissions(db *gorm.DB, appCfg *config.AppConfig) error {
	// 创建权限
	permissions := []*model.Permission{
		{Name: "文件上传", Code: auth.PermissionFileUpload, Description: "允许上传文件"},
		{Name: "文件下载", Code: auth.PermissionFileDownload, Description: "允许下载文件"},
		{Name: "文件删除", Code: auth.PermissionFileDelete, Description: "允许删除文件"},
		{Name: "数据处理", Code: auth.PermissionDataProcess, Description: "允许处理数据"},
		{Name: "数据查看", Code: auth.PermissionDataView, Description: "允许查看数据"},
		{Name: "数据导出", Code: auth.PermissionDataExport, Description: "允许导出数据"},
		{Name: "补发邮件", Code: auth.PermissionResendMail, Description: "允许补发邮件（高危操作）"},
		{Name: "查看历史", Code: auth.PermissionHistoryView, Description: "允许查看历史记录"},
		{Name: "删除历史", Code: auth.PermissionHistoryDelete, Description: "允许删除历史记录"},
		{Name: "查看配置", Code: auth.PermissionConfigView, Description: "允许查看配置"},
		{Name: "修改配置", Code: auth.PermissionConfigModify, Description: "允许修改配置"},
		{Name: "查看用户", Code: auth.PermissionUserView, Description: "允许查看用户"},
		{Name: "创建用户", Code: auth.PermissionUserCreate, Description: "允许创建用户"},
		{Name: "更新用户", Code: auth.PermissionUserUpdate, Description: "允许更新用户"},
		{Name: "删除用户", Code: auth.PermissionUserDelete, Description: "允许删除用户"},
		{Name: "查看角色", Code: auth.PermissionRoleView, Description: "允许查看角色"},
		{Name: "创建角色", Code: auth.PermissionRoleCreate, Description: "允许创建角色"},
		{Name: "更新角色", Code: auth.PermissionRoleUpdate, Description: "允许更新角色"},
		{Name: "删除角色", Code: auth.PermissionRoleDelete, Description: "允许删除角色"},
		{Name: "查看审计日志", Code: auth.PermissionAuditView, Description: "允许查看审计日志"},
	}

	for _, perm := range permissions {
		var existing model.Permission
		if err := db.Where("code = ?", perm.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(perm).Error; err != nil {
					logger.Logger.Error("Failed to create permission", zap.String("code", perm.Code), zap.Error(err))
				}
			} else {
				logger.Logger.Error("Failed to check permission", zap.String("code", perm.Code), zap.Error(err))
			}
		}
	}

	// 获取所有权限
	var allPermissions []model.Permission
	if err := db.Find(&allPermissions).Error; err != nil {
		return err
	}

	// 创建角色
	roles := []struct {
		role        *model.Role
		permissions []string
	}{
		{
			role: &model.Role{
				Name:        auth.RoleSuperAdmin,
				Description: "超级管理员，拥有所有权限",
			},
			permissions: auth.GetRolePermissions(auth.RoleSuperAdmin),
		},
		{
			role: &model.Role{
				Name:        auth.RoleAdmin,
				Description: "GM管理员，拥有大部分权限",
			},
			permissions: auth.GetRolePermissions(auth.RoleAdmin),
		},
		{
			role: &model.Role{
				Name:        auth.RoleOperator,
				Description: "操作员，可以上传和处理数据",
			},
			permissions: auth.GetRolePermissions(auth.RoleOperator),
		},
		{
			role: &model.Role{
				Name:        auth.RoleViewer,
				Description: "查看者，只能查看数据",
			},
			permissions: auth.GetRolePermissions(auth.RoleViewer),
		},
	}

	for _, roleData := range roles {
		var existing model.Role
		if err := db.Where("name = ?", roleData.role.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(roleData.role).Error; err != nil {
					logger.Logger.Error("Failed to create role", zap.String("name", roleData.role.Name), zap.Error(err))
					continue
				}
				existing = *roleData.role
			} else {
				logger.Logger.Error("Failed to check role", zap.String("name", roleData.role.Name), zap.Error(err))
				continue
			}
		}

		// 关联权限
		var rolePerms []model.Permission
		for _, permCode := range roleData.permissions {
			for _, perm := range allPermissions {
				if perm.Code == permCode {
					rolePerms = append(rolePerms, perm)
					break
				}
			}
		}
		if err := db.Model(&existing).Association("Permissions").Replace(rolePerms); err != nil {
			logger.Logger.Error("Failed to associate permissions", zap.String("role", existing.Name), zap.Error(err))
		}
	}

	// 创建默认管理员用户（如果不存在）
	var adminUser model.User
	if err := db.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 获取超级管理员角色
			var superAdminRole model.Role
			if err := db.Where("name = ?", auth.RoleSuperAdmin).First(&superAdminRole).Error; err != nil {
				logger.Logger.Error("Failed to find super admin role", zap.Error(err))
				return err
			}

			// 加密密码（默认密码：admin123）
			hashedPassword, err := auth.HashPassword("admin123")
			if err != nil {
				logger.Logger.Error("Failed to hash password", zap.Error(err))
				return err
			}

			adminUser = model.User{
				Username: "admin",
				Password: hashedPassword,
				Email:    "admin@example.com",
				RoleID:   superAdminRole.ID,
				Status:   "active",
			}

			if err := db.Create(&adminUser).Error; err != nil {
				logger.Logger.Error("Failed to create admin user", zap.Error(err))
				return err
			}

			logger.Logger.Info("Default admin user created",
				zap.String("username", "admin"),
				zap.String("password", "admin123"),
				zap.String("note", "Please change the default password after first login"),
			)
		}
	}

	return nil
}
