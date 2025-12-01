# 测试教程概览

## 学习目标

掌握 Go 应用的测试策略，从单元测试到性能测试。

## 章节结构

| 章节 | 主题 | 难度 | 时间 |
|------|------|------|------|
| TEST-01 | 单元测试 | 入门 | 2小时 |
| TEST-02 | 集成测试 | 中级 | 2小时 |
| TEST-03 | E2E测试 | 中级 | 2小时 |
| TEST-04 | 性能测试 | 中级 | 2小时 |
| TEST-05 | 调试技巧 | 中级 | 2小时 |

## 学习路径

```
┌─────────────────────────────────────────────────────────────────┐
│                       测试轨道学习路径                           │
└─────────────────────────────────────────────────────────────────┘

 TEST-01           TEST-02           TEST-03
┌──────────┐      ┌──────────┐      ┌──────────┐
│ 单元测试  │─────▶│ 集成测试  │─────▶│ E2E测试  │
│ go test  │      │ testcontainers │  │ 端到端   │
└──────────┘      └──────────┘      └──────────┘
                        │
                        ▼
              ┌──────────┐      ┌──────────┐
              │ 性能测试  │─────▶│ 调试技巧  │
              │ benchmark │      │  pprof   │
              └──────────┘      └──────────┘
               TEST-04           TEST-05
```

## 测试金字塔

```
                    ┌───────┐
                   /   E2E   \         少量，运行慢
                  /───────────\
                 /  集成测试    \       适量，中等速度
                /───────────────\
               /    单元测试      \     大量，运行快
              /─────────────────────\
```

| 测试类型 | 目的 | 速度 | 数量 |
|----------|------|------|------|
| 单元测试 | 测试函数/方法 | 毫秒级 | 最多 |
| 集成测试 | 测试组件交互 | 秒级 | 适量 |
| E2E测试 | 测试完整流程 | 分钟级 | 最少 |

## 前置要求

- 完成开发轨道 DEV-01（语法基础）
- 了解 Go 包和模块
- 基本的 HTTP API 知识

## 工具清单

| 工具 | 用途 | 安装方式 |
|------|------|----------|
| go test | 内置测试工具 | Go 自带 |
| testify | 断言和 Mock | go get github.com/stretchr/testify |
| mockgen | Mock 生成 | go install github.com/golang/mock/mockgen |
| testcontainers | 容器化测试 | go get github.com/testcontainers/testcontainers-go |
| pprof | 性能分析 | Go 自带 |
| delve | 调试器 | go install github.com/go-delve/delve/cmd/dlv |

## 测试命名规范

```go
// 文件命名: xxx_test.go
// 函数命名: Test + 函数名 + 场景

func TestAdd(t *testing.T) {}                    // 基本测试
func TestAdd_WithNegativeNumbers(t *testing.T) {} // 场景测试
func TestUser_Create_Success(t *testing.T) {}    // 方法测试
```

## 目录

1. **[单元测试](./01-unit-testing/)** - go test、表驱动、mocking
2. **[集成测试](./02-integration/)** - 数据库、API 测试
3. **[E2E测试](./03-e2e/)** - 端到端测试
4. **[性能测试](./04-performance/)** - 基准测试、pprof
5. **[调试技巧](./05-debugging/)** - 错误调试、delve

## 下一步

从 **[单元测试](./01-unit-testing/)** 开始学习！
