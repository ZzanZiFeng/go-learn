# 复合类型

本文档介绍 Go 的复合类型：数组(array)、切片(slice)和映射(map)。

## 目录

- [数组](#数组)
- [切片](#切片)
- [映射](#映射)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 数组

### 数组基础

数组是固定长度、同类型元素的序列。

```go
package main

import "fmt"

func main() {
    // 声明数组（元素初始化为零值）
    var arr1 [5]int
    fmt.Println("Zero array:", arr1)  // [0 0 0 0 0]

    // 声明并初始化
    arr2 := [5]int{1, 2, 3, 4, 5}
    fmt.Println("Initialized:", arr2)

    // 部分初始化
    arr3 := [5]int{1, 2}
    fmt.Println("Partial:", arr3)  // [1 2 0 0 0]

    // 让编译器计算长度
    arr4 := [...]int{1, 2, 3}
    fmt.Println("Auto length:", arr4, "len:", len(arr4))

    // 指定索引初始化
    arr5 := [5]int{1: 10, 3: 30}
    fmt.Println("By index:", arr5)  // [0 10 0 30 0]
}
```

### 数组操作

```go
package main

import "fmt"

func main() {
    arr := [5]int{10, 20, 30, 40, 50}

    // 访问元素
    fmt.Println("First:", arr[0])
    fmt.Println("Last:", arr[len(arr)-1])

    // 修改元素
    arr[0] = 100
    fmt.Println("Modified:", arr)

    // 遍历数组
    fmt.Println("Range loop:")
    for i, v := range arr {
        fmt.Printf("  arr[%d] = %d\n", i, v)
    }

    // 只需要值
    for _, v := range arr {
        fmt.Print(v, " ")
    }
    fmt.Println()

    // 只需要索引
    for i := range arr {
        fmt.Print(i, " ")
    }
    fmt.Println()
}
```

### 数组是值类型

**重要**: Go 的数组是值类型，赋值和传参会复制整个数组。

```go
package main

import "fmt"

func modifyArray(arr [3]int) {
    arr[0] = 999
    fmt.Println("Inside function:", arr)
}

func main() {
    arr1 := [3]int{1, 2, 3}
    arr2 := arr1  // 复制

    arr2[0] = 100
    fmt.Println("arr1:", arr1)  // [1 2 3] - 未改变
    fmt.Println("arr2:", arr2)  // [100 2 3]

    // 传参也是复制
    modifyArray(arr1)            // Inside function: [999 2 3]
    fmt.Println("arr1:", arr1)  // [1 2 3] - 仍未改变
}
```

---

## 切片

### 切片基础

切片是动态长度的序列，是对底层数组的引用。这是 Go 中最常用的集合类型。

```go
package main

import "fmt"

func main() {
    // 使用 make 创建切片
    s1 := make([]int, 5)      // 长度5，容量5
    s2 := make([]int, 5, 10)  // 长度5，容量10

    fmt.Printf("s1: %v, len=%d, cap=%d\n", s1, len(s1), cap(s1))
    fmt.Printf("s2: %v, len=%d, cap=%d\n", s2, len(s2), cap(s2))

    // 字面量创建
    s3 := []int{1, 2, 3, 4, 5}
    fmt.Println("Literal:", s3)

    // 从数组创建切片
    arr := [5]int{10, 20, 30, 40, 50}
    s4 := arr[1:4]  // [20, 30, 40]
    fmt.Println("From array:", s4)

    // nil 切片
    var s5 []int
    fmt.Println("nil slice:", s5, "is nil:", s5 == nil)
}
```

### 切片操作

```go
package main

import "fmt"

func main() {
    // append - 追加元素
    s := []int{1, 2, 3}
    s = append(s, 4)
    s = append(s, 5, 6, 7)
    fmt.Println("After append:", s)

    // 追加另一个切片
    s2 := []int{8, 9}
    s = append(s, s2...)
    fmt.Println("Append slice:", s)

    // 切片操作
    fmt.Println("s[2:5]:", s[2:5])   // 从索引2到4
    fmt.Println("s[:3]:", s[:3])     // 前3个
    fmt.Println("s[5:]:", s[5:])     // 从索引5开始
    fmt.Println("s[:]:", s[:])       // 整个切片

    // copy - 复制切片
    src := []int{1, 2, 3}
    dst := make([]int, len(src))
    copied := copy(dst, src)
    fmt.Printf("Copied %d elements: %v\n", copied, dst)
}
```

### 切片是引用类型

切片变量存储的是对底层数组的引用，修改会影响原数据。

```go
package main

import "fmt"

func modifySlice(s []int) {
    s[0] = 999
    fmt.Println("Inside function:", s)
}

func main() {
    s1 := []int{1, 2, 3}
    s2 := s1  // 引用同一底层数组

    s2[0] = 100
    fmt.Println("s1:", s1)  // [100 2 3] - 也改变了！
    fmt.Println("s2:", s2)  // [100 2 3]

    // 传参传的是引用
    modifySlice(s1)          // Inside function: [999 2 3]
    fmt.Println("s1:", s1)  // [999 2 3] - 改变了！

    // 如果需要独立副本
    s3 := make([]int, len(s1))
    copy(s3, s1)
    s3[0] = 111
    fmt.Println("s1:", s1)  // [999 2 3] - 未改变
    fmt.Println("s3:", s3)  // [111 2 3]
}
```

### 切片容量与扩容

```go
package main

import "fmt"

func main() {
    s := make([]int, 0, 5)

    for i := 0; i < 10; i++ {
        s = append(s, i)
        fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
    }
}
```

输出:
```
len=1 cap=5 [0]
len=2 cap=5 [0 1]
len=3 cap=5 [0 1 2]
len=4 cap=5 [0 1 2 3]
len=5 cap=5 [0 1 2 3 4]
len=6 cap=10 [0 1 2 3 4 5]      // 容量翻倍
len=7 cap=10 [0 1 2 3 4 5 6]
...
```

### 删除切片元素

```go
package main

import "fmt"

func main() {
    s := []int{1, 2, 3, 4, 5}

    // 删除索引 i 的元素
    i := 2
    s = append(s[:i], s[i+1:]...)
    fmt.Println("After remove index 2:", s)  // [1 2 4 5]

    // 删除第一个元素
    s = s[1:]
    fmt.Println("After remove first:", s)  // [2 4 5]

    // 删除最后一个元素
    s = s[:len(s)-1]
    fmt.Println("After remove last:", s)  // [2 4]
}
```

---

## 映射

### Map 基础

Map 是键值对的无序集合，类似于 JavaScript 的对象或 Map。

```go
package main

import "fmt"

func main() {
    // 使用 make 创建
    m1 := make(map[string]int)
    m1["one"] = 1
    m1["two"] = 2
    fmt.Println("m1:", m1)

    // 字面量创建
    m2 := map[string]int{
        "one":   1,
        "two":   2,
        "three": 3,
    }
    fmt.Println("m2:", m2)

    // nil map（只读）
    var m3 map[string]int
    fmt.Println("nil map:", m3, "is nil:", m3 == nil)
    // m3["key"] = 1  // panic: 不能写入 nil map
}
```

### Map 操作

```go
package main

import "fmt"

func main() {
    m := map[string]int{
        "apple":  5,
        "banana": 3,
        "orange": 7,
    }

    // 读取值
    fmt.Println("apple:", m["apple"])

    // 检查键是否存在
    value, exists := m["grape"]
    if exists {
        fmt.Println("grape:", value)
    } else {
        fmt.Println("grape not found")
    }

    // 简写：只检查存在性
    if _, ok := m["apple"]; ok {
        fmt.Println("apple exists")
    }

    // 添加/修改
    m["grape"] = 10
    m["apple"] = 15
    fmt.Println("After modify:", m)

    // 删除
    delete(m, "banana")
    fmt.Println("After delete:", m)

    // 获取长度
    fmt.Println("Length:", len(m))

    // 遍历（顺序不确定）
    for key, value := range m {
        fmt.Printf("%s: %d\n", key, value)
    }
}
```

### Map 是引用类型

```go
package main

import "fmt"

func modifyMap(m map[string]int) {
    m["modified"] = 999
}

func main() {
    m1 := map[string]int{"a": 1, "b": 2}
    m2 := m1  // 引用同一个 map

    m2["c"] = 3
    fmt.Println("m1:", m1)  // map[a:1 b:2 c:3]
    fmt.Println("m2:", m2)  // map[a:1 b:2 c:3]

    modifyMap(m1)
    fmt.Println("After func:", m1)  // 包含 modified:999
}
```

### Map 作为集合

```go
package main

import "fmt"

func main() {
    // 使用 map[T]bool 或 map[T]struct{} 作为集合
    set := make(map[string]struct{})

    // 添加元素
    set["apple"] = struct{}{}
    set["banana"] = struct{}{}
    set["orange"] = struct{}{}

    // 检查元素是否存在
    if _, exists := set["apple"]; exists {
        fmt.Println("apple is in set")
    }

    // 删除元素
    delete(set, "banana")

    // 遍历
    for item := range set {
        fmt.Println(item)
    }
}
```

---

## 与 JavaScript/TypeScript 对比

### 数组对比

```javascript
// JavaScript - 动态数组
let arr = [1, 2, 3];
arr.push(4);           // 可以添加
arr[10] = 11;          // 可以稀疏
console.log(arr);      // [1, 2, 3, 4, empty × 6, 11]
console.log(arr.length); // 11
```

```go
// Go 数组 - 固定长度
arr := [3]int{1, 2, 3}
// arr[3] = 4  // 编译错误: 索引越界
// 不能改变长度

// Go 切片 - 动态长度
slice := []int{1, 2, 3}
slice = append(slice, 4)
// slice[10] = 11  // 运行时 panic: 索引越界
```

### 切片 vs JavaScript Array

| 操作 | JavaScript Array | Go Slice |
|-----|-----------------|----------|
| 创建 | `[]`, `new Array()` | `[]T{}`, `make([]T, len)` |
| 添加元素 | `push()` | `append()` (需重新赋值) |
| 删除元素 | `splice()` | 切片操作 |
| 长度 | `.length` | `len()` |
| 查找 | `indexOf()`, `find()` | 手动遍历或 `slices.Index()` |
| 过滤 | `filter()` | 手动遍历 |
| 映射 | `map()` | 手动遍历 |

```javascript
// JavaScript 链式操作
const result = [1, 2, 3, 4, 5]
    .filter(x => x > 2)
    .map(x => x * 2);
// [6, 8, 10]
```

```go
// Go 需要手动实现
nums := []int{1, 2, 3, 4, 5}
var result []int
for _, n := range nums {
    if n > 2 {
        result = append(result, n*2)
    }
}
// [6 8 10]
```

### Map 对比

```javascript
// JavaScript Object
const obj = { name: "Go", year: 2009 };
obj.version = "1.21";
delete obj.year;

// JavaScript Map
const map = new Map();
map.set("name", "Go");
map.get("name");
map.has("name");
map.delete("name");
```

```go
// Go map
m := map[string]interface{}{
    "name": "Go",
    "year": 2009,
}
m["version"] = "1.21"
delete(m, "year")

// 检查键存在
if value, ok := m["name"]; ok {
    fmt.Println(value)
}
```

### 主要差异总结

| 特性 | JavaScript | Go |
|-----|-----------|-----|
| 数组长度 | 动态 | 固定 |
| 类似动态数组 | Array | Slice |
| 数组类型 | 可混合类型 | 单一类型 |
| 稀疏数组 | 支持 | 不支持 |
| 数组方法 | 丰富的内置方法 | 标准库函数 |
| Map 键类型 | 任意 | 可比较类型 |
| 空值处理 | `undefined` | 零值 |
| 引用/值 | 都是引用 | 数组是值，切片是引用 |

### 重要陷阱

```javascript
// JavaScript - 空值是 undefined
const obj = {};
console.log(obj.name);  // undefined
if (obj.name) { }       // false (falsy)
```

```go
// Go - 不存在的键返回零值
m := map[string]int{}
fmt.Println(m["name"])  // 0 (int 的零值)
if m["name"] != 0 { }   // 但 0 可能是有效值！

// 正确做法：检查键是否存在
if value, ok := m["name"]; ok {
    fmt.Println(value)
}
```

---

## 最佳实践

1. **优先使用切片而非数组**: 除非长度确定不变
2. **预分配切片容量**: 当知道大概大小时，使用 `make([]T, 0, capacity)`
3. **nil 切片是安全的**: 可以 `append`、`len`、`range`
4. **nil map 不安全**: 写入会 panic，必须初始化
5. **检查 map 键存在**: 使用 `, ok` 模式

```go
// 好的实践
// 预分配提高性能
items := make([]string, 0, 100)
for _, item := range data {
    items = append(items, item)
}

// nil 切片是安全的
var s []int
s = append(s, 1, 2, 3)  // OK

// map 必须初始化
m := make(map[string]int)  // 或 map[string]int{}
m["key"] = value
```

---

## 下一步

- [控制流](./04-control-flow.md) - 学习 if、for、switch 语句
