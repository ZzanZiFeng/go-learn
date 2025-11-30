# 并发编程 (DEV-03)

本章将帮助你掌握 Go 语言最强大的特性之一：并发编程。

## 学习目标

完成本章后，你将能够：

- 使用 goroutine 创建并发任务
- 使用 channel 在 goroutine 之间通信
- 使用 select 进行多路复用
- 使用 sync 包的同步原语
- 使用 context 进行取消和超时控制
- 实现常见的并发模式
- 检测和避免竞态条件

## 章节内容

| 文档 | 说明 | 预计时间 |
|------|------|----------|
| [Goroutines](./01-goroutines.md) | goroutine 基础与 async/await 对比 | 30min |
| [Channels](./02-channels.md) | channel 创建、发送和接收 | 35min |
| [缓冲 Channel](./03-buffered-chan.md) | 缓冲通道的使用 | 20min |
| [Select](./04-select.md) | 多路复用和超时处理 | 25min |
| [Sync 包](./05-sync-package.md) | WaitGroup, Mutex, RWMutex | 30min |
| [Context](./06-context.md) | 取消、超时和传值 | 30min |
| [并发模式](./07-patterns.md) | Worker Pool, Fan-out/Fan-in | 35min |
| [竞态检测](./08-race-detection.md) | 竞态检测和避免 | 25min |

## 前置要求

- 已完成 [结构体与接口](../02-struct-interface/) 章节
- 理解函数和指针
- 理解接口的基本概念

## 难度

🔴 **高级** - Go 并发是强大但需要谨慎使用的特性

## 核心概念

### Go 的并发哲学

> "Don't communicate by sharing memory; share memory by communicating."
> 不要通过共享内存来通信；通过通信来共享内存。

Go 的并发模型基于 **CSP**（Communicating Sequential Processes），使用 channel 作为主要的同步机制。

### 并发 vs 并行

| 概念 | 说明 | 示例 |
|-----|------|------|
| 并发 (Concurrency) | 同时处理多个任务的能力 | 一个人边吃饭边看手机 |
| 并行 (Parallelism) | 同时执行多个任务 | 两个人各自吃饭 |

Go 的 goroutine 是并发的，是否并行执行取决于 CPU 核心数和调度器。

### 与 JavaScript/TypeScript 对比

| 概念 | JavaScript/TypeScript | Go |
|-----|----------------------|-----|
| 异步函数 | `async function` | 普通函数 + `go` |
| 等待结果 | `await promise` | 从 channel 接收 |
| 并发模型 | 事件循环（单线程） | goroutine（多线程） |
| 同步原语 | Promise, async/await | channel, sync 包 |
| 错误处理 | try/catch | 通过 channel 传递 error |

```javascript
// JavaScript
async function fetchData() {
    const result = await fetch(url);
    return result.json();
}
```

```go
// Go
func fetchData(url string, ch chan<- Result) {
    result, err := http.Get(url)
    ch <- Result{Data: result, Err: err}
}

// 调用
ch := make(chan Result)
go fetchData(url, ch)
result := <-ch
```

## 配套代码

本章的示例代码位于 `examples/concurrency/` 目录：

```
examples/concurrency/
├── goroutines/     # goroutine 示例
├── channels/       # channel 示例
├── select/         # select 示例
├── sync/           # sync 包示例
├── context/        # context 示例
└── patterns/       # 并发模式示例
```

## 验收标准

完成本章后，请确认你能够：

- [ ] 创建 goroutine 并理解其生命周期
- [ ] 使用无缓冲和有缓冲 channel
- [ ] 使用 select 处理多个 channel
- [ ] 使用 WaitGroup 等待 goroutine 完成
- [ ] 使用 Mutex 保护共享数据
- [ ] 使用 context 实现取消和超时
- [ ] 实现 worker pool 模式
- [ ] 使用 `go run -race` 检测竞态条件

## 常见陷阱

⚠️ **本章需要特别注意的问题**：

1. **goroutine 泄漏** - goroutine 永远阻塞
2. **死锁** - 所有 goroutine 都在等待
3. **竞态条件** - 多个 goroutine 同时访问共享数据
4. **channel 死锁** - 发送/接收不匹配
5. **忘记关闭 channel** - 导致接收方永远等待

---

下一章: [后端架构](../04-architecture/)
