package auth

// 权限常量定义
const (
	// 文件操作权限
	PermissionFileUpload   = "file.upload"   // 文件上传
	PermissionFileDownload = "file.download" // 文件下载
	PermissionFileDelete   = "file.delete"   // 文件删除

	// 数据处理权限
	PermissionDataProcess = "data.process" // 数据处理
	PermissionDataView    = "data.view"    // 数据查看
	PermissionDataExport  = "data.export"  // 数据导出

	// 补发操作权限（高危操作）
	PermissionResendMail = "resend.mail" // 补发邮件

	// 历史记录权限
	PermissionHistoryView   = "history.view"   // 查看历史记录
	PermissionHistoryDelete = "history.delete" // 删除历史记录

	// 配置管理权限
	PermissionConfigView   = "config.view"   // 查看配置
	PermissionConfigModify = "config.modify" // 修改配置

	// 用户管理权限
	PermissionUserView   = "user.view"   // 查看用户
	PermissionUserCreate = "user.create" // 创建用户
	PermissionUserUpdate = "user.update" // 更新用户
	PermissionUserDelete = "user.delete" // 删除用户

	// 角色管理权限
	PermissionRoleView   = "role.view"   // 查看角色
	PermissionRoleCreate = "role.create" // 创建角色
	PermissionRoleUpdate = "role.update" // 更新角色
	PermissionRoleDelete = "role.delete" // 删除角色

	// 审计日志权限
	PermissionAuditView = "audit.view" // 查看审计日志
)

// 角色常量定义
const (
	RoleSuperAdmin = "super_admin" // 超级管理员
	RoleAdmin      = "admin"       // GM管理员
	RoleOperator   = "operator"    // 操作员
	RoleViewer     = "viewer"      // 查看者
)

// GetRolePermissions 获取角色的默认权限
func GetRolePermissions(roleName string) []string {
	switch roleName {
	case RoleSuperAdmin:
		// 超级管理员拥有所有权限
		return []string{
			PermissionFileUpload, PermissionFileDownload, PermissionFileDelete,
			PermissionDataProcess, PermissionDataView, PermissionDataExport,
			PermissionResendMail,
			PermissionHistoryView, PermissionHistoryDelete,
			PermissionConfigView, PermissionConfigModify,
			PermissionUserView, PermissionUserCreate, PermissionUserUpdate, PermissionUserDelete,
			PermissionRoleView, PermissionRoleCreate, PermissionRoleUpdate, PermissionRoleDelete,
			PermissionAuditView,
		}
	case RoleAdmin:
		// GM管理员拥有大部分权限，但不能管理用户和角色
		return []string{
			PermissionFileUpload, PermissionFileDownload,
			PermissionDataProcess, PermissionDataView, PermissionDataExport,
			PermissionResendMail,
			PermissionHistoryView, PermissionHistoryDelete,
			PermissionConfigView,
		}
	case RoleOperator:
		// 操作员可以上传、处理和查看数据，但不能删除和补发
		return []string{
			PermissionFileUpload, PermissionFileDownload,
			PermissionDataProcess, PermissionDataView, PermissionDataExport,
			PermissionHistoryView,
		}
	case RoleViewer:
		// 查看者只能查看数据
		return []string{
			PermissionFileDownload,
			PermissionDataView,
			PermissionHistoryView,
		}
	default:
		return []string{}
	}
}

// HasPermission 检查用户是否有指定权限
func HasPermission(userPermissions []string, requiredPermission string) bool {
	for _, perm := range userPermissions {
		if perm == requiredPermission {
			return true
		}
	}
	return false
}
