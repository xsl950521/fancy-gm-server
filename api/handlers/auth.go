package handlers

import (
	"redis_data/api/validators"
	"redis_data/pkg/auth"
	"redis_data/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *auth.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"omitempty,email"`
	RoleID   uint   `json:"role_id"`
}

// RefreshTokenRequest 刷新token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Login 用户登录
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if !validators.ValidateJSON(c, &req) {
		return
	}

	loginReq := &auth.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	result, err := h.authService.Login(loginReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// Register 用户注册
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if !validators.ValidateJSON(c, &req) {
		return
	}

	registerReq := &auth.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		RoleID:   req.RoleID,
	}

	user, err := h.authService.Register(registerReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// RefreshToken 刷新token
// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if !validators.ValidateJSON(c, &req) {
		return
	}

	result, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetProfile 获取当前用户信息
// GET /api/auth/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// 从上下文获取用户信息（由认证中间件设置）
	user, exists := c.Get("user")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	response.Success(c, user)
}
