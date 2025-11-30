# DEV-01 练习与解答

本文档包含语法基础章节的练习题和参考解答。

## 练习 1: 变量与类型

### 题目

编写一个程序，完成以下任务：

1. 声明一个整数变量 `age`，使用短变量声明赋值为 25
2. 声明一个字符串变量 `name`，使用 var 声明赋值为你的名字
3. 声明一个布尔变量 `isStudent`，使用类型推断
4. 使用 `fmt.Printf` 打印所有变量的类型和值

### 解答

```go
package main

import "fmt"

func main() {
    // 1. 短变量声明
    age := 25

    // 2. var 声明
    var name string = "Gopher"

    // 3. 类型推断
    var isStudent = true

    // 4. 打印类型和值
    fmt.Printf("age: type=%T, value=%v\n", age, age)
    fmt.Printf("name: type=%T, value=%v\n", name, name)
    fmt.Printf("isStudent: type=%T, value=%v\n", isStudent, isStudent)
}
```

---

## 练习 2: 切片操作

### 题目

1. 创建一个包含 1-10 的整数切片
2. 使用切片操作获取前 5 个元素
3. 使用切片操作获取后 3 个元素
4. 删除索引为 4 的元素
5. 在末尾添加 100, 200, 300

### 解答

```go
package main

import "fmt"

func main() {
    // 1. 创建切片
    nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    fmt.Println("原始切片:", nums)

    // 2. 前 5 个元素
    first5 := nums[:5]
    fmt.Println("前5个:", first5)

    // 3. 后 3 个元素
    last3 := nums[len(nums)-3:]
    fmt.Println("后3个:", last3)

    // 4. 删除索引 4 的元素
    i := 4
    nums = append(nums[:i], nums[i+1:]...)
    fmt.Println("删除索引4后:", nums)

    // 5. 添加元素
    nums = append(nums, 100, 200, 300)
    fmt.Println("添加后:", nums)
}
```

---

## 练习 3: Map 统计

### 题目

给定一个字符串切片，统计每个单词出现的次数，并找出出现次数最多的单词。

```go
words := []string{"go", "is", "awesome", "go", "is", "fun", "go", "rocks"}
```

### 解答

```go
package main

import "fmt"

func main() {
    words := []string{"go", "is", "awesome", "go", "is", "fun", "go", "rocks"}

    // 统计词频
    wordCount := make(map[string]int)
    for _, word := range words {
        wordCount[word]++
    }
    fmt.Println("词频统计:", wordCount)

    // 找出最高频词
    maxWord := ""
    maxCount := 0
    for word, count := range wordCount {
        if count > maxCount {
            maxCount = count
            maxWord = word
        }
    }
    fmt.Printf("最高频词: %s (出现 %d 次)\n", maxWord, maxCount)
}
```

---

## 练习 4: FizzBuzz

### 题目

编写经典的 FizzBuzz 程序：

- 打印 1 到 30 的数字
- 如果数字能被 3 整除，打印 "Fizz"
- 如果数字能被 5 整除，打印 "Buzz"
- 如果同时能被 3 和 5 整除，打印 "FizzBuzz"
- 否则打印数字本身

### 解答

```go
package main

import "fmt"

func main() {
    for i := 1; i <= 30; i++ {
        switch {
        case i%15 == 0:
            fmt.Println("FizzBuzz")
        case i%3 == 0:
            fmt.Println("Fizz")
        case i%5 == 0:
            fmt.Println("Buzz")
        default:
            fmt.Println(i)
        }
    }
}
```

---

## 练习 5: 多返回值函数

### 题目

编写一个函数 `minMax`，接收一个整数切片，返回最小值和最大值。如果切片为空，返回错误。

### 解答

```go
package main

import (
    "errors"
    "fmt"
)

func minMax(nums []int) (min, max int, err error) {
    if len(nums) == 0 {
        return 0, 0, errors.New("slice is empty")
    }

    min, max = nums[0], nums[0]
    for _, n := range nums[1:] {
        if n < min {
            min = n
        }
        if n > max {
            max = n
        }
    }
    return min, max, nil
}

func main() {
    nums := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5}

    min, max, err := minMax(nums)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Min: %d, Max: %d\n", min, max)

    // 测试空切片
    _, _, err = minMax([]int{})
    if err != nil {
        fmt.Println("Empty slice error:", err)
    }
}
```

---

## 练习 6: 闭包计数器

### 题目

创建一个闭包工厂函数 `makeAccumulator`，返回一个累加器函数。每次调用累加器时，传入一个数字，返回到目前为止所有数字的总和。

### 解答

```go
package main

import "fmt"

func makeAccumulator() func(int) int {
    sum := 0
    return func(n int) int {
        sum += n
        return sum
    }
}

func main() {
    acc := makeAccumulator()

    fmt.Println(acc(1))   // 1
    fmt.Println(acc(5))   // 6
    fmt.Println(acc(10))  // 16
    fmt.Println(acc(3))   // 19

    // 新的累加器实例
    acc2 := makeAccumulator()
    fmt.Println(acc2(100))  // 100
    fmt.Println(acc(7))     // 26 (原来的继续累加)
}
```

---

## 练习 7: 指针交换

### 题目

编写一个函数 `swap`，使用指针交换两个整数的值。

### 解答

```go
package main

import "fmt"

func swap(a, b *int) {
    *a, *b = *b, *a
}

func main() {
    x, y := 10, 20
    fmt.Printf("交换前: x=%d, y=%d\n", x, y)

    swap(&x, &y)
    fmt.Printf("交换后: x=%d, y=%d\n", x, y)
}
```

---

## 练习 8: 错误处理

### 题目

编写一个函数 `parsePositiveInt`，将字符串解析为正整数。处理以下错误情况：

1. 字符串为空
2. 字符串不是有效数字
3. 数字为负数或零

### 解答

```go
package main

import (
    "errors"
    "fmt"
    "strconv"
)

var (
    ErrEmptyString    = errors.New("empty string")
    ErrNotPositive    = errors.New("number must be positive")
)

func parsePositiveInt(s string) (int, error) {
    if s == "" {
        return 0, ErrEmptyString
    }

    n, err := strconv.Atoi(s)
    if err != nil {
        return 0, fmt.Errorf("invalid number '%s': %w", s, err)
    }

    if n <= 0 {
        return 0, ErrNotPositive
    }

    return n, nil
}

func main() {
    testCases := []string{"42", "", "abc", "-5", "0", "100"}

    for _, tc := range testCases {
        result, err := parsePositiveInt(tc)
        if err != nil {
            fmt.Printf("parsePositiveInt(%q) error: %v\n", tc, err)
        } else {
            fmt.Printf("parsePositiveInt(%q) = %d\n", tc, result)
        }
    }
}
```

---

## 练习 9: 字符串处理

### 题目

编写一个函数 `countChars`，统计字符串中每个字符（rune）出现的次数。正确处理中文等 UTF-8 字符。

### 解答

```go
package main

import "fmt"

func countChars(s string) map[rune]int {
    counts := make(map[rune]int)
    for _, r := range s {
        counts[r]++
    }
    return counts
}

func main() {
    text := "Hello, 世界! Hello!"
    counts := countChars(text)

    fmt.Println("字符统计:")
    for char, count := range counts {
        fmt.Printf("  '%c': %d\n", char, count)
    }
}
```

---

## 练习 10: 综合练习 - 简单计算器

### 题目

编写一个简单的计算器程序：

1. 定义一个 `Calculator` 结构体，包含 `result` 字段
2. 实现 `Add`, `Subtract`, `Multiply`, `Divide` 方法
3. `Divide` 方法需要处理除零错误
4. 实现 `Reset` 方法重置结果
5. 实现链式调用

### 解答

```go
package main

import (
    "errors"
    "fmt"
)

type Calculator struct {
    result float64
    err    error
}

func NewCalculator(initial float64) *Calculator {
    return &Calculator{result: initial}
}

func (c *Calculator) Add(n float64) *Calculator {
    if c.err != nil {
        return c
    }
    c.result += n
    return c
}

func (c *Calculator) Subtract(n float64) *Calculator {
    if c.err != nil {
        return c
    }
    c.result -= n
    return c
}

func (c *Calculator) Multiply(n float64) *Calculator {
    if c.err != nil {
        return c
    }
    c.result *= n
    return c
}

func (c *Calculator) Divide(n float64) *Calculator {
    if c.err != nil {
        return c
    }
    if n == 0 {
        c.err = errors.New("division by zero")
        return c
    }
    c.result /= n
    return c
}

func (c *Calculator) Reset() *Calculator {
    c.result = 0
    c.err = nil
    return c
}

func (c *Calculator) Result() (float64, error) {
    return c.result, c.err
}

func main() {
    // 基本使用
    calc := NewCalculator(10)
    result, err := calc.Add(5).Multiply(2).Subtract(10).Result()
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Result:", result)  // (10 + 5) * 2 - 10 = 20
    }

    // 链式调用
    result2, _ := NewCalculator(100).
        Divide(2).
        Add(50).
        Multiply(2).
        Result()
    fmt.Println("Result2:", result2)  // (100 / 2 + 50) * 2 = 200

    // 错误处理
    calc3 := NewCalculator(10)
    result3, err3 := calc3.Divide(0).Add(5).Result()
    if err3 != nil {
        fmt.Println("Error:", err3)
    } else {
        fmt.Println("Result3:", result3)
    }
}
```

---

## 自测清单

完成所有练习后，确认你能够：

- [ ] 使用 `var` 和 `:=` 声明变量
- [ ] 使用基本类型和类型转换
- [ ] 创建和操作切片
- [ ] 使用 map 存储和查询数据
- [ ] 使用 for 和 range 循环
- [ ] 使用 switch 语句
- [ ] 编写多返回值函数
- [ ] 创建和使用闭包
- [ ] 使用指针修改值
- [ ] 正确处理错误

---

## 下一步

完成这些练习后，你已经掌握了 Go 的基本语法。继续学习：

- [结构体与接口](../02-struct-interface/) - 学习 Go 的面向对象特性
