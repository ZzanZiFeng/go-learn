# 贡献指南

感谢你对 Go 后端学习教程的关注！我们欢迎任何形式的贡献，包括但不限于：

- 修复错别字或语法错误
- 改进代码示例
- 添加新的章节或示例
- 提出问题和建议

## 快速开始

### 1. Fork 仓库

点击页面右上角的 "Fork" 按钮，将仓库复制到你的 GitHub 账号下。

### 2. 克隆到本地

```bash
git clone https://github.com/YOUR_USERNAME/go-learn.git
cd go-learn
```

### 3. 创建分支

```bash
git checkout -b feature/your-feature-name
# 或
git checkout -b fix/your-fix-name
```

### 4. 进行修改

按照下面的规范进行修改。

### 5. 提交更改

```bash
git add .
git commit -m "类型: 简短描述"
git push origin feature/your-feature-name
```

### 6. 创建 Pull Request

在 GitHub 上创建 Pull Request，描述你的更改内容。

## 项目结构

```
go-learn/
├── docs/                         # 四轨教程文档
│   ├── 00-introduction/          # 总体介绍
│   ├── 01-development/           # 轨道一：开发教程
│   ├── 02-practice/              # 轨道二：实战教程
│   ├── 03-deployment/            # 轨道三：部署教程
│   └── 04-testing/               # 轨道四：测试教程
├── projects/                     # 实战项目源码
│   ├── todo-cli/
│   ├── todo-api/
│   ├── auth-service/
│   └── fullstack-demo/
├── tests/                        # 测试示例与验证
├── examples/                     # 独立代码示例
├── infra/                        # 基础设施配置
└── specs/                        # 设计规范文档
```

## 贡献规范

### 文档编写规范

#### 文件命名

- 使用小写字母和连字符：`01-variables.md`
- 章节文件以数字序号开头：`01-`, `02-`, ...
- README.md 作为章节入口

#### 文档结构

每个章节文档应包含：

```markdown
# 章节标题

> 简短描述

## 本章目标

完成本章后，你将能够：
- 目标 1
- 目标 2

## 核心内容

### 主题 1

内容...

### 主题 2

内容...

## JavaScript/TypeScript 对比

| 特性 | Go | JavaScript/TypeScript |
|------|-----|----------------------|
| ... | ... | ... |

## 练习

### 练习 1: 标题

**目标**: 描述
**要求**: 列表

## 小结

本章要点...

## 下一步

- [下一章节](./link)
```

#### 代码示例规范

1. **必须可运行**: 所有代码示例都应该能独立运行
2. **包含预期输出**: 使用注释标注预期输出

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
    // Output: Hello, World!
}
```

3. **添加清晰注释**: 解释关键代码

```go
// 创建一个带缓冲的 channel
// 容量为 3，意味着可以发送 3 个值而不阻塞
ch := make(chan int, 3)
```

4. **提供 JS/TS 对比** (适用时):

```go
// Go: 使用 defer 确保资源释放
file, err := os.Open("file.txt")
if err != nil {
    return err
}
defer file.Close()

// 对比 TypeScript:
// try {
//     const file = await fs.open("file.txt");
//     // 使用 file...
// } finally {
//     await file.close();
// }
```

### 代码示例规范

#### 目录结构

```
examples/<category>/<name>/
├── main.go          # 主程序
├── go.mod           # 模块文件（如需要独立依赖）
└── README.md        # 示例说明（可选）
```

#### 示例要求

1. **独立可运行**: 每个示例应该能独立运行
2. **清晰的入口**: 使用 `main.go` 作为入口文件
3. **适当的注释**: 解释代码的目的和关键步骤
4. **预期输出**: 在注释中包含预期输出

#### 示例模板

```go
// examples/category/name/main.go
// 示例名称: 简短描述
// 相关章节: DEV-XX
//
// 这个示例演示了...
//
// 运行方式:
//   go run main.go
//
// 预期输出:
//   Line 1 of output
//   Line 2 of output

package main

import (
    "fmt"
)

func main() {
    // 示例代码
    fmt.Println("Hello, World!")
}
```

### Git 提交规范

使用以下格式：

```
类型: 简短描述

详细描述（可选）
```

#### 类型

- `feat`: 新功能或新章节
- `fix`: 修复错误
- `docs`: 文档修改
- `style`: 格式调整（不影响代码逻辑）
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 其他杂项

#### 示例

```
docs: 添加 DEV-03 并发编程章节

- 添加 goroutine 基础教程
- 添加 channel 通信示例
- 添加与 async/await 的对比
```

```
fix: 修复 examples/syntax/slices 示例中的错误

切片追加时使用了错误的变量名
```

### Pull Request 规范

#### 标题

```
[轨道] 类型: 简短描述
```

示例：
- `[DEV] feat: 添加 GORM 关联关系章节`
- `[PRAC] fix: 修复 todo-api 认证逻辑`
- `[DEPLOY] docs: 补充 Docker 多阶段构建说明`

#### 描述模板

```markdown
## 变更内容

简要描述这个 PR 做了什么。

## 变更类型

- [ ] 新功能
- [ ] Bug 修复
- [ ] 文档更新
- [ ] 代码重构
- [ ] 其他

## 检查清单

- [ ] 代码可以正常运行
- [ ] 已添加必要的注释
- [ ] 文档格式符合规范
- [ ] 已更新相关索引（如 examples/README.md）

## 相关 Issue

Fixes #123
```

## 内容质量要求

### 教程文档

1. **准确性**: 确保技术内容正确
2. **完整性**: 包含完整的代码示例
3. **一致性**: 遵循现有的格式和风格
4. **可读性**: 使用清晰的语言和结构

### 代码示例

1. **可运行**: 所有代码必须能正常运行
2. **简洁**: 聚焦于演示的概念
3. **注释**: 解释关键代码
4. **安全**: 不包含硬编码的敏感信息

### JavaScript/TypeScript 对比

对比内容应该：

1. **准确**: 正确反映两种语言的特性
2. **公平**: 不偏向任何一方
3. **实用**: 帮助读者理解概念

## 四轨结构说明

### 轨道一：开发教程 (docs/01-development/)

- 12 个章节，涵盖 Go 语言和后端开发知识
- 每个章节约 4-6 小时学习内容
- 必须包含 JS/TS 对比

### 轨道二：实战教程 (docs/02-practice/)

- 4 个完整项目
- 从入门到高级难度递进
- 项目代码在 projects/ 目录

### 轨道三：部署教程 (docs/03-deployment/)

- 5 个章节
- 涵盖编译、Docker、CI/CD、Kubernetes
- 实用的生产部署指南

### 轨道四：测试教程 (docs/04-testing/)

- 5 个章节
- 涵盖单元测试、集成测试、性能测试
- 测试代码在 tests/ 目录

## 本地开发

### 环境要求

- Go 1.21+
- Docker Desktop
- VS Code（推荐）

### 启动基础设施

```bash
# 启动所有服务
docker-compose -f infra/docker-compose.yml up -d

# 验证服务
docker-compose -f infra/docker-compose.yml ps
```

### 运行示例

```bash
# 运行单个示例
cd examples/syntax/hello
go run main.go

# 运行测试
cd tests/unit
go test -v ./...
```

### 文档预览

推荐使用 VS Code 的 Markdown 预览功能，或使用其他 Markdown 预览工具。

## 获取帮助

如果你有任何问题：

1. 查看现有的 [Issues](https://github.com/your-username/go-learn/issues)
2. 创建新的 Issue 描述你的问题
3. 在 PR 中 @ 维护者寻求帮助

## 行为准则

请保持友善和尊重。我们致力于为所有人提供一个包容、友好的环境。

## 许可证

通过贡献代码，你同意你的贡献将按照 MIT 许可证授权。
