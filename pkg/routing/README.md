# Routing Package for Go

A lightweight HTTP routing library built on Go’s standard `net/http` package. This package gives you powerful route management driven by simple, expressive APIs. It supports dynamic URL parameters, query and JSON binding, global and group‑specific middleware (with options to skip certain middleware), and robust request handling across multiple HTTP methods (GET, POST, PUT, DELETE). The library now supports Laravel‑style route grouping with dedicated prefixes and nested groups.

## Features

- **Route Registration:**
  - Easily register routes with HTTP methods like GET, POST, PUT, DELETE.
  - Define dynamic segments (e.g., `/users/:id`) that extract URL parameters.

- **Middleware Support:**
  - Apply global middleware via `WithMiddleware`.
  - Chain and wrap middleware in the order of registration.
  - Exclude specific middleware on a per-route basis using `WithoutMiddleware`.
  - Remove middleware for an entire router (or group) using `WithoutNextMiddleware`.

- **Route Grouping and Prefixing:**
  - Group routes under a common prefix using the new `Prefix` method. Instead of passing a prefix directly to `Group`, you now call `Prefix` to yield a sub‑router with an updated path.
  - Define a group with group‑specific middleware. Any middleware added after calling `Prefix` is applied only to that group.
  - Supports the following syntax, similar to Laravel:
    ```go
    // Create a group with a prefix:
    r.Prefix("/admin").Group(func(admin *routing.Router) {
        admin.GET("/dashboard", dashboardHandler) // Matches "/admin/dashboard"
    })

    // Nested groups are supported:
    r.Prefix("/api").Group(func(api *routing.Router) {
        api.Prefix("/v1").Group(func(v1 *routing.Router) {
            v1.GET("/resource", resourceHandler) // Matches "/api/v1/resource"
        })
    })

    // Middleware can be excluded on specific routes:
    r.Prefix("/blog").Group(func(blog *routing.Router) {
        blog.GET("/post", postHandler).WithoutMiddleware(authMiddleware)
    })
    ```
  - Extra middleware added in a group wrap only the routes in that group. The underlying implementation ensures that group‑specific middleware is applied only once (preventing double wrapping).

- **Data Binding:**
  - Built‑in support for query parameter binding (using `BindQuery`).
  - Built‑in JSON payload binding (using `BindJSON`) for request bodies.

- **Routing Robustness:**
  - Path normalization automatically removes trailing slashes (except for the root).
  - Provides proper 405 (Method Not Allowed) responses when paths match but the method does not.
  - Offers comprehensive dynamic route matching.

## Installation

Simply copy the `pkg/routing` directory into your project. Since this is a lightweight package with no external dependencies (aside from Go’s standard library), you can integrate it directly into your codebase.

## Basic Usage

Below is a sample application demonstrating how to create a router, register routes, add middleware, use the new prefix‑based grouping, nest groups, and skip middleware when needed.

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "your_module/pkg/routing" // Update this import path as needed.
)

func main() {
    // Create a new router.
    r := routing.New()

    // Global middleware example:
    logger := func(next routing.HandlerFunc) routing.HandlerFunc {
        return func(ctx *routing.Context) {
            fmt.Println("Received request for", ctx.Request.URL.Path)
            next(ctx)
        }
    }
    r.WithMiddleware(logger)

    // Basic route registration.
    r.GET("/hello", func(ctx *routing.Context) {
        ctx.JSON(http.StatusOK, map[string]string{"message": "Hello, world!"})
    })

    // Dynamic route with URL parameter.
    r.GET("/users/:id", func(ctx *routing.Context) {
        userID := ctx.Params["id"]
        ctx.Text(http.StatusOK, "User ID: " + userID)
    })

    // Data binding: Query and JSON.
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

    // Extra middleware examples.
    authMiddleware := func(next routing.HandlerFunc) routing.HandlerFunc {
        return func(ctx *routing.Context) {
            token := ctx.Request.Header.Get("Authorization")
            if token != "secret" {
                ctx.Text(http.StatusUnauthorized, "Unauthorized")
                return
            }
            next(ctx)
        }
    }

    metricsMiddleware := func(next routing.HandlerFunc) routing.HandlerFunc {
        return func(ctx *routing.Context) {
            // (Imagine logging metrics here, e.g., incrementing counters)
            ctx.Response.Header().Set("X-Metrics", "recorded")
            next(ctx)
        }
    }

    // Global routes can continue to be registered as normal.
    r.GET("/public", func(ctx *routing.Context) {
        ctx.Text(http.StatusOK, "Public content")
    })

    // Group routes under "/api" while applying group-specific middleware.
    // Global auth middleware can be added and later selectively skipped.
    r.WithMiddleware(authMiddleware)

    // Create a prefixed group. Using the new Prefix method allows you to clearly separate the prefix.
    r.Prefix("/api").Group(func(api *routing.Router) {
        // This "/api/secure" route inherits the global auth middleware.
        api.GET("/secure", func(ctx *routing.Context) {
            ctx.Text(http.StatusOK, "Secure API Content")
        })
        // This "/api/public" route explicitly skips auth middleware.
        api.GET("/public", func(ctx *routing.Context) {
            ctx.Text(http.StatusOK, "Public API Content")
        }).WithoutMiddleware(authMiddleware)
    })

    // Nested group example.
    r.Prefix("/admin").Group(func(admin *routing.Router) {
        admin.GET("/dashboard", func(ctx *routing.Context) {
            ctx.Text(http.StatusOK, "Admin Dashboard")
        })
        // Nested group under "/admin":
        admin.Prefix("/settings").Group(func(settings *routing.Router) {
            settings.GET("/profile", func(ctx *routing.Context) {
                ctx.Text(http.StatusOK, "Profile Settings")
            })
        })
    })

    // List all registered routes (useful for debugging).
    for _, route := range r.Routes() {
        fmt.Println("Registered route:", route)
    }

    fmt.Println("Server starting on port :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
```

## Additional Considerations
- __Middleware Wrapping:__ Group‑specific middleware is applied only once. If no extra middleware is added in a group, the global middleware chain continues to wrap the route during request handling. You can remove unwanted middleware from a group using `WithoutNextMiddleware`.
- __Nested Groups:__ When groups are nested, the prefixes are concatenated automatically. For example, a group with prefix `/api` nested with another group with prefix `/v1` will yield routes under `/api/v1`.
- __Skipping Middleware:__ On a per-route basis, you can use .`WithoutMiddleware(mw)` to exclude specific middleware regardless of whether they were added globally or at the group level.