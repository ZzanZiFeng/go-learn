# Step 1: 项目整合

## 目标

整合 Auth Service 和 Todo API 服务。

## 1.1 服务依赖关系

```
┌────────────────┐      验证 Token      ┌────────────────┐
│   Todo API     │ ────────────────────▶│  Auth Service  │
│    :8080       │                      │     :8081      │
└────────────────┘                      └────────────────┘
        │                                       │
        │                                       │
        ▼                                       ▼
┌────────────────┐                      ┌────────────────┐
│   PostgreSQL   │                      │   PostgreSQL   │
│   todo_api     │                      │  auth_service  │
└────────────────┘                      └────────────────┘
        │                                       │
        └───────────────────┬───────────────────┘
                            │
                            ▼
                    ┌────────────────┐
                    │     Redis      │
                    │   (共享缓存)   │
                    └────────────────┘
```

## 1.2 共享 Token 验证

### 方式一：共享密钥

两个服务使用相同的 JWT 密钥，Todo API 可以直接验证令牌。

```yaml
# configs/config.yaml (Todo API)
jwt:
  secret: ${JWT_SECRET}  # 与 Auth Service 相同
```

### 方式二：Token 内省

Todo API 调用 Auth Service 验证令牌。

创建 `internal/clients/auth_client.go`:

```go
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ValidateToken 验证令牌并获取用户信息
func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/users/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token validation failed: status %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
```

### 方式三：使用中间件服务

```go
// internal/middleware/auth_client.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/your-username/go-learn/projects/todo-api/internal/clients"
)

func AuthWithClient(authClient *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "未提供认证令牌",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证格式",
			})
			return
		}

		userInfo, err := authClient.ValidateToken(c.Request.Context(), parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "令牌无效或已过期",
			})
			return
		}

		c.Set("user_id", userInfo.ID)
		c.Set("username", userInfo.Username)
		c.Set("email", userInfo.Email)

		c.Next()
	}
}
```

## 1.3 服务发现配置

创建 `configs/services.yaml`:

```yaml
services:
  auth:
    url: http://auth-service:8081
    timeout: 5s

  todo:
    url: http://todo-api:8080
    timeout: 5s

  redis:
    host: redis
    port: 6379

  postgres:
    host: postgres
    port: 5432
```

## 1.4 环境变量

创建 `.env.example`:

```bash
# 数据库
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres

# Auth Service
AUTH_DB_NAME=auth_service
AUTH_PORT=8081

# Todo API
TODO_DB_NAME=todo_api
TODO_PORT=8080

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_ACCESS_SECRET=your-access-secret-key
JWT_REFRESH_SECRET=your-refresh-secret-key

# OAuth (可选)
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
```

## 1.5 健康检查端点

两个服务都添加详细的健康检查：

```go
// internal/handlers/health.go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Check(c *gin.Context) {
	status := gin.H{
		"status": "ok",
		"services": gin.H{
			"database": h.checkDatabase(),
			"redis":    h.checkRedis(c),
		},
	}

	c.JSON(http.StatusOK, status)
}

func (h *HealthHandler) checkDatabase() string {
	sqlDB, err := h.db.DB()
	if err != nil {
		return "error: " + err.Error()
	}
	if err := sqlDB.Ping(); err != nil {
		return "error: " + err.Error()
	}
	return "ok"
}

func (h *HealthHandler) checkRedis(c *gin.Context) string {
	if h.redis == nil {
		return "not configured"
	}
	if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
		return "error: " + err.Error()
	}
	return "ok"
}
```

## 检查点

完成本步骤后：

- [x] 理解服务间通信方式
- [x] 实现 Token 验证客户端
- [x] 配置服务发现
- [x] 添加健康检查

## 下一步

[Step 2: API 网关](./step-02-gateway.md) - 配置 Nginx 网关。
