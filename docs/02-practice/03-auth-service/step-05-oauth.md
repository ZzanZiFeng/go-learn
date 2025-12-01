# Step 5: OAuth2 集成

## 目标

实现 GitHub 和 Google 第三方登录。

## 5.1 OAuth 服务

创建 `internal/services/oauth.go`:

```go
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/your-username/go-learn/projects/auth-service/internal/config"
	"github.com/your-username/go-learn/projects/auth-service/internal/models"
	"github.com/your-username/go-learn/projects/auth-service/internal/repositories"
)

var (
	ErrOAuthFailed       = errors.New("OAuth 认证失败")
	ErrProviderNotSupported = errors.New("不支持的 OAuth 提供商")
)

type OAuthService struct {
	cfg          *config.OAuthConfig
	userRepo     *repositories.UserRepository
	tokenService *TokenService
}

func NewOAuthService(cfg *config.OAuthConfig, userRepo *repositories.UserRepository, tokenService *TokenService) *OAuthService {
	return &OAuthService{
		cfg:          cfg,
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

type OAuthUserInfo struct {
	ID        string
	Email     string
	Username  string
	AvatarURL string
}

// GetAuthURL 获取 OAuth 授权 URL
func (s *OAuthService) GetAuthURL(provider, state string) (string, error) {
	switch provider {
	case "github":
		return s.getGitHubAuthURL(state), nil
	case "google":
		return s.getGoogleAuthURL(state), nil
	default:
		return "", ErrProviderNotSupported
	}
}

func (s *OAuthService) getGitHubAuthURL(state string) string {
	params := url.Values{
		"client_id":    {s.cfg.GitHub.ClientID},
		"redirect_uri": {s.cfg.GitHub.RedirectURL},
		"scope":        {"user:email"},
		"state":        {state},
	}
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

func (s *OAuthService) getGoogleAuthURL(state string) string {
	params := url.Values{
		"client_id":     {s.cfg.Google.ClientID},
		"redirect_uri":  {s.cfg.Google.RedirectURL},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

// HandleCallback 处理 OAuth 回调
func (s *OAuthService) HandleCallback(ctx context.Context, provider, code string) (*AuthResult, error) {
	var userInfo *OAuthUserInfo
	var err error

	switch provider {
	case "github":
		userInfo, err = s.handleGitHubCallback(code)
	case "google":
		userInfo, err = s.handleGoogleCallback(code)
	default:
		return nil, ErrProviderNotSupported
	}

	if err != nil {
		return nil, err
	}

	// 查找或创建用户
	user, err := s.findOrCreateUser(ctx, provider, userInfo)
	if err != nil {
		return nil, err
	}

	// 生成令牌
	accessToken, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *OAuthService) handleGitHubCallback(code string) (*OAuthUserInfo, error) {
	// 获取 access token
	tokenURL := "https://github.com/login/oauth/access_token"
	data := url.Values{
		"client_id":     {s.cfg.GitHub.ClientID},
		"client_secret": {s.cfg.GitHub.ClientSecret},
		"code":          {code},
		"redirect_uri":  {s.cfg.GitHub.RedirectURL},
	}

	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("github oauth error: %s", tokenResp.Error)
	}

	// 获取用户信息
	userReq, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&githubUser); err != nil {
		return nil, err
	}

	// 如果没有公开邮箱，获取私有邮箱
	if githubUser.Email == "" {
		emailReq, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		emailReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

		emailResp, err := http.DefaultClient.Do(emailReq)
		if err == nil {
			defer emailResp.Body.Close()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if json.NewDecoder(emailResp.Body).Decode(&emails) == nil {
				for _, e := range emails {
					if e.Primary {
						githubUser.Email = e.Email
						break
					}
				}
			}
		}
	}

	return &OAuthUserInfo{
		ID:        fmt.Sprintf("%d", githubUser.ID),
		Email:     githubUser.Email,
		Username:  githubUser.Login,
		AvatarURL: githubUser.AvatarURL,
	}, nil
}

func (s *OAuthService) handleGoogleCallback(code string) (*OAuthUserInfo, error) {
	// 获取 access token
	tokenURL := "https://oauth2.googleapis.com/token"
	data := url.Values{
		"client_id":     {s.cfg.Google.ClientID},
		"client_secret": {s.cfg.Google.ClientSecret},
		"code":          {code},
		"redirect_uri":  {s.cfg.Google.RedirectURL},
		"grant_type":    {"authorization_code"},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("google oauth error: %s", tokenResp.Error)
	}

	// 获取用户信息
	userReq, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()

	body, _ := io.ReadAll(userResp.Body)

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:        googleUser.ID,
		Email:     googleUser.Email,
		Username:  googleUser.Name,
		AvatarURL: googleUser.Picture,
	}, nil
}

func (s *OAuthService) findOrCreateUser(ctx context.Context, provider string, info *OAuthUserInfo) (*models.User, error) {
	// 尝试通过 provider 查找用户
	user, err := s.userRepo.FindByProvider(ctx, provider, info.ID)
	if err == nil {
		return user, nil
	}

	// 尝试通过邮箱查找用户
	if info.Email != "" {
		user, err = s.userRepo.FindByEmail(ctx, info.Email)
		if err == nil {
			// 更新 provider 信息
			user.Provider = provider
			user.ProviderID = info.ID
			user.AvatarURL = info.AvatarURL
			s.userRepo.Update(ctx, user)
			return user, nil
		}
	}

	// 创建新用户
	user = &models.User{
		Username:      info.Username,
		Email:         info.Email,
		Provider:      provider,
		ProviderID:    info.ID,
		AvatarURL:     info.AvatarURL,
		EmailVerified: true, // OAuth 用户邮箱已验证
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 分配默认角色
	s.userRepo.AssignRole(ctx, user.ID, 2)

	return user, nil
}
```

## 5.2 OAuth 处理器

创建 `internal/handlers/oauth.go`:

```go
package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-username/go-learn/projects/auth-service/internal/services"
)

type OAuthHandler struct {
	oauthService *services.OAuthService
}

func NewOAuthHandler(oauthService *services.OAuthService) *OAuthHandler {
	return &OAuthHandler{oauthService: oauthService}
}

// Authorize 发起 OAuth 授权
func (h *OAuthHandler) Authorize(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 provider 参数"})
		return
	}

	// 生成随机 state
	stateBytes := make([]byte, 16)
	rand.Read(stateBytes)
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// 存储 state 到 session 或 cookie
	c.SetCookie("oauth_state", state, 300, "/", "", false, true)

	authURL, err := h.oauthService.GetAuthURL(provider, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// Callback 处理 OAuth 回调
func (h *OAuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	// 验证 state
	savedState, err := c.Cookie("oauth_state")
	if err != nil || savedState != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 state"})
		return
	}

	// 清除 state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	result, err := h.oauthService.HandleCallback(c.Request.Context(), provider, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 返回令牌或重定向到前端
	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: UserResponse{
			ID:       result.User.ID,
			Username: result.User.Username,
			Email:    result.User.Email,
		},
	})
}
```

## 5.3 OAuth 流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                     OAuth2 授权码流程                            │
└─────────────────────────────────────────────────────────────────┘

  用户           前端          Auth Service        OAuth Provider
   │              │                 │                    │
   │ 1. 点击登录  │                 │                    │
   │ ────────────▶│                 │                    │
   │              │ 2. 请求授权 URL │                    │
   │              │ ───────────────▶│                    │
   │              │ 3. 返回授权 URL │                    │
   │              │ ◀─────────────── │                    │
   │              │                 │                    │
   │ 4. 重定向到 OAuth              │                    │
   │ ◀────────────│                 │                    │
   │ ─────────────────────────────────────────────────▶  │
   │              │                 │  5. 用户授权       │
   │ ◀───────────────────────────────────────────────── │
   │              │                 │                    │
   │ 6. 携带 code 回调              │                    │
   │ ─────────────────────────────▶ │                    │
   │              │                 │ 7. 用 code 换 token│
   │              │                 │ ──────────────────▶│
   │              │                 │ ◀───────────────── │
   │              │                 │ 8. 获取用户信息    │
   │              │                 │ ──────────────────▶│
   │              │                 │ ◀───────────────── │
   │              │                 │                    │
   │ 9. 返回 JWT  │                 │                    │
   │ ◀───────────────────────────── │                    │
   │              │                 │                    │
```

## 5.4 路由配置

```go
// 在 main.go 中配置路由

// OAuth 路由
oauth := api.Group("/oauth")
{
    oauth.GET("/authorize", oauthHandler.Authorize)
    oauth.GET("/callback/:provider", oauthHandler.Callback)
}
```

## 5.5 测试 OAuth

```bash
# 发起 GitHub 登录
curl "http://localhost:8081/api/oauth/authorize?provider=github"
# 浏览器会重定向到 GitHub 授权页面

# 发起 Google 登录
curl "http://localhost:8081/api/oauth/authorize?provider=google"
```

## 项目总结

完成 Auth Service 项目后，你已经掌握：

| 技能 | 说明 |
|------|------|
| 密码加密 | bcrypt 哈希、密码强度验证 |
| JWT 双令牌 | Access Token + Refresh Token |
| RBAC | 角色、权限、中间件 |
| OAuth2 | GitHub、Google 第三方登录 |
| 安全最佳实践 | 令牌撤销、黑名单机制 |

## 下一步

恭喜完成 Auth Service 项目！继续：

**[项目 4: Fullstack Demo](../04-fullstack-demo/README.md)** - 整合所有项目！
