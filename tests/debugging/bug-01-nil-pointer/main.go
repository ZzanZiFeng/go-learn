// tests/debugging/bug-01-nil-pointer/main.go
// Bug 场景 1: Nil 指针引用
// 这是一个常见的 Go 编程错误示例

package main

import "fmt"

// User 用户结构
type User struct {
	ID   int64
	Name string
}

// UserService 用户服务
type UserService struct {
	users map[int64]*User
}

func NewUserService() *UserService {
	return &UserService{
		users: make(map[int64]*User),
	}
}

func (s *UserService) GetUser(id int64) *User {
	return s.users[id] // 可能返回 nil
}

func (s *UserService) AddUser(user *User) {
	s.users[user.ID] = user
}

// ========================================
// 问题代码
// ========================================
func buggyCode() {
	service := NewUserService()

	// 尝试获取不存在的用户
	user := service.GetUser(999)

	// 问题: 没有检查 nil，直接访问字段
	fmt.Printf("User name: %s\n", user.Name) // panic: nil pointer dereference
}

// ========================================
// 修复后的代码
// ========================================
func fixedCode() {
	service := NewUserService()

	// 尝试获取用户
	user := service.GetUser(999)

	// 修复 1: 检查 nil
	if user == nil {
		fmt.Println("User not found")
		return
	}

	fmt.Printf("User name: %s\n", user.Name)
}

// 修复 2: 返回 error
func (s *UserService) GetUserSafe(id int64) (*User, error) {
	user, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user %d not found", id)
	}
	return user, nil
}

func fixedCodeWithError() {
	service := NewUserService()

	user, err := service.GetUserSafe(999)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("User name: %s\n", user.Name)
}

// ========================================
// 调试说明
// ========================================
/*
运行 buggyCode() 会产生以下 panic:

panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x8 pc=0x...]

goroutine 1 [running]:
main.buggyCode()
        /path/main.go:35 +0x...
main.main()
        /path/main.go:... +0x...

调试步骤:
1. 查看 panic 信息中的文件和行号
2. 检查该行访问的指针是否可能为 nil
3. 追溯指针的来源
4. 添加 nil 检查或使用 error 返回值

使用 delve 调试:
$ dlv debug main.go
(dlv) break main.go:35
(dlv) continue
(dlv) print user
nil
*/

func main() {
	// 取消注释以查看 bug
	// buggyCode()

	// 修复后的代码
	fixedCode()
	fixedCodeWithError()
}
