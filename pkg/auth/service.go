package auth

import (
	"errors"
	"redis_data/internal/model"
	"redis_data/internal/repository"
	pkgerrors "redis_data/pkg/errors"
	"time"

	"gorm.io/gorm"
)


// AuthService 认证服务
type AuthService struct {
	userRepo  repository.UserRepository
	jwtService *JWTService
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repository.UserRepository, jwtService *JWTService) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtService: jwtService,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *UserInfo `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleID   uint   `json:"role_id"`
	RoleName string `json:"role_name"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"omitempty,email"`
	RoleID   uint   `json:"role_id"` // 可选，默认操作员
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.New(pkgerrors.ErrCodeUnauthorized, "用户名或密码错误")
		}
		return nil, pkgerrors.Wrap(err, pkgerrors.ErrCodeInternal, "获取用户失败")
	}

	// 检查用户状态
	if user.Status == "locked" {
		return nil, pkgerrors.New(pkgerrors.ErrCodeForbidden, "用户已被锁定")
	}
	if user.Status == "inactive" {
		return nil, pkgerrors.New(pkgerrors.ErrCodeForbidden, "用户未激活")
	}

	// 验证密码
	if !CheckPassword(req.Password, user.Password) {
		return nil, pkgerrors.New(pkgerrors.ErrCodeUnauthorized, "用户名或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	if err := s.userRepo.Update(user); err != nil {
		// 记录错误但不影响登录
	}

	// 生成token
	token, err := s.jwtService.GenerateToken(user.ID, user.Username, user.RoleID, user.Role.Name)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// 计算过期时间
	expiresAt := time.Now().Add(s.jwtService.config.Expiration)

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			RoleID:   user.RoleID,
			RoleName: user.Role.Name,
		},
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest) (*UserInfo, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		return nil, pkgerrors.New(pkgerrors.ErrCodeConflict, "用户名已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, pkgerrors.Wrap(err, pkgerrors.ErrCodeInternal, "检查用户名失败")
	}

	// 检查邮箱是否已存在（如果提供了邮箱）
	if req.Email != "" {
		if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
			return nil, pkgerrors.New(pkgerrors.ErrCodeConflict, "邮箱已存在")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.Wrap(err, pkgerrors.ErrCodeInternal, "检查邮箱失败")
		}
	}

	// 加密密码
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 设置默认角色（操作员）
	roleID := req.RoleID
	if roleID == 0 {
		roleID = 3 // 默认操作员角色
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		RoleID:   roleID,
		Status:   "active",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, pkgerrors.Wrap(err, pkgerrors.ErrCodeInternal, "创建用户失败")
	}

	// 重新获取用户（包含角色信息）
	user, err = s.userRepo.GetByID(user.ID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, pkgerrors.ErrCodeInternal, "获取用户信息失败")
	}

	return &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RoleID:   user.RoleID,
		RoleName: user.Role.Name,
	}, nil
}

// RefreshToken 刷新token
func (s *AuthService) RefreshToken(refreshToken string) (*LoginResponse, error) {
	// 解析刷新token
	userID, err := s.jwtService.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 获取用户
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 检查用户状态
	if user.Status == "locked" {
		return nil, pkgerrors.New(pkgerrors.ErrCodeForbidden, "用户已被锁定")
	}
	if user.Status == "inactive" {
		return nil, pkgerrors.New(pkgerrors.ErrCodeForbidden, "用户未激活")
	}

	// 生成新token
	token, err := s.jwtService.GenerateToken(user.ID, user.Username, user.RoleID, user.Role.Name)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(s.jwtService.config.Expiration)

	return &LoginResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			RoleID:   user.RoleID,
			RoleName: user.Role.Name,
		},
	}, nil
}
