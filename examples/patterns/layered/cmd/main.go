// Package main demonstrates layered architecture in Go
package main

import (
    "example/layered/internal/handlers"
    "example/layered/internal/repositories"
    "example/layered/internal/services"
    "fmt"
    "log"
    "net/http"
)

func main() {
    fmt.Println("=== Layered Architecture Demo ===")

    // Initialize dependencies (Dependency Injection)
    // In a real app, you would use a database
    repo := repositories.NewUserRepository()
    svc := services.NewUserService(repo)
    handler := handlers.NewUserHandler(svc)

    // Setup routes
    mux := http.NewServeMux()

    // User endpoints
    mux.HandleFunc("POST /api/users", handler.Create)
    mux.HandleFunc("GET /api/users/{id}", handler.GetByID)
    mux.HandleFunc("GET /api/users", handler.List)

    // Health check
    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"status":"ok"}`))
    })

    fmt.Println("Server starting on :8080")
    fmt.Println("\nEndpoints:")
    fmt.Println("  POST /api/users          - Create user")
    fmt.Println("  GET  /api/users/{id}     - Get user by ID")
    fmt.Println("  GET  /api/users          - List users")
    fmt.Println("  GET  /health             - Health check")

    log.Fatal(http.ListenAndServe(":8080", mux))
}
