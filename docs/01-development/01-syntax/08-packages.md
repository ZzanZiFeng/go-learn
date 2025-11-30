# 包管理

本文档介绍 Go 语言的包系统和模块管理。

## 目录

- [包基础](#包基础)
- [导入包](#导入包)
- [创建包](#创建包)
- [Go Modules](#go-modules)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 包基础

### 什么是包？

包是 Go 组织代码的基本单位。每个 Go 文件都属于一个包。

```go
package main  // 包声明

import "fmt"  // 导入其他包

func main() {
    fmt.Println("Hello!")
}
```

### 包的规则

1. **同一目录下所有文件必须属于同一个包**
2. **包名通常与目录名相同**（但不强制）
3. **main 包是特殊的**：包含 `main()` 函数的可执行程序
4. **首字母大写的标识符是导出的**（公开的）

```go
// 在 mypackage/ 目录下

// mypackage/utils.go
package mypackage

// 导出的（公开的）
func PublicFunction() {}
var PublicVar = 10
type PublicType struct{}

// 未导出的（私有的）
func privateFunction() {}
var privateVar = 20
type privateType struct{}
```

---

## 导入包

### 基本导入

```go
package main

import "fmt"
import "strings"
import "os"

func main() {
    fmt.Println(strings.ToUpper("hello"))
}
```

### 分组导入（推荐）

```go
package main

import (
    "fmt"
    "os"
    "strings"
)

func main() {
    // ...
}
```

### 导入别名

```go
package main

import (
    "fmt"
    str "strings"  // 别名
)

func main() {
    fmt.Println(str.ToUpper("hello"))
}
```

### 点导入（不推荐）

```go
package main

import (
    . "fmt"  // 点导入
)

func main() {
    Println("Hello")  // 不需要 fmt. 前缀
}
```

### 空白导入

只执行包的 init 函数，不使用其他内容。

```go
import (
    "database/sql"
    _ "github.com/lib/pq"  // 只为了注册 PostgreSQL 驱动
)
```

### 标准库常用包

| 包名 | 用途 |
|-----|------|
| `fmt` | 格式化输入输出 |
| `os` | 操作系统功能 |
| `io` | I/O 原语 |
| `strings` | 字符串操作 |
| `strconv` | 字符串与基本类型转换 |
| `time` | 时间处理 |
| `net/http` | HTTP 客户端和服务器 |
| `encoding/json` | JSON 编解码 |
| `errors` | 错误处理 |
| `context` | 上下文管理 |
| `sync` | 同步原语 |
| `log` | 日志 |

---

## 创建包

### 项目结构

```
myproject/
├── go.mod
├── main.go
└── utils/
    ├── math.go
    └── string.go
```

### utils/math.go

```go
package utils

// Add 两个数相加
func Add(a, b int) int {
    return a + b
}

// Multiply 两个数相乘
func Multiply(a, b int) int {
    return a * b
}

// 私有函数
func helper() {
    // ...
}
```

### utils/string.go

```go
package utils

import "strings"

// Reverse 反转字符串
func Reverse(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

// Contains 检查字符串是否包含子串
func Contains(s, substr string) bool {
    return strings.Contains(s, substr)
}
```

### main.go

```go
package main

import (
    "fmt"
    "myproject/utils"
)

func main() {
    sum := utils.Add(3, 5)
    fmt.Println("Sum:", sum)

    reversed := utils.Reverse("Hello")
    fmt.Println("Reversed:", reversed)
}
```

### init 函数

每个包可以有多个 `init` 函数，在包被导入时自动执行。

```go
package mypackage

import "log"

var config map[string]string

func init() {
    log.Println("Initializing mypackage...")
    config = make(map[string]string)
}

func init() {
    // 可以有多个 init
    config["default"] = "value"
}
```

执行顺序：
1. 导入的包的 init
2. 包级别变量初始化
3. init 函数（按声明顺序）
4. main 函数（仅 main 包）

---

## Go Modules

### 初始化模块

```bash
# 创建新项目
mkdir myproject
cd myproject

# 初始化模块
go mod init github.com/username/myproject

# 查看 go.mod
cat go.mod
```

### go.mod 文件

```
module github.com/username/myproject

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/spf13/viper v1.17.0
)
```

### 常用命令

```bash
# 添加依赖
go get github.com/gin-gonic/gin

# 添加特定版本
go get github.com/gin-gonic/gin@v1.9.1

# 更新依赖
go get -u github.com/gin-gonic/gin

# 更新所有依赖
go get -u ./...

# 整理依赖（添加缺失的，删除未使用的）
go mod tidy

# 下载依赖
go mod download

# 查看依赖图
go mod graph

# 验证依赖
go mod verify

# 清理模块缓存
go clean -modcache
```

### 工作区 (Go 1.18+)

多模块项目可以使用工作区：

```bash
# 创建工作区
go work init ./module1 ./module2

# 添加模块到工作区
go work use ./module3
```

`go.work` 文件：
```
go 1.21

use (
    ./module1
    ./module2
    ./module3
)
```

### 私有模块

```bash
# 设置私有模块路径（不通过代理）
go env -w GOPRIVATE=github.com/mycompany/*

# 或在环境变量中设置
export GOPRIVATE=github.com/mycompany/*
```

---

## 与 JavaScript/TypeScript 对比

### 导入语法对比

```javascript
// JavaScript ES Modules
import { readFile } from 'fs';
import * as path from 'path';
import express from 'express';
import { useState, useEffect } from 'react';

// CommonJS
const fs = require('fs');
const { readFile } = require('fs');
```

```go
// Go
import "fmt"
import "os"

import (
    "fmt"
    "os"
    "github.com/gin-gonic/gin"
)
```

### 导出对比

```javascript
// JavaScript
// 命名导出
export function add(a, b) { return a + b; }
export const PI = 3.14;

// 默认导出
export default class MyClass {}

// 混合
export { add, PI };
export default MyClass;
```

```go
// Go - 首字母大写 = 导出
package mypackage

func Add(a, b int) int { return a + b }  // 导出
const PI = 3.14                           // 导出

func helper() {}  // 未导出（私有）
const secret = 1  // 未导出
```

### 包管理工具对比

| 特性 | npm/yarn | Go Modules |
|-----|----------|------------|
| 配置文件 | package.json | go.mod |
| 锁定文件 | package-lock.json | go.sum |
| 依赖存储 | node_modules/ | 全局缓存 |
| 安装命令 | npm install | go get |
| 更新命令 | npm update | go get -u |
| 清理 | npm prune | go mod tidy |
| 发布 | npm publish | git tag + push |

### 依赖管理对比

```json
// package.json
{
  "name": "myproject",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.0",
    "lodash": "~4.17.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0"
  }
}
```

```
// go.mod
module github.com/username/myproject

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/spf13/viper v1.17.0
)

// Go 没有 devDependencies 的区分
```

### 主要差异

| 特性 | JavaScript | Go |
|-----|-----------|-----|
| 导出机制 | export 关键字 | 首字母大写 |
| 默认导出 | 支持 | 不支持 |
| 导入粒度 | 可选择性导入 | 导入整个包 |
| 循环依赖 | 可能有问题 | 编译错误 |
| 包版本 | 可多版本共存 | 单一版本 |
| 依赖存储 | 项目本地 | 全局缓存 |
| 发布机制 | npm registry | Git 仓库 |

### 目录结构对比

```
// JavaScript 项目
myproject/
├── package.json
├── node_modules/
├── src/
│   ├── index.js
│   └── utils/
│       └── helper.js
└── tests/
    └── helper.test.js
```

```
// Go 项目
myproject/
├── go.mod
├── go.sum
├── main.go
├── internal/          # 内部包（不可被外部导入）
│   └── helper/
│       └── helper.go
├── pkg/               # 可被外部导入的包
│   └── utils/
│       └── utils.go
└── cmd/               # 可执行程序入口
    └── myapp/
        └── main.go
```

---

## 最佳实践

### 1. 包命名

```go
// 好的命名
package http
package json
package user
package repository

// 不好的命名
package util       // 太泛
package common     // 太泛
package helpers    // 太泛
package myPackage  // 不要用驼峰
```

### 2. 导入分组

```go
import (
    // 标准库
    "context"
    "fmt"
    "net/http"

    // 第三方库
    "github.com/gin-gonic/gin"
    "github.com/spf13/viper"

    // 本项目包
    "myproject/internal/service"
    "myproject/pkg/utils"
)
```

### 3. 使用 internal 包

```
myproject/
├── internal/           # 只能被本项目导入
│   └── secret/
│       └── secret.go
└── pkg/                # 可被外部项目导入
    └── public/
        └── public.go
```

### 4. 避免循环依赖

```go
// 不好：循环依赖
// package a 导入 package b
// package b 导入 package a

// 好：提取公共部分到第三个包
// package a 导入 package common
// package b 导入 package common
```

### 5. 文档注释

```go
// Package utils provides utility functions for string manipulation.
package utils

// Reverse reverses the given string.
// It handles UTF-8 strings correctly.
//
// Example:
//
//  result := utils.Reverse("hello")
//  // result is "olleh"
func Reverse(s string) string {
    // ...
}
```

---

## 下一步

- [JS/TS 对比总结](./js-comparison.md) - 完整的 Go 与 JavaScript/TypeScript 差异对比
