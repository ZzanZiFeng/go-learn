# Go 安装指南

本文档将指导你在不同操作系统上安装 Go 开发环境。

## 目录

- [系统要求](#系统要求)
- [macOS 安装](#macos-安装)
- [Windows 安装](#windows-安装)
- [Linux 安装](#linux-安装)
- [验证安装](#验证安装)
- [常见问题](#常见问题)

---

## 系统要求

Go 支持以下操作系统和架构：

| 操作系统 | 架构 | 最低版本 |
|---------|------|---------|
| macOS | amd64, arm64 | macOS 10.15+ |
| Windows | amd64 | Windows 10+ |
| Linux | amd64, arm64 | 内核 2.6.32+ |

**推荐版本**: Go 1.21 或更高版本

---

## macOS 安装

### 方法一：使用 Homebrew（推荐）

```bash
# 安装 Homebrew（如果尚未安装）
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安装 Go
brew install go

# 验证安装
go version
```

### 方法二：官方安装包

1. 访问 [Go 官方下载页面](https://go.dev/dl/)
2. 下载 macOS 安装包（`.pkg` 文件）
   - Intel Mac: `go1.21.x.darwin-amd64.pkg`
   - Apple Silicon: `go1.21.x.darwin-arm64.pkg`
3. 双击安装包，按照提示完成安装
4. 默认安装路径为 `/usr/local/go`

### 配置环境变量

```bash
# 编辑 ~/.zshrc 或 ~/.bash_profile
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.zshrc

# 重新加载配置
source ~/.zshrc
```

---

## Windows 安装

### 方法一：MSI 安装程序（推荐）

1. 访问 [Go 官方下载页面](https://go.dev/dl/)
2. 下载 Windows MSI 安装程序 `go1.21.x.windows-amd64.msi`
3. 运行安装程序，按照向导完成安装
4. 默认安装路径为 `C:\Go`

### 方法二：使用 Chocolatey

```powershell
# 安装 Chocolatey（如果尚未安装）
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# 安装 Go
choco install golang
```

### 方法三：使用 Scoop

```powershell
# 安装 Scoop（如果尚未安装）
irm get.scoop.sh | iex

# 安装 Go
scoop install go
```

### 配置环境变量

Windows 安装程序会自动配置环境变量。如需手动配置：

1. 打开"系统属性" > "高级" > "环境变量"
2. 在"系统变量"中添加/编辑：
   - `GOROOT`: `C:\Go`（Go 安装目录）
   - `GOPATH`: `%USERPROFILE%\go`（工作目录）
3. 在 `Path` 变量中添加：
   - `%GOROOT%\bin`
   - `%GOPATH%\bin`

---

## Linux 安装

### 方法一：官方二进制包（推荐）

```bash
# 下载最新版本（以 1.21.5 为例）
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# 删除旧版本（如果存在）
sudo rm -rf /usr/local/go

# 解压到 /usr/local
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc

# 重新加载配置
source ~/.bashrc

# 验证安装
go version
```

### 方法二：使用包管理器

#### Ubuntu/Debian

```bash
# 使用 apt（可能不是最新版本）
sudo apt update
sudo apt install golang-go

# 或使用 snap 获取最新版本
sudo snap install go --classic
```

#### Fedora

```bash
sudo dnf install golang
```

#### Arch Linux

```bash
sudo pacman -S go
```

### ARM64 架构（如树莓派）

```bash
# 下载 ARM64 版本
wget https://go.dev/dl/go1.21.5.linux-arm64.tar.gz

# 其余步骤同上
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-arm64.tar.gz
```

---

## 验证安装

安装完成后，打开新的终端窗口，运行以下命令：

```bash
# 检查 Go 版本
go version
# 预期输出: go version go1.21.x darwin/arm64 (或其他平台)

# 检查 Go 环境配置
go env
```

### 关键环境变量说明

```bash
go env GOROOT   # Go 安装目录
go env GOPATH   # Go 工作目录（存放第三方包和你的代码）
go env GOPROXY  # Go 模块代理
```

### 配置国内镜像（推荐）

由于网络原因，建议配置 Go 模块代理：

```bash
# 设置 Go 模块代理为国内镜像
go env -w GOPROXY=https://goproxy.cn,direct

# 或使用七牛云镜像
go env -w GOPROXY=https://goproxy.io,direct
```

---

## 常见问题

### Q1: `go: command not found`

**原因**: PATH 环境变量未正确配置

**解决方案**:
```bash
# macOS/Linux
export PATH=$PATH:/usr/local/go/bin
source ~/.bashrc  # 或 ~/.zshrc

# Windows: 检查环境变量设置
```

### Q2: 下载包速度慢或超时

**原因**: 默认代理在国内访问较慢

**解决方案**:
```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### Q3: 权限问题 (Linux/macOS)

**解决方案**:
```bash
# 确保 GOPATH 目录存在且有权限
mkdir -p $HOME/go
chmod 755 $HOME/go
```

### Q4: VS Code 提示 Go tools 安装失败

**解决方案**:
```bash
# 先配置代理
go env -w GOPROXY=https://goproxy.cn,direct

# 然后在 VS Code 中重新安装 Go tools
# Cmd/Ctrl + Shift + P > Go: Install/Update Tools
```

---

## 下一步

安装完成后，请继续学习：

- [GOPATH 与 Go Modules](./gopath-modules.md) - 理解 Go 的依赖管理
- [VS Code 配置](./vscode-setup.md) - 配置开发环境
- [第一个程序](./first-program.md) - 编写 Hello World

---

## 参考资源

- [Go 官方下载页面](https://go.dev/dl/)
- [Go 安装文档](https://go.dev/doc/install)
- [Go 环境配置](https://go.dev/doc/gopath_code)
