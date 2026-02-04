package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("无效的token")
	ErrExpiredToken  = errors.New("token已过期")
)

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey     string        // 密钥
	Expiration    time.Duration // 过期时间
	RefreshExp    time.Duration // 刷新token过期时间
	Issuer        string        // 签发者
}

// Claims JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoleID   uint   `json:"role_id"`
	RoleName string `json:"role_name"`
	jwt.RegisteredClaims
}

// RefreshClaims 刷新token声明
type RefreshClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTService JWT服务
type JWTService struct {
	config JWTConfig
}

// NewJWTService 创建JWT服务
func NewJWTService(config JWTConfig) *JWTService {
	if config.Expiration == 0 {
		config.Expiration = 24 * time.Hour // 默认24小时
	}
	if config.RefreshExp == 0 {
		config.RefreshExp = 7 * 24 * time.Hour // 默认7天
	}
	if config.Issuer == "" {
		config.Issuer = "gm-server"
	}
	return &JWTService{config: config}
}

// GenerateToken 生成访问token
func (s *JWTService) GenerateToken(userID uint, username string, roleID uint, roleName string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.Expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}

// GenerateRefreshToken 生成刷新token
func (s *JWTService) GenerateRefreshToken(userID uint) (string, error) {
	now := time.Now()
	claims := RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshExp)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey + "_refresh"))
}

// ParseToken 解析token
func (s *JWTService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// ParseRefreshToken 解析刷新token
func (s *JWTService) ParseRefreshToken(tokenString string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(s.config.SecretKey + "_refresh"), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrExpiredToken
		}
		return 0, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, ErrInvalidToken
}
