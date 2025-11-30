# OAuth2 (Third-party Login)

## 概述

OAuth2 是一种授权框架，允许用户授权第三方应用访问其在其他服务上的资源，常用于第三方登录。

## OAuth2 流程

### Authorization Code Flow

```
用户 ──> 应用 ──> 授权服务器 ──> 用户授权 ──> 回调应用
              │                              │
              └── 用 code 换取 access_token ──┘
              │                              │
              └── 用 token 获取用户信息 ──────┘
```

### 详细步骤

```
1. 用户点击 "使用 GitHub 登录"
2. 重定向到: https://github.com/login/oauth/authorize?
     client_id=xxx&
     redirect_uri=http://localhost:8080/callback&
     scope=user:email&
     state=random_state

3. 用户在 GitHub 授权

4. GitHub 回调: http://localhost:8080/callback?
     code=xxx&
     state=random_state

5. 应用用 code 换取 access_token:
   POST https://github.com/login/oauth/access_token
   Body: { client_id, client_secret, code }

6. 用 access_token 获取用户信息:
   GET https://api.github.com/user
   Authorization: Bearer {access_token}
```

## 安装

```bash
go get -u golang.org/x/oauth2
go get -u golang.org/x/oauth2/github
go get -u golang.org/x/oauth2/google
```

## GitHub OAuth2

### 配置

```go
package auth

import (
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/github"
)

type OAuthConfig struct {
    GitHub *oauth2.Config
    Google *oauth2.Config
}

func NewOAuthConfig() *OAuthConfig {
    return &OAuthConfig{
        GitHub: &oauth2.Config{
            ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
            ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
            RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
            Scopes:       []string{"user:email", "read:user"},
            Endpoint:     github.Endpoint,
        },
    }
}
```

### 处理器

```go
package handlers

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "net/http"

    "myapp/internal/auth"
    "myapp/internal/services"

    "github.com/gin-gonic/gin"
    "golang.org/x/oauth2"
)

type OAuthHandler struct {
    config      *auth.OAuthConfig
    userService *services.UserService
    jwtService  *auth.JWTService
}

func NewOAuthHandler(config *auth.OAuthConfig, userService *services.UserService, jwtService *auth.JWTService) *OAuthHandler {
    return &OAuthHandler{
        config:      config,
        userService: userService,
        jwtService:  jwtService,
    }
}

// GitHubLogin 开始 GitHub 登录
func (h *OAuthHandler) GitHubLogin(c *gin.Context) {
    // 生成随机 state 防止 CSRF
    state := generateState()

    // 存储 state（可以用 Redis 或 Session）
    c.SetCookie("oauth_state", state, 600, "/", "", true, true)

    // 重定向到 GitHub
    url := h.config.GitHub.AuthCodeURL(state, oauth2.AccessTypeOffline)
    c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback 处理 GitHub 回调
func (h *OAuthHandler) GitHubCallback(c *gin.Context) {
    // 验证 state
    state := c.Query("state")
    storedState, _ := c.Cookie("oauth_state")
    if state != storedState {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
        return
    }
    c.SetCookie("oauth_state", "", -1, "/", "", true, true)

    // 获取 code
    code := c.Query("code")
    if code == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
        return
    }

    // 用 code 换取 token
    token, err := h.config.GitHub.Exchange(context.Background(), code)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
        return
    }

    // 获取用户信息
    githubUser, err := h.getGitHubUser(token.AccessToken)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
        return
    }

    // 查找或创建用户
    user, err := h.userService.FindOrCreateByOAuth("github", githubUser.ID, githubUser.Email, githubUser.Name)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    // 生成 JWT
    tokens, err := h.jwtService.GenerateTokenPair(user.ID, user.Email, user.Role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    // 重定向到前端（携带 token）
    // 或直接返回 JSON
    c.JSON(http.StatusOK, gin.H{
        "user":   user,
        "tokens": tokens,
    })
}

type GitHubUser struct {
    ID    int64  `json:"id"`
    Login string `json:"login"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (h *OAuthHandler) getGitHubUser(accessToken string) (*GitHubUser, error) {
    req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
    req.Header.Set("Authorization", "Bearer "+accessToken)
    req.Header.Set("Accept", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var user GitHubUser
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }

    // 如果没有公开邮箱，获取私有邮箱
    if user.Email == "" {
        user.Email, _ = h.getGitHubEmail(accessToken)
    }

    return &user, nil
}

func (h *OAuthHandler) getGitHubEmail(accessToken string) (string, error) {
    req, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
    req.Header.Set("Authorization", "Bearer "+accessToken)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var emails []struct {
        Email   string `json:"email"`
        Primary bool   `json:"primary"`
    }
    json.NewDecoder(resp.Body).Decode(&emails)

    for _, e := range emails {
        if e.Primary {
            return e.Email, nil
        }
    }
    return "", nil
}

func generateState() string {
    b := make([]byte, 16)
    rand.Read(b)
    return hex.EncodeToString(b)
}
```

## Google OAuth2

```go
import (
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

func NewGoogleOAuthConfig() *oauth2.Config {
    return &oauth2.Config{
        ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
        Scopes: []string{
            "https://www.googleapis.com/auth/userinfo.email",
            "https://www.googleapis.com/auth/userinfo.profile",
        },
        Endpoint: google.Endpoint,
    }
}

type GoogleUser struct {
    ID            string `json:"id"`
    Email         string `json:"email"`
    VerifiedEmail bool   `json:"verified_email"`
    Name          string `json:"name"`
    Picture       string `json:"picture"`
}

func (h *OAuthHandler) getGoogleUser(accessToken string) (*GoogleUser, error) {
    resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var user GoogleUser
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }

    return &user, nil
}
```

## 用户服务

```go
package services

type OAuthAccount struct {
    ID       uint   `json:"id"`
    UserID   uint   `json:"user_id"`
    Provider string `json:"provider"` // github, google
    OAuthID  string `json:"oauth_id"`
}

type UserService struct {
    userRepo  *repositories.UserRepository
    oauthRepo *repositories.OAuthAccountRepository
}

// FindOrCreateByOAuth 通过 OAuth 查找或创建用户
func (s *UserService) FindOrCreateByOAuth(provider string, oauthID int64, email, name string) (*User, error) {
    // 先尝试通过 OAuth 账号查找
    oauthAccount, err := s.oauthRepo.FindByProviderAndID(provider, fmt.Sprint(oauthID))
    if err == nil {
        return s.userRepo.FindByID(oauthAccount.UserID)
    }

    // 尝试通过邮箱查找现有用户
    user, err := s.userRepo.FindByEmail(email)
    if err != nil {
        // 创建新用户
        user = &User{
            Name:  name,
            Email: email,
            Role:  "user",
        }
        if err := s.userRepo.Create(user); err != nil {
            return nil, err
        }
    }

    // 关联 OAuth 账号
    oauthAccount = &OAuthAccount{
        UserID:   user.ID,
        Provider: provider,
        OAuthID:  fmt.Sprint(oauthID),
    }
    s.oauthRepo.Create(oauthAccount)

    return user, nil
}
```

## 路由配置

```go
func SetupOAuthRoutes(r *gin.Engine, oauthHandler *OAuthHandler) {
    oauth := r.Group("/oauth")
    {
        // GitHub
        oauth.GET("/github", oauthHandler.GitHubLogin)
        oauth.GET("/github/callback", oauthHandler.GitHubCallback)

        // Google
        oauth.GET("/google", oauthHandler.GoogleLogin)
        oauth.GET("/google/callback", oauthHandler.GoogleCallback)
    }
}
```

## 前端集成

```html
<!-- 登录页面 -->
<a href="/oauth/github" class="btn btn-github">
    使用 GitHub 登录
</a>

<a href="/oauth/google" class="btn btn-google">
    使用 Google 登录
</a>
```

```javascript
// 处理回调（如果使用弹窗方式）
window.addEventListener('message', (event) => {
    if (event.data.type === 'oauth_success') {
        const { token, user } = event.data;
        localStorage.setItem('token', token);
        // 更新状态...
    }
});
```

## 安全考虑

### State 参数

```go
// 防止 CSRF 攻击
// 1. 生成随机 state
state := generateState()

// 2. 存储 state（Redis/Session）
redis.Set(ctx, "oauth_state:"+state, "1", 10*time.Minute)

// 3. 发送请求
url := config.AuthCodeURL(state)

// 4. 回调时验证
if !redis.Exists(ctx, "oauth_state:"+state) {
    return errors.New("invalid state")
}
redis.Del(ctx, "oauth_state:"+state)
```

### Token 安全

```go
// 不要在 URL 中传递敏感 token
// ❌ 错误
c.Redirect(302, "http://frontend.com?token="+accessToken)

// ✅ 正确 - 使用短期 code
code := generateOneTimeCode(user.ID)
c.Redirect(302, "http://frontend.com?code="+code)

// 前端用 code 换取 token
// POST /oauth/exchange { code: "xxx" }
```

## 完整示例

```go
package main

import (
    "myapp/internal/auth"
    "myapp/internal/handlers"
    "myapp/internal/services"

    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // 初始化
    oauthConfig := auth.NewOAuthConfig()
    userService := services.NewUserService()
    jwtService := auth.NewJWTService(os.Getenv("JWT_SECRET"))
    oauthHandler := handlers.NewOAuthHandler(oauthConfig, userService, jwtService)

    // OAuth 路由
    oauth := r.Group("/oauth")
    {
        oauth.GET("/github", oauthHandler.GitHubLogin)
        oauth.GET("/github/callback", oauthHandler.GitHubCallback)
        oauth.GET("/google", oauthHandler.GoogleLogin)
        oauth.GET("/google/callback", oauthHandler.GoogleCallback)
    }

    r.Run(":8080")
}
```

## 总结

| 提供商 | Endpoint | Scopes |
|--------|----------|--------|
| GitHub | github.Endpoint | user:email, read:user |
| Google | google.Endpoint | userinfo.email, userinfo.profile |
| Facebook | facebook.Endpoint | email, public_profile |

**下一节**：[RBAC](./06-rbac.md) - 学习角色权限控制
