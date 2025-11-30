// Package main demonstrates error handling in Go.
//
// To run this example:
//
//	go run main.go
//
// Expected output:
//
//	=== 基本错误处理 / Basic Error Handling ===
//	divide(10, 2) = 5.00
//	...
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

func main() {
	fmt.Println("=== 基本错误处理 / Basic Error Handling ===")

	// 基本错误检查
	// Basic error checking
	result, err := divide(10, 2)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("divide(10, 2) = %.2f\n", result)
	}

	result, err = divide(10, 0)
	if err != nil {
		fmt.Println("divide(10, 0) Error:", err)
	}

	fmt.Println("\n=== 创建错误 / Creating Errors ===")

	// 使用 errors.New
	// Using errors.New
	err1 := errors.New("something went wrong")
	fmt.Println("errors.New:", err1)

	// 使用 fmt.Errorf
	// Using fmt.Errorf
	userID := 42
	err2 := fmt.Errorf("user %d not found", userID)
	fmt.Println("fmt.Errorf:", err2)

	fmt.Println("\n=== 自定义错误类型 / Custom Error Type ===")

	err3 := validateUser("", 25)
	if err3 != nil {
		fmt.Println("Validation error:", err3)

		// 类型断言获取详细信息
		// Type assertion for details
		if ve, ok := err3.(*ValidationError); ok {
			fmt.Printf("  Field: %s, Message: %s\n", ve.Field, ve.Message)
		}
	}

	err4 := validateUser("Alice", 200)
	if err4 != nil {
		fmt.Println("Validation error:", err4)
	}

	fmt.Println("\n=== 错误包装 / Error Wrapping ===")

	_, err5 := readConfig("nonexistent.json")
	if err5 != nil {
		fmt.Println("Error:", err5)

		// 检查底层错误
		// Check underlying error
		if errors.Is(err5, os.ErrNotExist) {
			fmt.Println("  -> File does not exist")
		}
	}

	fmt.Println("\n=== 预定义错误 / Predefined Errors ===")

	_, err6 := getResource(404)
	if err6 != nil {
		switch {
		case errors.Is(err6, ErrNotFound):
			fmt.Println("Resource not found")
		case errors.Is(err6, ErrInvalidInput):
			fmt.Println("Invalid input")
		default:
			fmt.Println("Unknown error:", err6)
		}
	}

	fmt.Println("\n=== 多重错误检查 / Multiple Error Checks ===")

	// 链式错误处理示例
	// Chained error handling example
	if err := processUser(123); err != nil {
		fmt.Println("Process failed:", err)
	}

	fmt.Println("\n=== 类型转换错误 / Type Conversion Errors ===")

	// strconv 错误
	// strconv errors
	_, err7 := strconv.Atoi("not a number")
	if err7 != nil {
		fmt.Println("strconv error:", err7)

		// 获取详细错误信息
		// Get detailed error info
		var numErr *strconv.NumError
		if errors.As(err7, &numErr) {
			fmt.Printf("  Func: %s, Num: %s, Err: %v\n",
				numErr.Func, numErr.Num, numErr.Err)
		}
	}

	fmt.Println("\n=== panic 和 recover / panic and recover ===")

	// panic/recover 示例
	// panic/recover example
	safeDivide(10, 2)
	safeDivide(10, 0) // 会 panic 但被 recover

	fmt.Println("Program continues after recover...")

	// defer 中的 recover
	// recover in defer
	result2 := safeOperation()
	fmt.Println("safeOperation result:", result2)

	fmt.Println("\n=== 错误处理最佳实践 / Error Handling Best Practices ===")

	// 演示最佳实践
	// Demonstrate best practices
	if err := demonstrateBestPractices(); err != nil {
		fmt.Println("Operation failed:", err)
	}
}

// 基本除法函数
// Basic division function
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// 自定义错误类型
// Custom error type
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on '%s': %s", e.Field, e.Message)
}

func validateUser(name string, age int) error {
	if name == "" {
		return &ValidationError{Field: "name", Message: "cannot be empty"}
	}
	if age < 0 || age > 150 {
		return &ValidationError{Field: "age", Message: "must be between 0 and 150"}
	}
	return nil
}

// 错误包装
// Error wrapping
func readConfig(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config '%s': %w", filename, err)
	}
	return data, nil
}

// 预定义错误
// Predefined errors
var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
)

func getResource(id int) (string, error) {
	if id <= 0 {
		return "", ErrInvalidInput
	}
	if id == 404 {
		return "", ErrNotFound
	}
	return fmt.Sprintf("Resource-%d", id), nil
}

// 链式错误处理
// Chained error handling
func processUser(userID int) error {
	user, err := fetchUser(userID)
	if err != nil {
		return fmt.Errorf("process user: %w", err)
	}

	if err := validateUserData(user); err != nil {
		return fmt.Errorf("process user: %w", err)
	}

	if err := saveUser(user); err != nil {
		return fmt.Errorf("process user: %w", err)
	}

	fmt.Println("User processed successfully:", user)
	return nil
}

func fetchUser(id int) (string, error) {
	return fmt.Sprintf("User%d", id), nil
}

func validateUserData(user string) error {
	return nil
}

func saveUser(user string) error {
	return nil
}

// panic/recover 示例
// panic/recover example
func safeDivide(a, b int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	result := a / b // 如果 b=0 会 panic
	fmt.Printf("%d / %d = %d\n", a, b, result)
}

func safeOperation() (result int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
			result = -1
		}
	}()

	panic("intentional panic")
}

// 最佳实践示例
// Best practices example
func demonstrateBestPractices() error {
	// 1. 总是检查错误
	// 1. Always check errors
	data, err := getData()
	if err != nil {
		return fmt.Errorf("get data: %w", err) // 2. 添加上下文
	}
	_ = data

	// 3. 早返回
	// 3. Early return
	if err := validateData(data); err != nil {
		return err
	}

	// 4. 不要忽略错误（除非有明确理由）
	// 4. Don't ignore errors (unless there's a clear reason)
	result, err := processData(data)
	if err != nil {
		// 5. 日志记录但继续（如果适当）
		// 5. Log and continue (if appropriate)
		fmt.Println("Warning: process data failed:", err)
	}
	_ = result

	fmt.Println("Best practices demonstrated successfully")
	return nil
}

func getData() (string, error) {
	return "data", nil
}

func validateData(data string) error {
	return nil
}

func processData(data string) (string, error) {
	return "processed", nil
}
