# Step 3: 令牌管理

## 目标

实现 Access Token 和 Refresh Token 双令牌机制。

## 3.1 令牌服务

创建 `internal/services/token.go`:

```go
package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/your-username/go-learn/projects/auth-service/internal/config"
	"github.com/your-username/go-learn/projects/auth-service/internal/models"
	"github.com/your-username/go-learn/projects/auth-service/internal/repositories"
)

var (
	ErrInvalidToken  = errors.New("无效的令牌")
	ErrExpiredToken  = errors.New("令牌已过期")
	ErrRevokedToken  = errors.New("令牌已撤销")
)

type Claims struct {
	UserID      uint     `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

type TokenService struct {
	jwtCfg    *config.JWTConfig
	tokenRepo *repositories.TokenRepository
	redis     *redis.Client
}

func NewTokenService(jwtCfg *config.JWTConfig, tokenRepo *repositories.TokenRepository, redis *redis.Client) *TokenService {
	return &TokenService{
		jwtCfg:    jwtCfg,
		tokenRepo: tokenRepo,
		redis:     redis,
	}
}

// GenerateAccessToken 生成访问令牌
func (s *TokenService) GenerateAccessToken(user *models.User) (string, error) {
	// 收集角色和权限
	var roles []string
	var permissions []string
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
		for _, perm := range role.Permissions {
			permissions = append(permissions, perm.Name)
		}
	}

	claims := &Claims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtCfg.AccessExpireDuration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.AccessSecret))
}

// ValidateAccessToken 验证访问令牌
func (s *TokenService) ValidateAccessToken(tokenString string) (*Claims, error) {
	// 检查黑名单
	if s.redis != nil {
		blacklisted, _ := s.redis.Exists(context.Background(), "blacklist:"+tokenString).Result()
		if blacklisted > 0 {
			return nil, ErrRevokedToken
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.jwtCfg.AccessSecret), nil
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

// GenerateRefreshToken 生成刷新令牌
func (s *TokenService) GenerateRefreshToken(ctx context.Context, userID uint) (string, error) {
	// 生成随机令牌
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)

	// 保存到数据库
	refreshToken := &models.RefreshToken{
		UserID:    userID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(s.jwtCfg.RefreshExpireDuration()),
	}

	if err := s.tokenRepo.Create(ctx, refreshToken); err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateRefreshToken 验证刷新令牌
func (s *TokenService) ValidateRefreshToken(ctx context.Context, tokenString string) (*models.RefreshToken, error) {
	token, err := s.tokenRepo.FindByToken(ctx, tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if token.Revoked {
		return nil, ErrRevokedToken
	}

	if token.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return token, nil
}

// RefreshTokens 刷新令牌对
func (s *TokenService) RefreshTokens(ctx context.Context, refreshTokenString string, userRepo *repositories.UserRepository) (string, string, error) {
	// 验证刷新令牌
	refreshToken, err := s.ValidateRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return "", "", err
	}

	// 获取用户信息
	user, err := userRepo.FindByID(ctx, refreshToken.UserID)
	if err != nil {
		return "", "", err
	}

	// 撤销旧的刷新令牌
	if err := s.tokenRepo.Revoke(ctx, refreshTokenString); err != nil {
		return "", "", err
	}

	// 生成新的令牌对
	newAccessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// RevokeRefreshToken 撤销刷新令牌
func (s *TokenService) RevokeRefreshToken(ctx context.Context, tokenString string) error {
	return s.tokenRepo.Revoke(ctx, tokenString)
}

// RevokeAllUserTokens 撤销用户所有令牌
func (s *TokenService) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	return s.tokenRepo.RevokeAllByUserID(ctx, userID)
}

// BlacklistAccessToken 将访问令牌加入黑名单
func (s *TokenService) BlacklistAccessToken(ctx context.Context, tokenString string, expiration time.Duration) error {
	if s.redis == nil {
		return nil
	}
	return s.redis.Set(ctx, "blacklist:"+tokenString, "1", expiration).Err()
}
```

## 3.2 令牌仓储

创建 `internal/repositories/token.go`:

```go
package repositories

import (
	"context"
	"errors"

	"github.com/your-username/go-learn/projects/auth-service/internal/models"
	"gorm.io/gorm"
)

var ErrTokenNotFound = errors.New("令牌不存在")

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *TokenRepository) FindByToken(ctx context.Context, tokenString string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	result := r.db.WithContext(ctx).Where("token = ?", tokenString).First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		return nil, result.Error
	}
	return &token, nil
}

func (r *TokenRepository) Revoke(ctx context.Context, tokenString string) error {
	return r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("token = ?", tokenString).
		Update("revoked", true).Error
}

func (r *TokenRepository) RevokeAllByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
}

func (r *TokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < NOW()").
		Delete(&models.RefreshToken{}).Error
}
```

## 3.3 刷新令牌处理器

更新 `internal/handlers/auth.go`:

```go
// 添加到 AuthHandler

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	accessToken, refreshToken, err := h.tokenService.RefreshTokens(
		c.Request.Context(),
		req.RefreshToken,
		h.userRepo,
	)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidToken):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的刷新令牌"})
		case errors.Is(err, services.ErrExpiredToken):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "刷新令牌已过期"})
		case errors.Is(err, services.ErrRevokedToken):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌已撤销"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
```

## 3.4 令牌流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                        令牌刷新流程                              │
└─────────────────────────────────────────────────────────────────┘

  客户端                    API 服务器                 数据库/Redis
    │                          │                          │
    │  1. 登录请求             │                          │
    │ ─────────────────────▶   │                          │
    │                          │  2. 验证凭据              │
    │                          │ ─────────────────────────▶│
    │                          │ ◀───────────────────────── │
    │                          │                          │
    │  3. 返回 access + refresh │                          │
    │ ◀─────────────────────   │                          │
    │                          │                          │
    │  4. API 请求 (access)    │                          │
    │ ─────────────────────▶   │                          │
    │                          │  5. 验证 access token    │
    │                          │ ─────────────────────────▶│
    │  6. 返回数据             │                          │
    │ ◀─────────────────────   │                          │
    │                          │                          │
    │  7. access token 过期    │                          │
    │                          │                          │
    │  8. 刷新请求 (refresh)   │                          │
    │ ─────────────────────▶   │                          │
    │                          │  9. 验证 refresh token   │
    │                          │ ─────────────────────────▶│
    │                          │  10. 撤销旧 refresh      │
    │                          │ ─────────────────────────▶│
    │                          │  11. 生成新令牌对        │
    │  12. 返回新令牌对        │                          │
    │ ◀─────────────────────   │                          │
    │                          │                          │
```

## 3.5 测试令牌刷新

```bash
# 登录获取令牌
RESPONSE=$(curl -s -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Password123"}')

ACCESS_TOKEN=$(echo $RESPONSE | jq -r '.access_token')
REFRESH_TOKEN=$(echo $RESPONSE | jq -r '.refresh_token')

# 使用 access token
curl http://localhost:8081/api/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 刷新令牌
curl -X POST http://localhost:8081/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"
```

## 检查点

完成本步骤后：

- [x] 实现了 Access Token 生成和验证
- [x] 实现了 Refresh Token 管理
- [x] 实现了令牌刷新机制
- [x] 实现了令牌撤销功能

## 下一步

[Step 4: RBAC 权限](./step-04-rbac.md) - 实现角色和权限管理。
