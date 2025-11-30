# 结构体与接口 (DEV-02)

本章将帮助你掌握 Go 语言的结构体、方法和接口，理解 Go 独特的面向对象设计。

## 学习目标

完成本章后，你将能够：

- 定义和使用结构体
- 理解结构体标签的用途
- 编写值接收者和指针接收者方法
- 定义和实现接口
- 使用类型断言和类型开关
- 使用结构体嵌入实现组合
- 使用 Go 1.18+ 的泛型基础

## 章节内容

| 文档 | 说明 | 预计时间 |
|------|------|----------|
| [结构体基础](./01-structs.md) | 结构体定义和初始化 | 30min |
| [结构体标签](./02-struct-tags.md) | json, db 等标签的使用 | 20min |
| [方法](./03-methods.md) | 值接收者与指针接收者 | 30min |
| [接口](./04-interfaces.md) | 接口定义和隐式实现 | 35min |
| [类型断言](./05-type-assertion.md) | 类型断言和类型开关 | 25min |
| [嵌入与组合](./06-embedding.md) | 结构体嵌入（组合优于继承） | 30min |
| [泛型基础](./07-generics.md) | Go 1.18+ 泛型入门 | 30min |

## 前置要求

- 已完成 [语法基础](../01-syntax/) 章节
- 理解函数和指针的基本概念

## 难度

🟡 **进阶** - 需要理解 Go 独特的设计哲学

## 核心概念

### Go 的面向对象特点

Go 不是传统的面向对象语言，没有类和继承，但通过结构体和接口实现了更灵活的设计：

| 传统 OOP | Go |
|----------|-----|
| 类 (Class) | 结构体 (Struct) |
| 继承 (Inheritance) | 嵌入/组合 (Embedding/Composition) |
| 多态 (Polymorphism) | 接口 (Interface) |
| 构造函数 (Constructor) | 工厂函数 (Factory Function) |
| this/self | 接收者 (Receiver) |

### 与 JavaScript/TypeScript 对比

| 概念 | TypeScript | Go |
|-----|-----------|-----|
| 类型定义 | `class`, `interface`, `type` | `struct`, `interface` |
| 继承 | `extends` | 嵌入 (Embedding) |
| 实现接口 | `implements` (显式) | 隐式实现 |
| 访问控制 | `public`, `private`, `protected` | 首字母大小写 |
| 泛型 | `<T>` | `[T any]` |

## 验收标准

完成本章后，请确认你能够：

- [ ] 定义结构体并使用多种方式初始化
- [ ] 使用 json 标签进行 JSON 序列化
- [ ] 编写指针接收者方法修改结构体
- [ ] 定义接口并实现
- [ ] 使用类型断言安全地转换类型
- [ ] 使用嵌入实现代码复用

## 配套代码

本章的示例代码位于 `examples/syntax/` 目录：

```
examples/syntax/
├── structs/        # 结构体示例
├── methods/        # 方法示例
├── interfaces/     # 接口示例
└── embedding/      # 嵌入示例
```

---

下一章: [并发编程](../03-concurrency/)
