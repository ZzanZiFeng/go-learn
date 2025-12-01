# 单元测试

## 学习目标

掌握 Go 单元测试的编写、组织和最佳实践。

## 1. go test 基础

### 1.1 第一个测试

```go
// calculator.go
package calculator

func Add(a, b int) int {
    return a + b
}

func Subtract(a, b int) int {
    return a - b
}
```

```go
// calculator_test.go
package calculator

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}

func TestSubtract(t *testing.T) {
    result := Subtract(5, 3)
    if result != 2 {
        t.Errorf("Subtract(5, 3) = %d; want 2", result)
    }
}
```

### 1.2 运行测试

```bash
# 运行当前包的测试
go test

# 运行所有包的测试
go test ./...

# 显示详细输出
go test -v

# 运行特定测试
go test -run TestAdd
go test -run "TestAdd|TestSubtract"

# 检测竞态条件
go test -race

# 查看测试覆盖率
go test -cover

# 生成覆盖率报告
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 2. 表驱动测试

### 2.1 基本模式

```go
func TestAdd_TableDriven(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed numbers", -2, 3, 1},
        {"zeros", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

### 2.2 包含错误场景

```go
func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func TestDivide(t *testing.T) {
    tests := []struct {
        name      string
        a, b      float64
        expected  float64
        expectErr bool
    }{
        {"normal division", 10, 2, 5, false},
        {"division by zero", 10, 0, 0, true},
        {"negative result", -10, 2, -5, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Divide(tt.a, tt.b)

            if tt.expectErr {
                if err == nil {
                    t.Error("expected error but got nil")
                }
                return
            }

            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }

            if result != tt.expected {
                t.Errorf("Divide(%f, %f) = %f; want %f",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

## 3. 使用 testify

### 3.1 安装

```bash
go get github.com/stretchr/testify
```

### 3.2 断言

```go
package calculator

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAdd_Testify(t *testing.T) {
    // assert - 失败后继续执行
    assert.Equal(t, 5, Add(2, 3), "should add correctly")
    assert.NotEqual(t, 6, Add(2, 3))

    // require - 失败后立即停止
    require.Equal(t, 5, Add(2, 3))

    // 更多断言
    assert.True(t, Add(1, 1) == 2)
    assert.False(t, Add(1, 1) == 3)
    assert.Nil(t, nil)
    assert.NotNil(t, "value")
    assert.Contains(t, "hello world", "hello")
    assert.Len(t, []int{1, 2, 3}, 3)
}

func TestDivide_Testify(t *testing.T) {
    // 测试错误
    _, err := Divide(10, 0)
    assert.Error(t, err)
    assert.EqualError(t, err, "division by zero")

    // 测试无错误
    result, err := Divide(10, 2)
    assert.NoError(t, err)
    assert.Equal(t, 5.0, result)
}
```

## 4. Mock 和接口

### 4.1 定义接口

```go
// repository.go
package user

type UserRepository interface {
    FindByID(id int64) (*User, error)
    Create(user *User) error
    Update(user *User) error
}

type User struct {
    ID    int64
    Name  string
    Email string
}
```

### 4.2 手动 Mock

```go
// repository_mock.go
package user

import "errors"

type MockUserRepository struct {
    Users map[int64]*User
}

func NewMockUserRepository() *MockUserRepository {
    return &MockUserRepository{
        Users: make(map[int64]*User),
    }
}

func (m *MockUserRepository) FindByID(id int64) (*User, error) {
    user, ok := m.Users[id]
    if !ok {
        return nil, errors.New("user not found")
    }
    return user, nil
}

func (m *MockUserRepository) Create(user *User) error {
    m.Users[user.ID] = user
    return nil
}

func (m *MockUserRepository) Update(user *User) error {
    if _, ok := m.Users[user.ID]; !ok {
        return errors.New("user not found")
    }
    m.Users[user.ID] = user
    return nil
}
```

### 4.3 使用 testify/mock

```go
// repository_testify_mock.go
package user

import (
    "github.com/stretchr/testify/mock"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) FindByID(id int64) (*User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) Create(user *User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockRepository) Update(user *User) error {
    args := m.Called(user)
    return args.Error(0)
}
```

### 4.4 使用 Mock 测试

```go
// service.go
package user

type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int64) (*User, error) {
    return s.repo.FindByID(id)
}
```

```go
// service_test.go
package user

import (
    "errors"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestUserService_GetUser(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        mockRepo := new(MockRepository)
        service := NewUserService(mockRepo)

        expectedUser := &User{ID: 1, Name: "Alice"}
        mockRepo.On("FindByID", int64(1)).Return(expectedUser, nil)

        user, err := service.GetUser(1)

        assert.NoError(t, err)
        assert.Equal(t, expectedUser, user)
        mockRepo.AssertExpectations(t)
    })

    t.Run("not found", func(t *testing.T) {
        mockRepo := new(MockRepository)
        service := NewUserService(mockRepo)

        mockRepo.On("FindByID", int64(999)).Return(nil, errors.New("not found"))

        user, err := service.GetUser(999)

        assert.Error(t, err)
        assert.Nil(t, user)
        mockRepo.AssertExpectations(t)
    })
}
```

## 5. 测试覆盖率

### 5.1 生成覆盖率报告

```bash
# 查看覆盖率百分比
go test -cover ./...

# 生成覆盖率文件
go test -coverprofile=coverage.out ./...

# 查看函数级别覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html

# 只覆盖特定包
go test -coverprofile=coverage.out -coverpkg=./internal/... ./...
```

### 5.2 覆盖率模式

```bash
# set: 是否覆盖 (默认)
go test -covermode=set -coverprofile=coverage.out

# count: 执行次数
go test -covermode=count -coverprofile=coverage.out

# atomic: 原子计数 (并发安全)
go test -covermode=atomic -coverprofile=coverage.out
```

## 6. 测试辅助函数

### 6.1 setup 和 teardown

```go
func TestMain(m *testing.M) {
    // 全局 setup
    fmt.Println("Setting up tests...")

    code := m.Run()

    // 全局 teardown
    fmt.Println("Tearing down tests...")
    os.Exit(code)
}

func setupTest(t *testing.T) func() {
    // 单个测试 setup
    t.Log("Setting up test...")

    return func() {
        // teardown
        t.Log("Tearing down test...")
    }
}

func TestSomething(t *testing.T) {
    teardown := setupTest(t)
    defer teardown()

    // 测试代码
}
```

### 6.2 使用 t.Cleanup

```go
func TestWithCleanup(t *testing.T) {
    // 创建临时文件
    tmpFile, err := os.CreateTemp("", "test")
    require.NoError(t, err)

    // 注册清理函数
    t.Cleanup(func() {
        os.Remove(tmpFile.Name())
    })

    // 使用临时文件进行测试
    // ...
}
```

### 6.3 测试辅助器

```go
// testutil/helper.go
package testutil

import "testing"

func AssertPanic(t *testing.T, f func()) {
    t.Helper() // 标记为辅助函数
    defer func() {
        if r := recover(); r == nil {
            t.Error("expected panic but didn't get one")
        }
    }()
    f()
}

// 使用
func TestPanic(t *testing.T) {
    testutil.AssertPanic(t, func() {
        panic("expected")
    })
}
```

## 7. 跳过测试

```go
func TestSlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping slow test in short mode")
    }
    // 耗时测试
}

func TestRequiresDocker(t *testing.T) {
    if os.Getenv("DOCKER_HOST") == "" {
        t.Skip("skipping: Docker not available")
    }
    // 需要 Docker 的测试
}
```

```bash
# 运行短测试
go test -short ./...
```

## 练习

1. 为一个计算器模块编写完整的表驱动测试
2. 使用 testify 重写测试
3. 实现 Mock 测试服务层
4. 达到 80% 以上的测试覆盖率

## 下一步

[集成测试](../02-integration/) - 测试组件之间的交互。
