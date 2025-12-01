// Package storage 提供待办事项的持久化存储
package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/user/todo-cli/internal/todo"
)

// Storage 定义存储接口
type Storage interface {
	Load() ([]*todo.Todo, error)
	Save(todos []*todo.Todo) error
}

// JSONStorage 使用 JSON 文件存储待办事项
type JSONStorage struct {
	filepath string
}

// NewJSONStorage 创建 JSON 存储实例
func NewJSONStorage(filepath string) *JSONStorage {
	return &JSONStorage{
		filepath: filepath,
	}
}

// DefaultJSONStorage 使用默认路径创建存储
func DefaultJSONStorage() (*JSONStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	fp := filepath.Join(homeDir, ".todos.json")
	return NewJSONStorage(fp), nil
}

// Load 从 JSON 文件加载待办列表
func (s *JSONStorage) Load() ([]*todo.Todo, error) {
	if _, err := os.Stat(s.filepath); errors.Is(err, os.ErrNotExist) {
		return []*todo.Todo{}, nil
	}

	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []*todo.Todo{}, nil
	}

	var todos []*todo.Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, err
	}

	return todos, nil
}

// Save 保存待办列表到 JSON 文件
func (s *JSONStorage) Save(todos []*todo.Todo) error {
	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filepath, data, 0644)
}

// GetFilePath 返回存储文件路径
func (s *JSONStorage) GetFilePath() string {
	return s.filepath
}
