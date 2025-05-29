# Routing Package for Go

A lightweight HTTP routing library built on Go’s standard `net/http` package. This package gives you powerful route management driven by simple, expressive APIs. It supports dynamic URL parameters, query and JSON binding, global and group-specific middleware (with the option to skip certain middleware), and it now handles several HTTP methods (GET, POST, PUT, DELETE) with robust request handling.

## Features

- **Route Registration:**
  Easily register routes with HTTP methods like GET, POST, PUT, DELETE.

- **Middleware Support:**
  - Apply global middleware via `WithMiddleware`.
  - Chain and wrap middleware in the order of registration.
  - Remove middleware from specific routers using `WithoutNextMiddleware`.
  - On a per-route level, you can skip middleware using `WithoutMiddleware`.

- **Route Grouping:**
  Group routes under a common prefix with the `Group` function. Extra middleware added in a group wrap each grouped route automatically. This enables you to:
  - Inherit global middleware.
  - Add group-specific middleware.
  - Skip group middleware on individual routes.

- **Path Parameters:**
  Define dynamic segments using a colon (e.g., `/users/:id`) that extract URL parameters into a map.

- **Data Binding:**
  Built-in support for:
  - Query parameters binding (via `BindQuery`).
  - JSON payload binding (via `BindJSON`).

- **Path Normalization and Robust Matching:**
  Removes trailing slashes (except for the root) to improve matching consistency. Returns a proper 405 (Method Not Allowed) when the path matches but the HTTP method does not.

## Installation

Simply copy the `pkg/routing` directory into your project. Since this is a lightweight package with no external dependencies (aside from the Go standard library), you can integrate it directly into your codebase.

## Basic Usage

Below is a sample application that demonstrates how to create a router, register routes, add middleware, and group routes with middleware skipping.

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "your_module/pkg/routing" // Update this import path to match your project structure.
)

func main() {
    // Create a new router.
    r := routing.New()

    // Global middleware: log each incoming request.
    logger := func(next routing.HandlerFunc) routing.HandlerFunc {
        return func(ctx *routing.Context) {
            fmt.Println("Received request for", ctx.Request.URL.Path)
            next(ctx)
        }
    }
    r.WithMiddleware(logger)

    // Extra middleware examples.
    auth := func(next routing.HandlerFunc) routing.HandlerFunc {
        return func(ctx *routing.Context) {
            token := ctx.Request.Header.Get("Authorization")
            if token != "secret" {
                ctx.Text(http.StatusUnauthorized, "Unauthorized")
                return
            }
            next(ctx)
        }
    }

    metrics := func(next routing.HandlerFunc) routing.HandlerFunc {
        return func(ctx *routing.Context) {
            // (Imagine logging metrics here)
            ctx.Response.Header().Set("X-Metrics", "recorded")
            next(ctx)
        }
    }

    // Register basic routes.
    r.GET("/hello", func(ctx *routing.Context) {
        ctx.JSON(http.StatusOK, map[string]string{"message": "Hello, world!"})
    })

    // Dynamic route example: retrieves a user by ID.
    r.GET("/users/:id", func(ctx *routing.Context) {
        userID := ctx.Params["id"]
        ctx.Text(http.StatusOK, "User ID: "+userID)
    })

    // Binding example: parse query and JSON.
    r.GET("/search", func(ctx *routing.Context) {
        type SearchQuery struct {
            Term string `query:"term"`
        }
        var q SearchQuery
        if err := ctx.BindQuery(&q); err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
            return
        }
        ctx.JSON(http.StatusOK, map[string]string{"searched": q.Term})
    })

    r.POST("/register", func(ctx *routing.Context) {
        type User struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }
        var user User
        if err := ctx.BindJSON(&user); err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
            return
        }
        ctx.JSON(http.StatusOK, user)
    })

    // Group routes under "/api" with group-specific middleware.
    // Here we add an authentication middleware at the group level.
    // Also, we can skip a middleware for a specific route.
    r.WithMiddleware(auth)

    // Create a subgroup which skips the "metrics" middleware.
    r.WithoutNextMiddleware(metrics).Group("/api", func(api *routing.Router) {
        // This route retains global auth middleware.
        api.GET("/secure", func(ctx *routing.Context) {
            ctx.Text(http.StatusOK, "Secure API Content")
        })
        // This route explicitly skips the auth middleware.
        api.GET("/public", func(ctx *routing.Context) {
            ctx.Text(http.StatusOK, "Public API Content")
        }).WithoutMiddleware(auth)
    })

    // List all registered routes (useful for debugging)
    for _, route := range r.Routes() {
        fmt.Println("Registered route:", route)
    }

    fmt.Println("Server starting on port :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
