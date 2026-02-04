package middleware

import (
	"redis_data/pkg/auth"
	"redis_data/pkg/errors"
	"redis_data/pkg/logger"
	"redis_data/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	UserKey      = "user"
	UserIDKey    = "user_id"
	UsernameKey  = "username"
	RoleIDKey    = "role_id"
	RoleNameKey  = "role_name"
	PermissionsKey = "permissions"
)

// Auth 认证中间件
func Auth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未提供认证token")
			c.Abort()
			return
		}

		// 解析Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "认证token格式错误")
			c.Abort()
			return
		}

		token := parts[1]

		// 解析token
		claims, err := jwtService.ParseToken(token)
		if err != nil {
			if err == auth.ErrExpiredToken {
				response.Unauthorized(c, "token已过期")
			} else {
				response.Unauthorized(c, "无效的token")
			}
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set(UserKey, claims)
		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(RoleIDKey, claims.RoleID)
		c.Set(RoleNameKey, claims.RoleName)

		c.Next()
	}
}

// RequirePermission 权限检查中间件
func RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		roleName, exists := c.Get(RoleNameKey)
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		roleNameStr, ok := roleName.(string)
		if !ok {
			response.InternalError(c, errors.New(errors.ErrCodeInternal, "角色信息错误"))
			c.Abort()
			return
		}

		// 获取角色权限
		permissions := auth.GetRolePermissions(roleNameStr)

		// 检查权限
		if !auth.HasPermission(permissions, requiredPermission) {
			logger.Logger.Warn("Permission denied",
				zap.String("username", c.GetString(UsernameKey)),
				zap.String("role", roleNameStr),
				zap.String("required_permission", requiredPermission),
				zap.String("path", c.Request.URL.Path),
			)
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}

		// 将权限列表存储到上下文
		c.Set(PermissionsKey, permissions)

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（如果提供了token则验证，否则跳过）
func OptionalAuth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		claims, err := jwtService.ParseToken(token)
		if err != nil {
			// token无效，但不阻止请求
			c.Next()
			return
		}

		// 将用户信息存储到上下文
		c.Set(UserKey, claims)
		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(RoleIDKey, claims.RoleID)
		c.Set(RoleNameKey, claims.RoleName)

		c.Next()
	}
}
