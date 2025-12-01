// tests/unit/calculator_test.go
// 计算器单元测试示例 - 展示表驱动测试

package unit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Calculator 简单计算器
type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Add(a, b int) int {
	return a + b
}

func (c *Calculator) Subtract(a, b int) int {
	return a - b
}

func (c *Calculator) Multiply(a, b int) int {
	return a * b
}

func (c *Calculator) Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// 基本测试
func TestCalculator_Add(t *testing.T) {
	calc := NewCalculator()
	result := calc.Add(2, 3)
	if result != 5 {
		t.Errorf("Add(2, 3) = %d; want 5", result)
	}
}

// 表驱动测试
func TestCalculator_Add_TableDriven(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 2, 3, 5},
		{"negative numbers", -2, -3, -5},
		{"mixed numbers", -2, 3, 1},
		{"zeros", 0, 0, 0},
		{"large numbers", 1000000, 2000000, 3000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d; want %d",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// 使用 testify 断言
func TestCalculator_Subtract_Testify(t *testing.T) {
	calc := NewCalculator()

	assert.Equal(t, 2, calc.Subtract(5, 3), "5 - 3 should equal 2")
	assert.Equal(t, -2, calc.Subtract(3, 5), "3 - 5 should equal -2")
	assert.Equal(t, 0, calc.Subtract(5, 5), "5 - 5 should equal 0")
}

// 测试错误场景
func TestCalculator_Divide(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name      string
		a, b      int
		expected  int
		expectErr bool
	}{
		{"normal division", 10, 2, 5, false},
		{"division by zero", 10, 0, 0, true},
		{"negative division", -10, 2, -5, false},
		{"integer division", 7, 3, 2, false}, // 整数除法
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.Divide(tt.a, tt.b)

			if tt.expectErr {
				require.Error(t, err, "expected an error")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 并行测试
func TestCalculator_Parallel(t *testing.T) {
	calc := NewCalculator()

	t.Run("Add", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 5, calc.Add(2, 3))
	})

	t.Run("Subtract", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 2, calc.Subtract(5, 3))
	})

	t.Run("Multiply", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 15, calc.Multiply(3, 5))
	})
}
