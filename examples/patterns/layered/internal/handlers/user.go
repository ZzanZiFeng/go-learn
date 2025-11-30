// Package handlers contains HTTP handlers
package handlers

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "example/layered/internal/services"
)

// UserHandler handles user HTTP requests
type UserHandler struct {
    service services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(service services.UserService) *UserHandler {
    return &UserHandler{service: service}
}

// CreateUserRequest is the request body for creating a user
type CreateUserRequest struct {
    Email string `json:"email"`
    Name  string `json:"name"`
}

// UserResponse is the response for user endpoints
type UserResponse struct {
    ID    uint   `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
}

// ErrorResponse is the response for errors
type ErrorResponse struct {
    Error string `json:"error"`
    Code  string `json:"code,omitempty"`
}

// Create handles POST /api/users
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
        return
    }

    user, err := h.service.CreateUser(r.Context(), services.CreateUserInput{
        Email: req.Email,
        Name:  req.Name,
    })
    if err != nil {
        handleError(w, err)
        return
    }

    respondJSON(w, http.StatusCreated, UserResponse{
        ID:    user.ID,
        Email: user.Email,
        Name:  user.Name,
    })
}

// GetByID handles GET /api/users/{id}
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
        return
    }

    user, err := h.service.GetUserByID(r.Context(), uint(id))
    if err != nil {
        handleError(w, err)
        return
    }

    respondJSON(w, http.StatusOK, UserResponse{
        ID:    user.ID,
        Email: user.Email,
        Name:  user.Name,
    })
}

// List handles GET /api/users
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

    if page == 0 {
        page = 1
    }
    if pageSize == 0 {
        pageSize = 10
    }

    users, total, err := h.service.ListUsers(r.Context(), page, pageSize)
    if err != nil {
        handleError(w, err)
        return
    }

    response := make([]UserResponse, len(users))
    for i, u := range users {
        response[i] = UserResponse{
            ID:    u.ID,
            Email: u.Email,
            Name:  u.Name,
        }
    }

    respondJSON(w, http.StatusOK, map[string]interface{}{
        "data":      response,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    })
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func handleError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, services.ErrUserNotFound):
        respondJSON(w, http.StatusNotFound, ErrorResponse{
            Error: "User not found",
            Code:  "USER_NOT_FOUND",
        })
    case errors.Is(err, services.ErrUserExists):
        respondJSON(w, http.StatusConflict, ErrorResponse{
            Error: "User already exists",
            Code:  "USER_EXISTS",
        })
    case errors.Is(err, services.ErrInvalidInput):
        respondJSON(w, http.StatusBadRequest, ErrorResponse{
            Error: "Invalid input",
            Code:  "INVALID_INPUT",
        })
    default:
        respondJSON(w, http.StatusInternalServerError, ErrorResponse{
            Error: "Internal server error",
            Code:  "INTERNAL_ERROR",
        })
    }
}
