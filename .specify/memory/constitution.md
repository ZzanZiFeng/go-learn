<!--
## Sync Impact Report

**Version change**: N/A (initial) → 1.0.0
**Modified principles**: N/A (new constitution)
**Added sections**:
  - I. 渐进式学习 (Progressive Learning)
  - II. 实践驱动 (Practice-Driven)
  - III. 可验证性 (Verifiability)
  - IV. 问题解决导向 (Problem-Solving Oriented)
  - 学习路径 (Learning Path)
  - 内容规范 (Content Standards)
  - Governance
**Removed sections**: N/A
**Templates status**:
  - .specify/templates/plan-template.md: ✅ Compatible (no changes needed)
  - .specify/templates/spec-template.md: ✅ Compatible (no changes needed)
  - .specify/templates/tasks-template.md: ✅ Compatible (no changes needed)
**Follow-up TODOs**: None
-->

# Go语言初学者教程 Constitution

## Core Principles

### I. 渐进式学习 (Progressive Learning)

教程内容 MUST 遵循从基础到高级的渐进式结构：
- 每个模块 MUST 建立在前置知识之上，明确列出先决条件
- 复杂概念 MUST 分解为可消化的小单元
- 每个知识点 MUST 提供可运行的代码示例
- 学习曲线 SHOULD 平滑，避免跳跃式难度提升

**理由**：初学者需要稳固的基础才能构建更复杂的理解。跳跃式学习会导致知识断层。

### II. 实践驱动 (Practice-Driven)

所有教学内容 MUST 以实践为核心：
- 理论讲解后 MUST 紧跟动手练习
- 每章 MUST 包含至少一个完整的、可独立运行的项目
- 代码示例 MUST 是可编译、可执行的完整程序（非片段）
- 练习题 MUST 有明确的验收标准和参考答案

**理由**：编程是实践技能，纸上谈兵无法培养真正的编程能力。

### III. 可验证性 (Verifiability)

学习成果 MUST 可量化验证：
- 每个单元 MUST 提供自测检查点
- 所有代码示例 MUST 附带预期输出
- 调试章节 MUST 提供可复现的错误场景
- 部署章节 MUST 提供验证部署成功的检查清单

**理由**：学习者需要明确知道自己是否掌握了知识点，避免盲目自信或不必要的焦虑。

### IV. 问题解决导向 (Problem-Solving Oriented)

内容组织 MUST 围绕实际问题展开：
- 调试章节 MUST 从真实错误场景出发
- 部署章节 MUST 解决实际部署中的常见问题
- 每个高级主题 MUST 说明"为什么需要这个"和"什么场景下使用"
- SHOULD 提供"常见错误"和"最佳实践"对比

**理由**：初学者需要理解知识的应用场景，而非孤立地记忆语法。

## 学习路径 (Learning Path)

教程 MUST 覆盖以下四个维度，按顺序组织：

### 1. 开发基础 (Development Fundamentals)
- Go环境搭建与工具链
- 语法基础（变量、类型、控制流）
- 函数与包管理
- 数据结构（数组、切片、map、struct）
- 接口与错误处理
- 并发基础（goroutine、channel）

### 2. 实战项目 (Hands-on Projects)
- 命令行工具开发
- Web API服务开发
- 数据库操作
- 文件处理与I/O
- 第三方库集成

### 3. 部署实践 (Deployment)
- 代码编译与交叉编译
- Docker容器化
- 配置管理与环境变量
- 日志与监控基础
- CI/CD基础概念

### 4. 调试技能 (Debugging)
- 编译错误解读与修复
- 运行时错误排查
- 使用delve调试器
- 性能分析基础（pprof）
- 常见错误模式与解决方案

## 内容规范 (Content Standards)

### 代码示例规范

所有代码示例 MUST 遵循：
- 使用`go fmt`格式化
- 包含必要的注释说明
- 提供完整的包声明和import
- 标注Go版本要求（如适用）

### 文档格式规范

- 每章 MUST 以学习目标开始
- 每章 MUST 以总结和练习结束
- 复杂概念 MUST 配图说明
- 代码块 MUST 使用语法高亮

### 难度标记

内容 SHOULD 使用以下难度标记：
- 🟢 入门级：无需先决条件
- 🟡 进阶级：需要掌握基础知识
- 🔴 高级：需要较深入的Go经验

## Governance

### 修订流程

宪法修订 MUST 遵循以下流程：
1. 提出修订建议并说明理由
2. 评估对现有内容的影响
3. 更新版本号（语义化版本）
4. 记录修订历史

### 版本规则

- **MAJOR**: 核心原则变更或学习路径重组
- **MINOR**: 新增原则或显著扩展内容规范
- **PATCH**: 澄清、措辞调整、错误修正

### 合规检查

所有新增教程内容 MUST 在发布前验证：
- 符合渐进式学习原则
- 代码可编译运行
- 包含必要的练习和自测

**Version**: 1.0.0 | **Ratified**: 2025-11-30 | **Last Amended**: 2025-11-30
