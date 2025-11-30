# VS Code Go 开发环境配置

本文档将指导你配置 VS Code 进行 Go 开发，获得最佳的开发体验。

## 目录

- [安装 VS Code](#安装-vs-code)
- [安装 Go 扩展](#安装-go-扩展)
- [安装 Go 工具](#安装-go-工具)
- [配置设置](#配置设置)
- [常用功能](#常用功能)
- [快捷键](#快捷键)
- [调试配置](#调试配置)
- [常见问题](#常见问题)

---

## 安装 VS Code

### 下载安装

1. 访问 [VS Code 官网](https://code.visualstudio.com/)
2. 下载对应平台的安装包
3. 安装并启动 VS Code

### 命令行工具（可选）

安装 `code` 命令，可以从终端快速打开项目：

1. 打开 VS Code
2. 按 `Cmd/Ctrl + Shift + P`
3. 输入 "Shell Command: Install 'code' command in PATH"
4. 选择并执行

```bash
# 之后可以在终端中使用
code .              # 打开当前目录
code myproject/     # 打开指定目录
code main.go        # 打开指定文件
```

---

## 安装 Go 扩展

### 官方 Go 扩展

1. 打开 VS Code 扩展面板（`Cmd/Ctrl + Shift + X`）
2. 搜索 "Go"
3. 安装由 **Go Team at Google** 发布的官方扩展
4. 重启 VS Code

或通过命令行安装：

```bash
code --install-extension golang.go
```

### 扩展功能

安装后你将获得：

| 功能 | 说明 |
|-----|------|
| 智能提示 | 代码补全、参数提示 |
| 代码导航 | 跳转到定义、查找引用 |
| 自动格式化 | 保存时自动 `gofmt` |
| 代码诊断 | 实时语法检查、错误提示 |
| 调试支持 | 断点、变量查看、堆栈追踪 |
| 测试支持 | 运行测试、查看覆盖率 |
| 代码重构 | 重命名、提取函数 |

---

## 安装 Go 工具

首次打开 Go 文件时，VS Code 会提示安装 Go 工具。

### 自动安装

1. 打开任意 `.go` 文件
2. 点击右下角弹出的 "Install All" 提示
3. 等待安装完成

### 手动安装

如果自动安装失败：

```bash
# 确保配置了代理（国内用户）
go env -w GOPROXY=https://goproxy.cn,direct

# 通过命令面板安装
# Cmd/Ctrl + Shift + P > Go: Install/Update Tools
# 选择全部工具，点击 OK
```

### 必需工具列表

| 工具 | 用途 |
|-----|------|
| `gopls` | Go 语言服务器（核心） |
| `dlv` | 调试器 |
| `staticcheck` | 静态分析 |
| `gofumpt` | 代码格式化（增强版） |
| `gomodifytags` | 结构体标签管理 |
| `impl` | 接口实现生成 |
| `gotests` | 测试生成 |

---

## 配置设置

### 打开设置

1. `Cmd/Ctrl + ,` 打开设置
2. 搜索 "Go" 相关配置
3. 或直接编辑 `settings.json`

### 推荐设置

按 `Cmd/Ctrl + Shift + P`，输入 "Open User Settings (JSON)"，添加以下配置：

```json
{
    // === Go 基本设置 ===
    "go.useLanguageServer": true,
    "go.languageServerFlags": [
        "-rpc.trace"
    ],

    // === 保存时自动操作 ===
    "[go]": {
        "editor.formatOnSave": true,
        "editor.codeActionsOnSave": {
            "source.organizeImports": "explicit"
        },
        "editor.defaultFormatter": "golang.go"
    },
    "[go.mod]": {
        "editor.formatOnSave": true
    },

    // === 代码格式化 ===
    "go.formatTool": "gofumpt",

    // === 代码检查 ===
    "go.lintTool": "staticcheck",
    "go.lintOnSave": "package",
    "go.vetOnSave": "package",

    // === 测试设置 ===
    "go.testFlags": ["-v"],
    "go.coverOnSave": false,
    "go.coverOnSingleTest": true,
    "go.coverageDecorator": {
        "type": "highlight",
        "coveredHighlightColor": "rgba(64,128,128,0.2)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.2)"
    },

    // === 自动补全 ===
    "go.autocompleteUnimportedPackages": true,
    "go.useCodeSnippetsOnFunctionSuggest": true,
    "go.useCodeSnippetsOnFunctionSuggestWithoutType": true,

    // === 构建标签（如需要）===
    "go.buildTags": "",

    // === 调试设置 ===
    "go.delveConfig": {
        "debugAdapter": "dlv-dap",
        "showGlobalVariables": true
    }
}
```

### 工作区设置

对于特定项目，可以在项目根目录创建 `.vscode/settings.json`：

```json
{
    "go.testEnvVars": {
        "DB_HOST": "localhost",
        "DB_PORT": "5432"
    },
    "go.buildFlags": ["-tags=integration"]
}
```

---

## 常用功能

### 代码导航

| 功能 | 快捷键 | 说明 |
|-----|--------|------|
| 跳转到定义 | `F12` 或 `Cmd/Ctrl + Click` | 跳转到函数/变量定义 |
| 查看定义 | `Alt + F12` | 弹窗预览定义 |
| 查找引用 | `Shift + F12` | 查找所有使用位置 |
| 返回 | `Ctrl + -` | 返回上一个位置 |
| 前进 | `Ctrl + Shift + -` | 前进到下一个位置 |
| 转到符号 | `Cmd/Ctrl + Shift + O` | 当前文件符号列表 |
| 工作区符号 | `Cmd/Ctrl + T` | 搜索工作区符号 |

### 代码操作

| 功能 | 快捷键 | 说明 |
|-----|--------|------|
| 重命名 | `F2` | 重命名符号 |
| 快速修复 | `Cmd/Ctrl + .` | 显示可用的代码操作 |
| 添加导入 | 自动 | 保存时自动添加/删除 |
| 格式化 | `Shift + Alt + F` | 格式化当前文件 |
| 注释切换 | `Cmd/Ctrl + /` | 切换行注释 |

### 测试功能

在测试文件中：

- 测试函数上方会显示 "run test | debug test" 按钮
- 点击运行单个测试
- 测试结果显示在输出面板

```go
func TestAdd(t *testing.T) {  // <- 这里会显示 run test | debug test
    result := Add(1, 2)
    if result != 3 {
        t.Errorf("expected 3, got %d", result)
    }
}
```

### 命令面板

按 `Cmd/Ctrl + Shift + P` 打开命令面板，常用 Go 命令：

| 命令 | 功能 |
|-----|------|
| Go: Add Import | 添加导入 |
| Go: Add Tags To Struct Fields | 添加结构体标签 |
| Go: Remove Tags From Struct Fields | 移除结构体标签 |
| Go: Generate Interface Stubs | 生成接口实现 |
| Go: Generate Unit Tests | 生成单元测试 |
| Go: Fill Struct | 填充结构体字段 |
| Go: Run on Go Playground | 在 Playground 运行 |
| Go: Toggle Test Coverage | 显示测试覆盖率 |

---

## 快捷键

### 最常用快捷键

| 快捷键 (Mac) | 快捷键 (Win/Linux) | 功能 |
|-------------|-------------------|------|
| `Cmd + Shift + P` | `Ctrl + Shift + P` | 命令面板 |
| `Cmd + P` | `Ctrl + P` | 快速打开文件 |
| `Cmd + B` | `Ctrl + B` | 切换侧边栏 |
| `Cmd + J` | `Ctrl + J` | 切换终端 |
| `Cmd + \`` | `Ctrl + \`` | 打开终端 |
| `Cmd + Shift + E` | `Ctrl + Shift + E` | 文件浏览器 |
| `Cmd + Shift + F` | `Ctrl + Shift + F` | 全局搜索 |
| `Cmd + Shift + G` | `Ctrl + Shift + G` | Git 面板 |

### 编辑快捷键

| 快捷键 (Mac) | 快捷键 (Win/Linux) | 功能 |
|-------------|-------------------|------|
| `Cmd + D` | `Ctrl + D` | 选中下一个相同词 |
| `Cmd + Shift + L` | `Ctrl + Shift + L` | 选中所有相同词 |
| `Alt + Up/Down` | `Alt + Up/Down` | 移动当前行 |
| `Shift + Alt + Up/Down` | `Shift + Alt + Up/Down` | 复制当前行 |
| `Cmd + Shift + K` | `Ctrl + Shift + K` | 删除当前行 |
| `Cmd + /` | `Ctrl + /` | 切换注释 |

---

## 调试配置

### 创建调试配置

1. 点击左侧活动栏的 "运行和调试" 图标
2. 点击 "创建 launch.json 文件"
3. 选择 "Go" 环境

### launch.json 示例

在项目根目录创建 `.vscode/launch.json`：

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${fileDirname}"
        },
        {
            "name": "Launch Main",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/api"
        },
        {
            "name": "Debug Test",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${fileDirname}"
        },
        {
            "name": "Attach to Process",
            "type": "go",
            "request": "attach",
            "mode": "local",
            "processId": "${command:pickProcess}"
        }
    ]
}
```

### 调试操作

| 操作 | 快捷键 | 说明 |
|-----|--------|------|
| 开始调试 | `F5` | 启动调试 |
| 继续 | `F5` | 继续执行 |
| 单步跳过 | `F10` | 执行下一行 |
| 单步进入 | `F11` | 进入函数 |
| 单步跳出 | `Shift + F11` | 跳出函数 |
| 停止调试 | `Shift + F5` | 停止调试 |
| 切换断点 | `F9` | 在当前行切换断点 |

### 调试技巧

1. **条件断点**: 右键断点 > 编辑断点 > 添加条件
2. **日志点**: 右键断点 > 添加日志消息（不会暂停执行）
3. **监视表达式**: 在监视面板添加要观察的变量
4. **调试控制台**: 可以在调试时执行表达式

---

## 常见问题

### Q1: gopls 占用大量内存

**解决方案**: 调整 gopls 设置

```json
{
    "gopls": {
        "ui.completion.usePlaceholders": false,
        "ui.diagnostic.analyses": {
            "unusedparams": false
        }
    }
}
```

### Q2: 无法识别 go.work 工作区

**解决方案**: 确保使用 Go 1.18+ 并重启 VS Code

```bash
go version  # 确认版本 >= 1.18
# 重启 VS Code
```

### Q3: 自动导入不工作

**解决方案**:
1. 确保 gopls 正在运行（查看输出面板）
2. 运行 `Go: Restart Language Server`
3. 检查 `go.mod` 是否存在

### Q4: 代码提示很慢

**解决方案**:
```json
{
    "gopls": {
        "completionBudget": "200ms"
    }
}
```

### Q5: 测试无法运行

**解决方案**:
1. 确保文件名以 `_test.go` 结尾
2. 确保测试函数以 `Test` 开头
3. 确保函数签名为 `func TestXxx(t *testing.T)`

---

## 推荐扩展

除了 Go 官方扩展，以下扩展也很有用：

| 扩展 | 用途 |
|-----|------|
| Error Lens | 内联显示错误信息 |
| GitLens | 增强 Git 功能 |
| Docker | Docker 支持 |
| REST Client | 测试 HTTP API |
| Thunder Client | API 测试工具 |
| YAML | YAML 语法支持 |

---

## 下一步

- [第一个程序](./first-program.md) - 编写并运行 Hello World

---

## 参考资源

- [VS Code Go 扩展文档](https://github.com/golang/vscode-go/wiki)
- [gopls 文档](https://github.com/golang/tools/tree/master/gopls)
- [VS Code 快捷键参考](https://code.visualstudio.com/shortcuts/keyboard-shortcuts-macos.pdf)
