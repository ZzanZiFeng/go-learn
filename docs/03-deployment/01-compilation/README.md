# 编译与打包

## 学习目标

掌握 Go 程序的编译、构建和发布。

## 1. go build 基础

### 1.1 基本编译

```bash
# 编译当前目录
go build

# 编译指定包
go build ./cmd/api

# 指定输出文件名
go build -o myapp ./cmd/api

# 编译并安装到 $GOPATH/bin
go install ./cmd/api
```

### 1.2 编译信息

```bash
# 查看编译过程
go build -v ./cmd/api

# 查看编译详情
go build -x ./cmd/api

# 编译但不链接（检查语法）
go build -n ./cmd/api
```

## 2. 交叉编译

### 2.1 GOOS 和 GOARCH

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o myapp-linux ./cmd/api

# Windows
GOOS=windows GOARCH=amd64 go build -o myapp.exe ./cmd/api

# macOS ARM (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o myapp-darwin ./cmd/api

# Linux ARM (树莓派)
GOOS=linux GOARCH=arm GOARM=7 go build -o myapp-arm ./cmd/api
```

### 2.2 常用平台组合

| GOOS | GOARCH | 描述 |
|------|--------|------|
| linux | amd64 | Linux 64位 |
| linux | arm64 | Linux ARM64 |
| darwin | amd64 | macOS Intel |
| darwin | arm64 | macOS Apple Silicon |
| windows | amd64 | Windows 64位 |

### 2.3 查看支持的平台

```bash
# 查看所有支持的平台
go tool dist list

# 输出类似：
# linux/amd64
# linux/arm64
# darwin/amd64
# darwin/arm64
# windows/amd64
# ...
```

## 3. 构建标志

### 3.1 ldflags 注入变量

```go
// main.go
package main

var (
    Version   = "dev"
    BuildTime = "unknown"
    GitCommit = "unknown"
)

func main() {
    fmt.Printf("Version: %s\n", Version)
    fmt.Printf("Build Time: %s\n", BuildTime)
    fmt.Printf("Git Commit: %s\n", GitCommit)
}
```

```bash
# 编译时注入变量
go build -ldflags "\
    -X main.Version=1.0.0 \
    -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
    -X main.GitCommit=$(git rev-parse --short HEAD)" \
    -o myapp ./cmd/api
```

### 3.2 减小二进制大小

```bash
# 移除调试信息
go build -ldflags "-s -w" -o myapp ./cmd/api

# 使用 UPX 进一步压缩
upx --best myapp

# 大小对比
# 普通编译:    ~15MB
# -s -w:      ~10MB
# UPX:        ~4MB
```

### 3.3 静态链接

```bash
# 完全静态链接（CGO_ENABLED=0）
CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o myapp ./cmd/api
```

## 4. 构建脚本

### 4.1 Makefile

```makefile
# Makefile
APP_NAME := myapp
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
LDFLAGS := -ldflags "-s -w \
    -X main.Version=$(VERSION) \
    -X main.BuildTime=$(BUILD_TIME) \
    -X main.GitCommit=$(GIT_COMMIT)"

.PHONY: all build clean test

all: clean build

build:
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/api

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux ./cmd/api

build-darwin:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin ./cmd/api

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME).exe ./cmd/api

build-all: build-linux build-darwin build-windows

clean:
	rm -rf bin/

test:
	go test -v ./...

run:
	go run ./cmd/api

.PHONY: docker
docker:
	docker build -t $(APP_NAME):$(VERSION) .
```

### 4.2 使用 Makefile

```bash
# 编译
make build

# 编译所有平台
make build-all

# 运行测试
make test

# 清理
make clean

# 构建 Docker 镜像
make docker
```

## 5. 版本管理

### 5.1 版本信息结构

```go
// internal/version/version.go
package version

import (
    "fmt"
    "runtime"
)

var (
    Version   = "dev"
    BuildTime = "unknown"
    GitCommit = "unknown"
)

type Info struct {
    Version   string `json:"version"`
    BuildTime string `json:"build_time"`
    GitCommit string `json:"git_commit"`
    GoVersion string `json:"go_version"`
    Platform  string `json:"platform"`
}

func Get() Info {
    return Info{
        Version:   Version,
        BuildTime: BuildTime,
        GitCommit: GitCommit,
        GoVersion: runtime.Version(),
        Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
    }
}

func (i Info) String() string {
    return fmt.Sprintf(
        "Version: %s\nBuild Time: %s\nGit Commit: %s\nGo Version: %s\nPlatform: %s",
        i.Version, i.BuildTime, i.GitCommit, i.GoVersion, i.Platform,
    )
}
```

### 5.2 版本端点

```go
// internal/handlers/version.go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/your-username/myapp/internal/version"
)

func Version(c *gin.Context) {
    c.JSON(http.StatusOK, version.Get())
}
```

## 练习

1. 为你的项目创建 Makefile
2. 添加版本信息注入
3. 编译不同平台的二进制文件
4. 比较不同编译选项的文件大小

## 下一步

[Docker 容器化](../02-docker/) - 将应用打包为 Docker 镜像。
