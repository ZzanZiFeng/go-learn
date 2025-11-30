# GOPATH 与 Go Modules

本文档解释 Go 依赖管理的演进历史，帮助你理解 GOPATH 和 Go Modules 的区别。

## 目录

- [历史背景](#历史背景)
- [GOPATH 模式](#gopath-模式)
- [Go Modules 模式](#go-modules-模式)
- [对比总结](#对比总结)
- [最佳实践](#最佳实践)

---

## 历史背景

Go 的依赖管理经历了几个阶段：

| 时期 | 方式 | 特点 |
|-----|------|------|
| Go 1.0 - 1.10 | GOPATH | 所有代码必须在 GOPATH 下 |
| Go 1.11 - 1.12 | Go Modules (实验) | 可选启用 modules |
| Go 1.13+ | Go Modules (默认) | modules 成为默认方式 |
| Go 1.16+ | Go Modules (强制) | 不再支持 GOPATH 模式 |

**重要**: 从 Go 1.16 开始，Go Modules 是唯一推荐的依赖管理方式。

---

## GOPATH 模式

### 什么是 GOPATH？

GOPATH 是 Go 的工作目录，用于存放源代码、编译后的二进制文件和下载的依赖包。

```bash
# 查看当前 GOPATH
go env GOPATH
# 默认值: $HOME/go (Linux/macOS) 或 %USERPROFILE%\go (Windows)
```

### GOPATH 目录结构

```
$GOPATH/
├── bin/          # 编译后的可执行文件
├── pkg/          # 编译后的包文件
└── src/          # 源代码
    └── github.com/
        └── username/
            └── project/
                ├── main.go
                └── ...
```

### GOPATH 模式的问题

1. **代码位置受限**: 所有项目必须在 `$GOPATH/src` 下
2. **版本管理困难**: 无法指定依赖的具体版本
3. **依赖冲突**: 同一个包只能有一个版本
4. **可重复构建难**: 不同时间构建可能获得不同版本的依赖

### 与 Node.js 对比

| 问题 | GOPATH | Node.js (npm) |
|-----|--------|---------------|
| 代码位置 | 必须在 GOPATH/src | 任意位置 |
| 版本锁定 | 无 | package-lock.json |
| 依赖隔离 | 全局共享 | node_modules 隔离 |

---

## Go Modules 模式

### 什么是 Go Modules？

Go Modules 是 Go 1.11 引入的官方依赖管理系统，解决了 GOPATH 的所有问题。

### 核心概念

| 概念 | 说明 | 类比 (Node.js) |
|-----|------|---------------|
| `go.mod` | 模块定义和依赖声明 | `package.json` |
| `go.sum` | 依赖版本校验 | `package-lock.json` |
| Module | 版本化的代码集合 | npm package |
| Module Path | 模块的唯一标识 | package name |

### 创建新模块

```bash
# 在任意目录下创建项目
mkdir myproject
cd myproject

# 初始化 Go Module
go mod init github.com/username/myproject

# 查看生成的 go.mod
cat go.mod
```

### go.mod 文件结构

```go
module github.com/username/myproject

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/spf13/viper v1.17.0
)

require (
    // indirect dependencies (间接依赖)
    github.com/pelletier/go-toml/v2 v2.1.0 // indirect
)
```

### 常用命令

```bash
# 初始化模块
go mod init <module-path>

# 添加依赖（自动下载最新版本）
go get github.com/gin-gonic/gin

# 添加指定版本的依赖
go get github.com/gin-gonic/gin@v1.9.1

# 更新依赖到最新版本
go get -u github.com/gin-gonic/gin

# 更新所有依赖
go get -u ./...

# 整理依赖（删除未使用的，添加缺失的）
go mod tidy

# 下载所有依赖到本地缓存
go mod download

# 查看依赖图
go mod graph

# 验证依赖
go mod verify
```

### 版本号规则

Go Modules 使用语义化版本 (Semantic Versioning)：

```
v1.2.3
│ │ │
│ │ └── 补丁版本 (bug fixes)
│ └──── 次版本 (新功能，向后兼容)
└────── 主版本 (破坏性变更)
```

**特殊版本**:
```bash
go get pkg@latest        # 最新稳定版
go get pkg@v1.2.3        # 指定版本
go get pkg@master        # 分支名
go get pkg@commit-hash   # 提交哈希
```

### go.sum 文件

```
github.com/gin-gonic/gin v1.9.1 h1:4+fr/el88TOO3ewCmQr8cx/CtZ/umlIRIs5M4NTNjf8=
github.com/gin-gonic/gin v1.9.1/go.mod h1:hPrL7YrpYKXt5YId3A/Tn+7SEVvkntRZ0mQ/xI0Kv4Q=
```

- 每个依赖有两行记录
- `h1:` 后面是内容哈希值
- 确保依赖内容的完整性和一致性

---

## 对比总结

### GOPATH vs Go Modules

| 特性 | GOPATH | Go Modules |
|-----|--------|------------|
| 项目位置 | 必须在 GOPATH/src | 任意位置 |
| 版本管理 | 无 | 支持语义化版本 |
| 依赖锁定 | 无 | go.sum |
| 可重复构建 | 困难 | 保证 |
| 多版本共存 | 不支持 | 支持 |
| 状态 | 已废弃 | 推荐使用 |

### Go Modules vs npm/yarn

| 特性 | Go Modules | npm/yarn |
|-----|------------|----------|
| 配置文件 | go.mod | package.json |
| 锁定文件 | go.sum | package-lock.json / yarn.lock |
| 依赖存储 | 全局缓存 + 校验 | node_modules (本地) |
| 添加依赖 | go get | npm install |
| 清理依赖 | go mod tidy | npm prune |

---

## 最佳实践

### 1. 始终使用 Go Modules

```bash
# 确保 GO111MODULE 设置为 on（Go 1.16+ 默认）
go env -w GO111MODULE=on
```

### 2. 选择合适的模块路径

```bash
# 公开项目：使用 GitHub 路径
go mod init github.com/username/project

# 私有项目：使用公司域名
go mod init company.com/team/project

# 学习/练习：使用任意名称
go mod init myproject
```

### 3. 及时运行 go mod tidy

```bash
# 在以下情况运行 tidy:
# - 添加新的 import
# - 删除 import
# - 提交代码前
go mod tidy
```

### 4. 提交 go.sum 到版本控制

```bash
# .gitignore 中不要忽略 go.sum
# go.sum 确保团队成员使用相同版本的依赖
```

### 5. 配置私有模块（企业环境）

```bash
# 设置私有模块路径
go env -w GOPRIVATE=company.com

# 设置 Git 使用 SSH
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

---

## 实践练习

### 练习 1: 创建新模块

```bash
# 1. 创建项目目录
mkdir ~/projects/hello-module
cd ~/projects/hello-module

# 2. 初始化模块
go mod init hello-module

# 3. 创建 main.go
cat > main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello, Modules!")
}
EOF

# 4. 运行程序
go run .
```

### 练习 2: 添加外部依赖

```bash
# 1. 添加 color 库
go get github.com/fatih/color

# 2. 修改 main.go 使用 color
cat > main.go << 'EOF'
package main

import "github.com/fatih/color"

func main() {
    color.Cyan("Hello, Modules!")
    color.Green("Dependencies work!")
}
EOF

# 3. 运行程序
go run .

# 4. 查看 go.mod 和 go.sum 的变化
cat go.mod
cat go.sum
```

---

## 下一步

- [VS Code 配置](./vscode-setup.md) - 配置 Go 开发环境
- [第一个程序](./first-program.md) - 深入理解 Hello World

---

## 参考资源

- [Go Modules 官方文档](https://go.dev/doc/modules/)
- [Go Modules 参考手册](https://go.dev/ref/mod)
- [Go Modules Wiki](https://github.com/golang/go/wiki/Modules)
