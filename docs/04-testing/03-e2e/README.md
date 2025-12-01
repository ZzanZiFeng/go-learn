# E2E 测试

## 学习目标

掌握端到端测试，验证完整用户流程。

## 1. E2E 测试概述

### 1.1 什么是 E2E 测试

E2E（End-to-End）测试验证完整的用户场景，从前端到后端到数据库的整个流程。

```
┌─────────────────────────────────────────────────────────────────┐
│                      E2E 测试范围                                │
└─────────────────────────────────────────────────────────────────┘

 用户                     API                     数据库
   │                       │                        │
   │ 1. 注册               │                        │
   │ ─────────────────────▶│                        │
   │                       │ ──────────────────────▶│
   │◀───────────────────── │◀────────────────────── │
   │                       │                        │
   │ 2. 登录               │                        │
   │ ─────────────────────▶│                        │
   │                       │ ──────────────────────▶│
   │◀───────────────────── │◀────────────────────── │
   │   (返回 Token)        │                        │
   │                       │                        │
   │ 3. 创建待办           │                        │
   │ ─────────────────────▶│                        │
   │   (带 Token)          │ ──────────────────────▶│
   │◀───────────────────── │◀────────────────────── │
   │                       │                        │
```

### 1.2 E2E 测试特点

| 特点 | 说明 |
|------|------|
| 真实环境 | 使用真实数据库和服务 |
| 完整流程 | 测试用户完整操作路径 |
| 运行慢 | 需要启动完整环境 |
| 维护成本高 | 环境依赖多 |

## 2. 测试环境设置

### 2.1 Docker Compose 测试环境

```yaml
# docker-compose.test.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=test
      - DB_PASSWORD=test
      - DB_NAME=testdb
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=test-secret
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: testdb
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U test"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
```

### 2.2 测试环境启动脚本

```bash
#!/bin/bash
# scripts/e2e-test.sh

set -e

echo "Starting test environment..."
docker-compose -f docker-compose.test.yml up -d --build

echo "Waiting for services..."
sleep 10

echo "Running E2E tests..."
go test -tags=e2e -v ./tests/e2e/...

echo "Cleaning up..."
docker-compose -f docker-compose.test.yml down -v
```

## 3. 用户流程测试

### 3.1 测试客户端

```go
// tests/e2e/client.go
package e2e

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type TestClient struct {
    baseURL    string
    httpClient *http.Client
    token      string
}

func NewTestClient(baseURL string) *TestClient {
    return &TestClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (c *TestClient) SetToken(token string) {
    c.token = token
}

func (c *TestClient) Post(path string, body interface{}) (*http.Response, error) {
    jsonBody, err := json.Marshal(body)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    if c.token != "" {
        req.Header.Set("Authorization", "Bearer "+c.token)
    }

    return c.httpClient.Do(req)
}

func (c *TestClient) Get(path string) (*http.Response, error) {
    req, err := http.NewRequest("GET", c.baseURL+path, nil)
    if err != nil {
        return nil, err
    }

    if c.token != "" {
        req.Header.Set("Authorization", "Bearer "+c.token)
    }

    return c.httpClient.Do(req)
}

func (c *TestClient) Delete(path string) (*http.Response, error) {
    req, err := http.NewRequest("DELETE", c.baseURL+path, nil)
    if err != nil {
        return nil, err
    }

    if c.token != "" {
        req.Header.Set("Authorization", "Bearer "+c.token)
    }

    return c.httpClient.Do(req)
}

func ParseJSON(resp *http.Response) (map[string]interface{}, error) {
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    err = json.Unmarshal(body, &result)
    return result, err
}
```

### 3.2 用户注册登录流程

```go
// tests/e2e/user_flow_test.go
// +build e2e

package e2e

import (
    "net/http"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

var client *TestClient

func TestMain(m *testing.M) {
    baseURL := os.Getenv("API_BASE_URL")
    if baseURL == "" {
        baseURL = "http://localhost:8080"
    }
    client = NewTestClient(baseURL)

    os.Exit(m.Run())
}

func TestUserFlow_RegisterLoginAndAccess(t *testing.T) {
    // 步骤 1: 注册新用户
    t.Run("register", func(t *testing.T) {
        resp, err := client.Post("/api/auth/register", map[string]string{
            "username": "e2euser",
            "email":    "e2e@example.com",
            "password": "Password123",
        })
        require.NoError(t, err)
        assert.Equal(t, http.StatusCreated, resp.StatusCode)

        data, err := ParseJSON(resp)
        require.NoError(t, err)
        assert.NotEmpty(t, data["access_token"])

        // 保存 token
        client.SetToken(data["access_token"].(string))
    })

    // 步骤 2: 获取用户信息
    t.Run("get profile", func(t *testing.T) {
        resp, err := client.Get("/api/users/me")
        require.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)

        data, err := ParseJSON(resp)
        require.NoError(t, err)
        assert.Equal(t, "e2euser", data["username"])
    })

    // 步骤 3: 登出并重新登录
    t.Run("login", func(t *testing.T) {
        client.SetToken("") // 清除 token

        resp, err := client.Post("/api/auth/login", map[string]string{
            "email":    "e2e@example.com",
            "password": "Password123",
        })
        require.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)

        data, err := ParseJSON(resp)
        require.NoError(t, err)
        assert.NotEmpty(t, data["access_token"])

        client.SetToken(data["access_token"].(string))
    })
}
```

### 3.3 待办事项完整流程

```go
// tests/e2e/todo_flow_test.go
// +build e2e

package e2e

import (
    "fmt"
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestTodoFlow_CRUD(t *testing.T) {
    // 前置: 登录获取 token
    loginAndSetToken(t)

    var todoID float64

    // 创建待办
    t.Run("create todo", func(t *testing.T) {
        resp, err := client.Post("/api/todos", map[string]string{
            "title": "E2E Test Todo",
        })
        require.NoError(t, err)
        assert.Equal(t, http.StatusCreated, resp.StatusCode)

        data, err := ParseJSON(resp)
        require.NoError(t, err)
        todoID = data["id"].(float64)
        assert.NotZero(t, todoID)
    })

    // 获取待办列表
    t.Run("list todos", func(t *testing.T) {
        resp, err := client.Get("/api/todos")
        require.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)

        data, err := ParseJSON(resp)
        require.NoError(t, err)
        todos := data["todos"].([]interface{})
        assert.GreaterOrEqual(t, len(todos), 1)
    })

    // 获取单个待办
    t.Run("get todo", func(t *testing.T) {
        resp, err := client.Get(fmt.Sprintf("/api/todos/%d", int(todoID)))
        require.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)

        data, err := ParseJSON(resp)
        require.NoError(t, err)
        assert.Equal(t, "E2E Test Todo", data["title"])
    })

    // 完成待办
    t.Run("complete todo", func(t *testing.T) {
        resp, err := client.Put(fmt.Sprintf("/api/todos/%d", int(todoID)), map[string]bool{
            "completed": true,
        })
        require.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)

        data, err := ParseJSON(resp)
        require.NoError(t, err)
        assert.True(t, data["completed"].(bool))
    })

    // 删除待办
    t.Run("delete todo", func(t *testing.T) {
        resp, err := client.Delete(fmt.Sprintf("/api/todos/%d", int(todoID)))
        require.NoError(t, err)
        assert.Equal(t, http.StatusNoContent, resp.StatusCode)
    })

    // 验证删除
    t.Run("verify deleted", func(t *testing.T) {
        resp, err := client.Get(fmt.Sprintf("/api/todos/%d", int(todoID)))
        require.NoError(t, err)
        assert.Equal(t, http.StatusNotFound, resp.StatusCode)
    })
}

func loginAndSetToken(t *testing.T) {
    t.Helper()

    // 先尝试注册
    client.Post("/api/auth/register", map[string]string{
        "username": "todouser",
        "email":    "todo@example.com",
        "password": "Password123",
    })

    // 登录
    resp, err := client.Post("/api/auth/login", map[string]string{
        "email":    "todo@example.com",
        "password": "Password123",
    })
    require.NoError(t, err)

    data, err := ParseJSON(resp)
    require.NoError(t, err)
    client.SetToken(data["access_token"].(string))
}
```

## 4. 场景测试

### 4.1 错误场景测试

```go
func TestErrorScenarios(t *testing.T) {
    t.Run("register with existing email", func(t *testing.T) {
        // 第一次注册
        client.Post("/api/auth/register", map[string]string{
            "username": "user1",
            "email":    "dup@example.com",
            "password": "Password123",
        })

        // 第二次注册相同邮箱
        resp, err := client.Post("/api/auth/register", map[string]string{
            "username": "user2",
            "email":    "dup@example.com",
            "password": "Password123",
        })
        require.NoError(t, err)
        assert.Equal(t, http.StatusConflict, resp.StatusCode)
    })

    t.Run("login with wrong password", func(t *testing.T) {
        resp, err := client.Post("/api/auth/login", map[string]string{
            "email":    "dup@example.com",
            "password": "WrongPassword",
        })
        require.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    })

    t.Run("access protected route without token", func(t *testing.T) {
        client.SetToken("")
        resp, err := client.Get("/api/todos")
        require.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    })
}
```

### 4.2 并发场景测试

```go
func TestConcurrentAccess(t *testing.T) {
    loginAndSetToken(t)

    // 并发创建待办
    var wg sync.WaitGroup
    errors := make(chan error, 10)

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            resp, err := client.Post("/api/todos", map[string]string{
                "title": fmt.Sprintf("Concurrent Todo %d", idx),
            })
            if err != nil {
                errors <- err
                return
            }
            if resp.StatusCode != http.StatusCreated {
                errors <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
            }
        }(i)
    }

    wg.Wait()
    close(errors)

    for err := range errors {
        t.Errorf("concurrent request failed: %v", err)
    }

    // 验证所有待办都创建成功
    resp, _ := client.Get("/api/todos")
    data, _ := ParseJSON(resp)
    todos := data["todos"].([]interface{})
    assert.GreaterOrEqual(t, len(todos), 10)
}
```

## 5. 测试数据管理

### 5.1 测试数据清理

```go
func cleanupTestData(t *testing.T) {
    t.Helper()

    // 删除测试用户的所有数据
    resp, err := client.Delete("/api/users/me")
    if err != nil {
        t.Logf("cleanup warning: %v", err)
    }
    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusUnauthorized {
        t.Logf("cleanup warning: unexpected status %d", resp.StatusCode)
    }
}
```

### 5.2 测试隔离

```go
func TestWithCleanup(t *testing.T) {
    // 使用唯一标识符
    uniqueID := fmt.Sprintf("test_%d", time.Now().UnixNano())

    resp, err := client.Post("/api/auth/register", map[string]string{
        "username": uniqueID,
        "email":    uniqueID + "@example.com",
        "password": "Password123",
    })
    require.NoError(t, err)

    t.Cleanup(func() {
        // 测试结束后清理
        client.Delete("/api/users/me")
    })

    // 测试代码...
}
```

## 6. CI/CD 集成

```yaml
# .github/workflows/e2e.yml
name: E2E Tests

on:
  push:
    branches: [ main ]

jobs:
  e2e:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Start services
        run: docker-compose -f docker-compose.test.yml up -d --build

      - name: Wait for services
        run: |
          sleep 30
          curl --retry 10 --retry-delay 5 --retry-connrefused http://localhost:8080/health

      - name: Run E2E tests
        run: go test -tags=e2e -v ./tests/e2e/...
        env:
          API_BASE_URL: http://localhost:8080

      - name: Cleanup
        if: always()
        run: docker-compose -f docker-compose.test.yml down -v
```

## 练习

1. 编写用户注册登录的完整流程测试
2. 测试并发场景下的数据一致性
3. 实现测试数据的自动清理
4. 配置 CI 运行 E2E 测试

## 下一步

[性能测试](../04-performance/) - 基准测试和性能分析。
