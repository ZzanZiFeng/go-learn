// tests/benchmark/json_bench_test.go
// JSON 处理基准测试示例

package benchmark

import (
	"encoding/json"
	"testing"
)

// User 测试用户结构
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Active   bool   `json:"active"`
}

// UserList 用户列表
type UserList struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
}

var testUser = User{
	ID:       1,
	Username: "testuser",
	Email:    "test@example.com",
	Age:      25,
	Active:   true,
}

var testUserList = func() UserList {
	users := make([]User, 100)
	for i := 0; i < 100; i++ {
		users[i] = User{
			ID:       int64(i + 1),
			Username: "user",
			Email:    "user@example.com",
			Age:      25,
			Active:   true,
		}
	}
	return UserList{Users: users, Total: 100}
}()

// 基准测试: 单个对象序列化
func BenchmarkJSON_Marshal_SingleObject(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(testUser)
	}
}

// 基准测试: 单个对象反序列化
func BenchmarkJSON_Unmarshal_SingleObject(b *testing.B) {
	data, _ := json.Marshal(testUser)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var u User
		_ = json.Unmarshal(data, &u)
	}
}

// 基准测试: 列表序列化
func BenchmarkJSON_Marshal_List(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(testUserList)
	}
}

// 基准测试: 列表反序列化
func BenchmarkJSON_Unmarshal_List(b *testing.B) {
	data, _ := json.Marshal(testUserList)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var ul UserList
		_ = json.Unmarshal(data, &ul)
	}
}

// 比较: 预分配 vs 动态增长
func BenchmarkSlice_Append_Dynamic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result []int
		for j := 0; j < 1000; j++ {
			result = append(result, j)
		}
	}
}

func BenchmarkSlice_Append_Preallocated(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result := make([]int, 0, 1000)
		for j := 0; j < 1000; j++ {
			result = append(result, j)
		}
	}
}

// 子基准测试
func BenchmarkJSON_Operations(b *testing.B) {
	b.Run("Marshal/Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			json.Marshal(testUser)
		}
	})

	b.Run("Marshal/Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			json.Marshal(testUserList)
		}
	})

	data, _ := json.Marshal(testUser)
	b.Run("Unmarshal/Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var u User
			json.Unmarshal(data, &u)
		}
	})

	largeData, _ := json.Marshal(testUserList)
	b.Run("Unmarshal/Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var ul UserList
			json.Unmarshal(largeData, &ul)
		}
	})
}
