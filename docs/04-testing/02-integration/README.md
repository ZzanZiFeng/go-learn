# 集成测试

## 学习目标

掌握 Go 应用的集成测试，包括数据库和 API 测试。

## 1. 集成测试基础

### 1.1 使用构建标签

```go
// +build integration

package integration

// 运行集成测试
// go test -tags=integration ./...
```

### 1.2 测试目录结构

```
project/
├── internal/
│   ├── handlers/
│   ├── services/
│   └── repositories/
└── tests/
    ├── integration/
    │   ├── database_test.go
    │   └── api_test.go
    └── e2e/
        └── user_flow_test.go
```

## 2. 数据库集成测试

### 2.1 测试数据库配置

```go
// tests/integration/setup.go
package integration

import (
    "database/sql"
    "fmt"
    "os"
    "testing"

    _ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
    // 连接测试数据库
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        getEnv("TEST_DB_HOST", "localhost"),
        getEnv("TEST_DB_PORT", "5432"),
        getEnv("TEST_DB_USER", "test"),
        getEnv("TEST_DB_PASSWORD", "test"),
        getEnv("TEST_DB_NAME", "testdb"),
    )

    var err error
    testDB, err = sql.Open("postgres", dsn)
    if err != nil {
        panic(err)
    }

    // 运行迁移
    if err := runMigrations(testDB); err != nil {
        panic(err)
    }

    code := m.Run()

    // 清理
    testDB.Close()
    os.Exit(code)
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}

func cleanupTable(t *testing.T, table string) {
    t.Helper()
    _, err := testDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
    if err != nil {
        t.Fatalf("failed to cleanup table %s: %v", table, err)
    }
}
```

### 2.2 Repository 测试

```go
// tests/integration/user_repository_test.go
package integration

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "myapp/internal/models"
    "myapp/internal/repositories"
)

func TestUserRepository_Create(t *testing.T) {
    // 清理
    cleanupTable(t, "users")

    repo := repositories.NewUserRepository(testDB)

    user := &models.User{
        Username: "testuser",
        Email:    "test@example.com",
    }

    err := repo.Create(user)
    require.NoError(t, err)
    assert.NotZero(t, user.ID)
}

func TestUserRepository_FindByEmail(t *testing.T) {
    cleanupTable(t, "users")

    repo := repositories.NewUserRepository(testDB)

    // 创建测试用户
    user := &models.User{
        Username: "testuser",
        Email:    "test@example.com",
    }
    require.NoError(t, repo.Create(user))

    // 测试查找
    found, err := repo.FindByEmail("test@example.com")
    require.NoError(t, err)
    assert.Equal(t, user.Email, found.Email)

    // 测试未找到
    _, err = repo.FindByEmail("notfound@example.com")
    assert.Error(t, err)
}
```

## 3. 使用 testcontainers

### 3.1 安装

```bash
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
go get github.com/testcontainers/testcontainers-go/modules/redis
```

### 3.2 PostgreSQL 容器

```go
// tests/integration/containers.go
package integration

import (
    "context"
    "testing"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

func SetupPostgres(t *testing.T) (string, func()) {
    ctx := context.Background()

    container, err := postgres.Run(ctx,
        "postgres:15-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(5*time.Second),
        ),
    )
    if err != nil {
        t.Fatalf("failed to start container: %v", err)
    }

    connStr, err := container.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        t.Fatalf("failed to get connection string: %v", err)
    }

    cleanup := func() {
        if err := container.Terminate(ctx); err != nil {
            t.Logf("failed to terminate container: %v", err)
        }
    }

    return connStr, cleanup
}
```

### 3.3 使用容器测试

```go
// tests/integration/database_test.go
package integration

import (
    "database/sql"
    "testing"

    _ "github.com/lib/pq"
    "github.com/stretchr/testify/require"
)

func TestWithPostgresContainer(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    connStr, cleanup := SetupPostgres(t)
    defer cleanup()

    db, err := sql.Open("postgres", connStr)
    require.NoError(t, err)
    defer db.Close()

    // 创建表
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            username VARCHAR(255) NOT NULL,
            email VARCHAR(255) NOT NULL
        )
    `)
    require.NoError(t, err)

    // 插入数据
    _, err = db.Exec("INSERT INTO users (username, email) VALUES ($1, $2)",
        "testuser", "test@example.com")
    require.NoError(t, err)

    // 查询验证
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
    require.NoError(t, err)
    require.Equal(t, 1, count)
}
```

### 3.4 Redis 容器

```go
package integration

import (
    "context"
    "testing"

    "github.com/redis/go-redis/v9"
    "github.com/testcontainers/testcontainers-go/modules/redis"
)

func SetupRedis(t *testing.T) (*redis.RedisContainer, *redis.Client) {
    ctx := context.Background()

    container, err := redis.Run(ctx, "redis:7-alpine")
    if err != nil {
        t.Fatalf("failed to start redis container: %v", err)
    }

    host, err := container.Host(ctx)
    if err != nil {
        t.Fatalf("failed to get host: %v", err)
    }

    port, err := container.MappedPort(ctx, "6379")
    if err != nil {
        t.Fatalf("failed to get port: %v", err)
    }

    client := redis.NewClient(&redis.Options{
        Addr: fmt.Sprintf("%s:%s", host, port.Port()),
    })

    return container, client
}
```

## 4. API 集成测试

### 4.1 使用 httptest

```go
// tests/integration/api_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "myapp/internal/handlers"
)

func TestAPI_CreateUser(t *testing.T) {
    // 设置路由
    router := setupTestRouter(t)

    // 创建请求
    body := map[string]string{
        "username": "testuser",
        "email":    "test@example.com",
        "password": "password123",
    }
    jsonBody, _ := json.Marshal(body)

    req := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")

    // 执行请求
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // 验证响应
    assert.Equal(t, http.StatusCreated, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)
    assert.NotEmpty(t, response["id"])
}

func TestAPI_GetUser(t *testing.T) {
    router := setupTestRouter(t)

    // 先创建用户
    createUser(t, router, "testuser", "test@example.com")

    // 获取用户
    req := httptest.NewRequest("GET", "/api/users/1", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}

func setupTestRouter(t *testing.T) http.Handler {
    // 使用测试数据库设置路由
    connStr, cleanup := SetupPostgres(t)
    t.Cleanup(cleanup)

    db, err := sql.Open("postgres", connStr)
    require.NoError(t, err)
    t.Cleanup(func() { db.Close() })

    // 运行迁移
    runMigrations(db)

    // 创建 handler
    return handlers.SetupRoutes(db)
}

func createUser(t *testing.T, router http.Handler, username, email string) {
    body := map[string]string{
        "username": username,
        "email":    email,
        "password": "password123",
    }
    jsonBody, _ := json.Marshal(body)

    req := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    require.Equal(t, http.StatusCreated, w.Code)
}
```

### 4.2 测试认证 API

```go
func TestAPI_Login(t *testing.T) {
    router := setupTestRouter(t)

    // 创建用户
    createUser(t, router, "testuser", "test@example.com")

    // 登录
    loginBody := map[string]string{
        "email":    "test@example.com",
        "password": "password123",
    }
    jsonBody, _ := json.Marshal(loginBody)

    req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.NotEmpty(t, response["token"])
}

func TestAPI_ProtectedRoute(t *testing.T) {
    router := setupTestRouter(t)

    // 无 token 访问
    req := httptest.NewRequest("GET", "/api/users/me", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    assert.Equal(t, http.StatusUnauthorized, w.Code)

    // 获取 token
    token := getAuthToken(t, router)

    // 带 token 访问
    req = httptest.NewRequest("GET", "/api/users/me", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w = httptest.NewRecorder()
    router.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)
}
```

## 5. 测试工具函数

### 5.1 测试 fixtures

```go
// tests/fixtures/fixtures.go
package fixtures

import (
    "database/sql"
    "myapp/internal/models"
)

func CreateTestUser(db *sql.DB) *models.User {
    user := &models.User{
        Username: "testuser",
        Email:    "test@example.com",
    }
    db.QueryRow(
        "INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id",
        user.Username, user.Email,
    ).Scan(&user.ID)
    return user
}

func CreateTestTodo(db *sql.DB, userID int64) *models.Todo {
    todo := &models.Todo{
        Title:  "Test Todo",
        UserID: userID,
    }
    db.QueryRow(
        "INSERT INTO todos (title, user_id) VALUES ($1, $2) RETURNING id",
        todo.Title, todo.UserID,
    ).Scan(&todo.ID)
    return todo
}
```

### 5.2 测试助手

```go
// tests/helper.go
package tests

import (
    "encoding/json"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/require"
)

func ParseJSONResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
    t.Helper()
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)
    return response
}

func AssertStatusCode(t *testing.T, expected, actual int) {
    t.Helper()
    if expected != actual {
        t.Errorf("expected status %d, got %d", expected, actual)
    }
}
```

## 6. CI/CD 集成

### 6.1 GitHub Actions 配置

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: --health-cmd pg_isready --health-interval 10s

      redis:
        image: redis:7
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run integration tests
        env:
          TEST_DB_HOST: localhost
          TEST_DB_PORT: 5432
          TEST_DB_USER: test
          TEST_DB_PASSWORD: test
          TEST_DB_NAME: testdb
          REDIS_HOST: localhost
          REDIS_PORT: 6379
        run: go test -tags=integration -v ./tests/integration/...
```

## 练习

1. 使用 testcontainers 编写数据库测试
2. 编写 API 集成测试套件
3. 创建测试 fixtures 管理
4. 配置 CI 集成测试

## 下一步

[E2E测试](../03-e2e/) - 端到端测试完整用户流程。
