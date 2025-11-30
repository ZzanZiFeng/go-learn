// Package services contains business logic
package services

import (
    "context"
    "errors"

    "example/layered/internal/models"
    "example/layered/internal/repositories"
)

// Business errors
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserExists   = errors.New("user already exists")
    ErrInvalidInput = errors.New("invalid input")
)

// UserService defines the interface for user business logic
type UserService interface {
    CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error)
    GetUserByID(ctx context.Context, id uint) (*models.User, error)
    ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int, error)
}

// CreateUserInput is the input for creating a user
type CreateUserInput struct {
    Email string
    Name  string
}

type userService struct {
    repo repositories.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo repositories.UserRepository) UserService {
    return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error) {
    // Business validation
    if input.Email == "" {
        return nil, ErrInvalidInput
    }
    if input.Name == "" {
        return nil, ErrInvalidInput
    }

    // Check if user exists
    existing, err := s.repo.FindByEmail(ctx, input.Email)
    if err == nil && existing != nil {
        return nil, ErrUserExists
    }

    // Create user
    user := &models.User{
        Email: input.Email,
        Name:  input.Name,
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, repositories.ErrNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    return user, nil
}

func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int, error) {
    // Business rules for pagination
    if pageSize > 100 {
        pageSize = 100
    }
    if pageSize < 1 {
        pageSize = 10
    }
    if page < 1 {
        page = 1
    }

    offset := (page - 1) * pageSize
    return s.repo.FindAll(ctx, offset, pageSize)
}
