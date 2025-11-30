# 密码安全 (Password Security)

## 概述

密码安全是认证系统的基础。本节介绍密码哈希和安全存储的最佳实践。

## 与 JavaScript 对比

### Node.js (bcrypt)

```javascript
const bcrypt = require('bcrypt');

// 哈希
const hash = await bcrypt.hash(password, 10);

// 验证
const match = await bcrypt.compare(password, hash);
```

### Go (bcrypt)

```go
import "golang.org/x/crypto/bcrypt"

// 哈希
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 验证
err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
```

## 为什么不能明文存储？

```
❌ 错误做法：
  users 表：
  | email              | password    |
  | alice@example.com  | password123 |  <- 明文！

  风险：
  - 数据库泄露 = 所有密码泄露
  - 管理员可以看到密码
  - 用户可能在多个网站使用相同密码
```

## 为什么不用 MD5/SHA1？

```go
// ❌ 不安全
hash := md5.Sum([]byte(password))
hash := sha1.Sum([]byte(password))
hash := sha256.Sum256([]byte(password))

// 问题：
// 1. 太快 - 容易暴力破解
// 2. 无盐 - 彩虹表攻击
// 3. 确定性 - 相同密码产生相同哈希
```

## bcrypt

bcrypt 是专为密码设计的哈希算法，具有以下特点：
- 自动加盐
- 计算密集（抗暴力破解）
- 可调整成本因子

### 安装

```bash
go get -u golang.org/x/crypto/bcrypt
```

### 基本使用

```go
package auth

import (
    "errors"

    "golang.org/x/crypto/bcrypt"
)

var (
    ErrPasswordTooShort = errors.New("password must be at least 8 characters")
    ErrPasswordMismatch = errors.New("password does not match")
)

// HashPassword 哈希密码
func HashPassword(password string) (string, error) {
    if len(password) < 8 {
        return "", ErrPasswordTooShort
    }

    // cost 越高越安全，但也越慢
    // 默认值是 10，推荐 12-14
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }

    return string(hash), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) error {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    if err != nil {
        return ErrPasswordMismatch
    }
    return nil
}

// ValidatePassword 验证密码并返回布尔值
func ValidatePassword(password, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
```

### 成本因子

```go
// 成本因子决定计算时间
// 每增加 1，时间翻倍

// bcrypt.MinCost = 4
// bcrypt.MaxCost = 31
// bcrypt.DefaultCost = 10

// 推荐根据服务器性能选择
func GetOptimalCost() int {
    // 目标：哈希操作耗时 100-250ms

    // 低端服务器
    // return 10  // ~100ms

    // 标准服务器
    return 12  // ~300ms

    // 高端服务器
    // return 14  // ~1s
}

func HashPasswordWithCost(password string, cost int) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
    return string(hash), err
}

// 测试不同成本因子的性能
func BenchmarkCost() {
    password := "testpassword123"

    for cost := 10; cost <= 14; cost++ {
        start := time.Now()
        bcrypt.GenerateFromPassword([]byte(password), cost)
        fmt.Printf("Cost %d: %v\n", cost, time.Since(start))
    }
}
```

## Argon2 (更现代的选择)

Argon2 是 2015 年密码哈希竞赛的获胜者，比 bcrypt 更新、更安全。

### 安装

```bash
go get -u golang.org/x/crypto/argon2
```

### 基本使用

```go
package auth

import (
    "crypto/rand"
    "crypto/subtle"
    "encoding/base64"
    "fmt"
    "strings"

    "golang.org/x/crypto/argon2"
)

type Argon2Params struct {
    Memory      uint32  // 内存使用 (KB)
    Iterations  uint32  // 迭代次数
    Parallelism uint8   // 并行度
    SaltLength  uint32  // 盐长度
    KeyLength   uint32  // 输出密钥长度
}

var DefaultParams = &Argon2Params{
    Memory:      64 * 1024,  // 64MB
    Iterations:  3,
    Parallelism: 2,
    SaltLength:  16,
    KeyLength:   32,
}

// HashPasswordArgon2 使用 Argon2id 哈希密码
func HashPasswordArgon2(password string, params *Argon2Params) (string, error) {
    if params == nil {
        params = DefaultParams
    }

    // 生成随机盐
    salt := make([]byte, params.SaltLength)
    if _, err := rand.Read(salt); err != nil {
        return "", err
    }

    // 生成哈希
    hash := argon2.IDKey(
        []byte(password),
        salt,
        params.Iterations,
        params.Memory,
        params.Parallelism,
        params.KeyLength,
    )

    // 编码为字符串格式: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
    b64Salt := base64.RawStdEncoding.EncodeToString(salt)
    b64Hash := base64.RawStdEncoding.EncodeToString(hash)

    encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
        argon2.Version,
        params.Memory,
        params.Iterations,
        params.Parallelism,
        b64Salt,
        b64Hash,
    )

    return encoded, nil
}

// VerifyPasswordArgon2 验证密码
func VerifyPasswordArgon2(password, encodedHash string) (bool, error) {
    // 解析哈希
    params, salt, hash, err := decodeArgon2Hash(encodedHash)
    if err != nil {
        return false, err
    }

    // 使用相同参数计算哈希
    otherHash := argon2.IDKey(
        []byte(password),
        salt,
        params.Iterations,
        params.Memory,
        params.Parallelism,
        params.KeyLength,
    )

    // 常量时间比较，防止时序攻击
    if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
        return true, nil
    }

    return false, nil
}

func decodeArgon2Hash(encodedHash string) (*Argon2Params, []byte, []byte, error) {
    parts := strings.Split(encodedHash, "$")
    if len(parts) != 6 {
        return nil, nil, nil, errors.New("invalid hash format")
    }

    var version int
    _, err := fmt.Sscanf(parts[2], "v=%d", &version)
    if err != nil {
        return nil, nil, nil, err
    }

    params := &Argon2Params{}
    _, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d",
        &params.Memory, &params.Iterations, &params.Parallelism)
    if err != nil {
        return nil, nil, nil, err
    }

    salt, err := base64.RawStdEncoding.DecodeString(parts[4])
    if err != nil {
        return nil, nil, nil, err
    }

    hash, err := base64.RawStdEncoding.DecodeString(parts[5])
    if err != nil {
        return nil, nil, nil, err
    }

    params.SaltLength = uint32(len(salt))
    params.KeyLength = uint32(len(hash))

    return params, salt, hash, nil
}
```

## 密码强度验证

```go
package auth

import (
    "errors"
    "regexp"
    "unicode"
)

type PasswordPolicy struct {
    MinLength        int
    MaxLength        int
    RequireUppercase bool
    RequireLowercase bool
    RequireNumber    bool
    RequireSpecial   bool
}

var DefaultPolicy = &PasswordPolicy{
    MinLength:        8,
    MaxLength:        128,
    RequireUppercase: true,
    RequireLowercase: true,
    RequireNumber:    true,
    RequireSpecial:   true,
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func ValidatePasswordStrength(password string, policy *PasswordPolicy) []ValidationError {
    if policy == nil {
        policy = DefaultPolicy
    }

    var errors []ValidationError

    // 长度检查
    if len(password) < policy.MinLength {
        errors = append(errors, ValidationError{
            Field:   "password",
            Message: fmt.Sprintf("must be at least %d characters", policy.MinLength),
        })
    }

    if len(password) > policy.MaxLength {
        errors = append(errors, ValidationError{
            Field:   "password",
            Message: fmt.Sprintf("must be at most %d characters", policy.MaxLength),
        })
    }

    var (
        hasUpper   bool
        hasLower   bool
        hasNumber  bool
        hasSpecial bool
    )

    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }

    if policy.RequireUppercase && !hasUpper {
        errors = append(errors, ValidationError{
            Field:   "password",
            Message: "must contain at least one uppercase letter",
        })
    }

    if policy.RequireLowercase && !hasLower {
        errors = append(errors, ValidationError{
            Field:   "password",
            Message: "must contain at least one lowercase letter",
        })
    }

    if policy.RequireNumber && !hasNumber {
        errors = append(errors, ValidationError{
            Field:   "password",
            Message: "must contain at least one number",
        })
    }

    if policy.RequireSpecial && !hasSpecial {
        errors = append(errors, ValidationError{
            Field:   "password",
            Message: "must contain at least one special character",
        })
    }

    return errors
}

// 检查常见弱密码
var commonPasswords = map[string]bool{
    "password":  true,
    "12345678":  true,
    "123456789": true,
    "qwerty":    true,
    "abc123":    true,
    // ... 更多常见密码
}

func IsCommonPassword(password string) bool {
    return commonPasswords[strings.ToLower(password)]
}
```

## 密码服务

```go
package auth

type PasswordService struct {
    policy *PasswordPolicy
    cost   int
}

func NewPasswordService() *PasswordService {
    return &PasswordService{
        policy: DefaultPolicy,
        cost:   12,
    }
}

// Hash 哈希密码
func (s *PasswordService) Hash(password string) (string, error) {
    // 验证强度
    if errs := ValidatePasswordStrength(password, s.policy); len(errs) > 0 {
        return "", &PasswordValidationError{Errors: errs}
    }

    // 检查常见密码
    if IsCommonPassword(password) {
        return "", errors.New("password is too common")
    }

    return HashPasswordWithCost(password, s.cost)
}

// Verify 验证密码
func (s *PasswordService) Verify(password, hash string) error {
    return CheckPassword(password, hash)
}

// NeedsRehash 检查是否需要重新哈希
func (s *PasswordService) NeedsRehash(hash string) bool {
    cost, err := bcrypt.Cost([]byte(hash))
    if err != nil {
        return true
    }
    return cost < s.cost
}

type PasswordValidationError struct {
    Errors []ValidationError
}

func (e *PasswordValidationError) Error() string {
    return "password validation failed"
}
```

## 用户服务集成

```go
package services

type UserService struct {
    repo     *repositories.UserRepository
    password *auth.PasswordService
}

func (s *UserService) Create(name, email, password string) (*User, error) {
    // 哈希密码
    hashedPassword, err := s.password.Hash(password)
    if err != nil {
        return nil, err
    }

    user := &User{
        Name:     name,
        Email:    email,
        Password: hashedPassword,
    }

    return s.repo.Create(user)
}

func (s *UserService) ValidateCredentials(email, password string) (*User, error) {
    user, err := s.repo.FindByEmail(email)
    if err != nil {
        return nil, ErrInvalidCredentials
    }

    if err := s.password.Verify(password, user.Password); err != nil {
        return nil, ErrInvalidCredentials
    }

    // 检查是否需要升级哈希
    if s.password.NeedsRehash(user.Password) {
        go s.rehashPassword(user, password)
    }

    return user, nil
}

func (s *UserService) rehashPassword(user *User, password string) {
    newHash, err := s.password.Hash(password)
    if err != nil {
        return
    }
    s.repo.UpdatePassword(user.ID, newHash)
}

func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
    user, err := s.repo.FindByID(userID)
    if err != nil {
        return err
    }

    // 验证旧密码
    if err := s.password.Verify(oldPassword, user.Password); err != nil {
        return ErrInvalidCredentials
    }

    // 哈希新密码
    hashedPassword, err := s.password.Hash(newPassword)
    if err != nil {
        return err
    }

    return s.repo.UpdatePassword(userID, hashedPassword)
}
```

## 密码重置

```go
package auth

import (
    "crypto/rand"
    "encoding/hex"
    "time"
)

type PasswordResetToken struct {
    Token     string
    UserID    uint
    ExpiresAt time.Time
}

type PasswordResetService struct {
    store    TokenStore
    password *PasswordService
    ttl      time.Duration
}

func NewPasswordResetService(store TokenStore, password *PasswordService) *PasswordResetService {
    return &PasswordResetService{
        store:    store,
        password: password,
        ttl:      1 * time.Hour,
    }
}

// GenerateToken 生成重置令牌
func (s *PasswordResetService) GenerateToken(userID uint) (string, error) {
    // 生成随机令牌
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    token := hex.EncodeToString(bytes)

    // 存储令牌
    resetToken := &PasswordResetToken{
        Token:     token,
        UserID:    userID,
        ExpiresAt: time.Now().Add(s.ttl),
    }

    if err := s.store.Save(resetToken); err != nil {
        return "", err
    }

    return token, nil
}

// ValidateToken 验证令牌
func (s *PasswordResetService) ValidateToken(token string) (*PasswordResetToken, error) {
    resetToken, err := s.store.Get(token)
    if err != nil {
        return nil, errors.New("invalid token")
    }

    if time.Now().After(resetToken.ExpiresAt) {
        s.store.Delete(token)
        return nil, errors.New("token expired")
    }

    return resetToken, nil
}

// ResetPassword 重置密码
func (s *PasswordResetService) ResetPassword(token, newPassword string) error {
    resetToken, err := s.ValidateToken(token)
    if err != nil {
        return err
    }

    hashedPassword, err := s.password.Hash(newPassword)
    if err != nil {
        return err
    }

    // 更新密码（需要注入 UserRepository）
    // ...

    // 删除令牌
    return s.store.Delete(token)
}
```

## 最佳实践总结

| 实践 | 说明 |
|------|------|
| 使用 bcrypt/argon2 | 专为密码设计的哈希算法 |
| 适当的成本因子 | 平衡安全性和性能 |
| 密码强度验证 | 强制复杂度要求 |
| 禁止常见密码 | 检查弱密码列表 |
| 限制登录尝试 | 防止暴力破解 |
| 安全的重置流程 | 随机令牌 + 过期时间 |
| 定期升级哈希 | 随技术发展提高安全性 |

**下一节**：[会话管理](./04-session-management.md) - 学习 Session 管理
