# 语法基础 (DEV-01)

本章将帮助你掌握 Go 语言的基本语法，包括变量、类型、控制流、函数和错误处理。

## 学习目标

完成本章后，你将能够：

- 声明和使用变量与常量
- 理解 Go 的基本类型和复合类型
- 使用 if、for、switch 控制流程
- 编写和调用函数
- 理解指针的基本概念
- 处理错误和异常
- 组织和导入包

## 章节内容

| 文档 | 说明 | 预计时间 |
|------|------|----------|
| [变量与常量](./01-variables.md) | var, :=, const 的使用 | 30min |
| [基本类型](./02-basic-types.md) | int, string, bool, float64 | 25min |
| [复合类型](./03-composite-types.md) | array, slice, map | 40min |
| [控制流](./04-control-flow.md) | if, for, switch | 30min |
| [函数](./05-functions.md) | 参数、返回值、多返回值 | 35min |
| [指针](./06-pointers.md) | 指针基础与使用场景 | 30min |
| [错误处理](./07-error-handling.md) | error, panic, recover | 35min |
| [包管理](./08-packages.md) | import, go mod | 25min |
| [JS/TS对比总结](./js-comparison.md) | Go 与 JavaScript/TypeScript 差异 | 20min |

## 前置要求

- 已完成 [环境搭建](../00-environment/) 章节
- Go 开发环境已配置完成
- VS Code 可以运行 Go 程序

## 难度

🟢 **入门** - 适合有其他编程语言基础的学习者

## 对 JavaScript/TypeScript 开发者的提示

如果你来自 JS/TS 背景，请特别注意以下差异：

| 概念 | JavaScript/TypeScript | Go |
|-----|----------------------|-----|
| 变量声明 | `let`, `const`, `var` | `var`, `:=`, `const` |
| 类型 | 动态/静态可选 | 静态强类型 |
| 数组 | 动态长度 | 固定长度 |
| 类似数组的动态结构 | Array | slice |
| 循环 | `for`, `while`, `for...of` | 只有 `for` |
| 错误处理 | `try/catch` | 返回 `error` |

## 验收标准

完成本章后，请确认你能够：

- [ ] 使用 `var` 和 `:=` 声明变量
- [ ] 使用 `const` 定义常量
- [ ] 使用切片(slice)和映射(map)
- [ ] 编写 `for` 循环遍历数据
- [ ] 编写多返回值函数
- [ ] 正确处理函数返回的错误

## 配套代码

本章的示例代码位于 `examples/syntax/` 目录：

```
examples/syntax/
├── hello/          # Hello World 示例
├── variables/      # 变量示例
├── types/          # 类型示例
├── slices/         # 切片示例
├── maps/           # Map 示例
├── control/        # 控制流示例
├── functions/      # 函数示例
├── pointers/       # 指针示例
└── errors/         # 错误处理示例
```

---

下一章: [结构体与接口](../02-struct-interface/)
