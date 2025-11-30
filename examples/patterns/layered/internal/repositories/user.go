// Package repositories handles data access
package repositories

import (
    "context"
    "errors"
    "sync"
    "time"

    "example/layered/internal/models"
)

var ErrNotFound = errors.New("not found")

// UserRepository defines the interface for user data access
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    FindByID(ctx context.Context, id uint) (*models.User, error)
    FindByEmail(ctx context.Context, email string) (*models.User, error)
    FindAll(ctx context.Context, offset, limit int) ([]*models.User, int, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id uint) error
}

// memoryRepository is an in-memory implementation for demo purposes
// In production, this would be a database-backed implementation
type memoryRepository struct {
    mu     sync.RWMutex
    users  map[uint]*models.User
    nextID uint
}

// NewUserRepository creates a new in-memory user repository
func NewUserRepository() UserRepository {
    return &memoryRepository{
        users:  make(map[uint]*models.User),
        nextID: 1,
    }
}

func (r *memoryRepository) Create(ctx context.Context, user *models.User) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    user.ID = r.nextID
    user.CreatedAt = time.Now()
    user.UpdatedAt = time.Now()
    r.nextID++

    r.users[user.ID] = user
    return nil
}

func (r *memoryRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    user, ok := r.users[id]
    if !ok {
        return nil, ErrNotFound
    }
    return user, nil
}

func (r *memoryRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    for _, user := range r.users {
        if user.Email == email {
            return user, nil
        }
    }
    return nil, ErrNotFound
}

func (r *memoryRepository) FindAll(ctx context.Context, offset, limit int) ([]*models.User, int, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    // Convert map to slice
    allUsers := make([]*models.User, 0, len(r.users))
    for _, user := range r.users {
        allUsers = append(allUsers, user)
    }

    total := len(allUsers)

    // Apply pagination
    if offset >= len(allUsers) {
        return []*models.User{}, total, nil
    }

    end := offset + limit
    if end > len(allUsers) {
        end = len(allUsers)
    }

    return allUsers[offset:end], total, nil
}

func (r *memoryRepository) Update(ctx context.Context, user *models.User) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, ok := r.users[user.ID]; !ok {
        return ErrNotFound
    }

    user.UpdatedAt = time.Now()
    r.users[user.ID] = user
    return nil
}

func (r *memoryRepository) Delete(ctx context.Context, id uint) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, ok := r.users[id]; !ok {
        return ErrNotFound
    }

    delete(r.users, id)
    return nil
}
