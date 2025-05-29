package routing_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/martin3zra/acme/pkg/routing"
)

// -----------------
// MatchRoute Tests
// -----------------

func TestMatchRoute(t *testing.T) {
	// Table-driven tests for the MatchRoute function.
	tests := []struct {
		name    string
		pattern string
		path    string
		match   bool
		params  map[string]string
	}{
		{
			name:    "Static route match",
			pattern: "/users",
			path:    "/users",
			match:   true,
			params:  map[string]string{},
		},
		{
			name:    "Static route match with trailing slash in path",
			pattern: "/users",
			path:    "/users/",
			match:   true,
			params:  map[string]string{},
		},
		{
			name:    "Dynamic parameter match",
			pattern: "/users/:id",
			path:    "/users/123",
			match:   true,
			params:  map[string]string{"id": "123"},
		},
		{
			name:    "Dynamic parameter missing segment",
			pattern: "/users/:id",
			path:    "/users/",
			match:   false,
			params:  nil,
		},
		{
			name:    "Multiple dynamic parameters",
			pattern: "/users/:id/messages/:msgId",
			path:    "/users/45/messages/789",
			match:   true,
			params:  map[string]string{"id": "45", "msgId": "789"},
		},
		{
			name:    "Partial dynamic route match fails",
			pattern: "/users/:id/messages/:msgId",
			path:    "/users/45/messages",
			match:   false,
			params:  nil,
		},
		{
			name:    "Static route mismatch",
			pattern: "/about",
			path:    "/contact",
			match:   false,
			params:  nil,
		},
		{
			name:    "Root route match",
			pattern: "/",
			path:    "/",
			match:   true,
			params:  map[string]string{},
		},
		{
			name:    "Empty pattern and path match",
			pattern: "",
			path:    "",
			match:   true,
			params:  map[string]string{},
		},
		{
			name:    "Extra path segments should fail",
			pattern: "/users/:id",
			path:    "/users/123/extra",
			match:   false,
			params:  nil,
		},
		{
			name:    "Trailing slash in pattern",
			pattern: "/users/:id/",
			path:    "/users/123",
			match:   true,
			params:  map[string]string{"id": "123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, gotParams := routing.MatchRoute(tt.pattern, tt.path)
			if gotMatch != tt.match {
				t.Errorf("MatchRoute(%q, %q) match = %v; want %v", tt.pattern, tt.path, gotMatch, tt.match)
			}
			if !reflect.DeepEqual(gotParams, tt.params) {
				t.Errorf("MatchRoute(%q, %q) params = %v; want %v", tt.pattern, tt.path, gotParams, tt.params)
			}
		})
	}
}

// ---------------------
// Middleware Chain Tests
// ---------------------

// Dummy middleware m1 writes "m1-" to the response.
var m1 routing.Middleware = func(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		ctx.Response.Write([]byte("m1-"))
		next(ctx)
	}
}

// Dummy middleware m2 writes "m2-" to the response.
var m2 routing.Middleware = func(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		ctx.Response.Write([]byte("m2-"))
		next(ctx)
	}
}

// finalHandler writes "handler" to the response.
func finalHandler(ctx *routing.Context) {
	ctx.Response.Write([]byte("handler"))
}

// TestMiddlewareChain verifies that when multiple middleware are added,
// they are executed in the proper order. The expected order is based
// on the middleware wrapping logic.
func TestMiddlewareChain(t *testing.T) {
	r := routing.New()
	// Add two middleware functions: m1 and then m2.
	r.WithMiddleware(m1, m2)
	// Register a GET route that uses finalHandler.
	r.GET("/test", finalHandler)

	// Create a GET request for "/test".
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	resp := rr.Result()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %v", err)
	}
	body := string(bodyBytes)

	// With middleware wrapping in reverse order, the expected output is "m1-m2-handler".
	expected := "m1-m2-handler"
	if body != expected {
		t.Errorf("Expected response %q, got %q", expected, body)
	}
}

// TestWithoutNextMiddleware verifies that removing a specific middleware works as expected.
// In this test, we remove m1 so that only m2 is applied before the final handler.
func TestWithoutNextMiddleware(t *testing.T) {
	r := routing.New()
	// Add both m1 and m2.
	r.WithMiddleware(m1, m2)

	// Create a new router instance that excludes m1.
	r.GET("/test2", finalHandler).WithoutMiddleware(m1)

	req := httptest.NewRequest("GET", "/test2", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	resp := rr.Result()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %v", err)
	}
	body := string(bodyBytes)

	// With m1 removed, only m2 should wrap the handler, resulting in "m2-handler".
	expected := "m2-handler"
	if body != expected {
		t.Errorf("Expected response %q, got %q", expected, body)
	}
}

// testMiddleware sets the "X-Test-Middleware" header in the response.
var testMiddleware routing.Middleware = func(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		ctx.Response.Header().Set("X-Test-Middleware", "yes")
		next(ctx)
	}
}

// skipMiddleware sets the "X-Skip-Middleware" header in the response.
var skipMiddleware routing.Middleware = func(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		ctx.Response.Header().Set("X-Skip-Middleware", "yes")
		next(ctx)
	}
}

// TestGroupMiddlewareAndSkipNext verifies that when grouping routes using
// WithoutNextMiddleware the skipMiddleware is removed while testMiddleware remains.
// We capture the router returned by WithoutNextMiddleware and use it for serving the request.
func TestGroupMiddlewareAndSkipNext(t *testing.T) {
	// Create the global router and add both middlewares.
	r := routing.New()
	r.WithMiddleware(testMiddleware, skipMiddleware)

	// Remove skipMiddleware by creating a new router instance and work with it.
	r2 := r.WithoutNextMiddleware(skipMiddleware)

	// Group routes under "/api" on the new router.
	r2.Group("/api", func(api *routing.Router) {
		api.GET("/grouped", func(ctx *routing.Context) {
			// Return a basic text response.
			ctx.Text(http.StatusOK, "grouped")
		})
	})

	// Create a simulated HTTP GET request for "/api/grouped".
	req := httptest.NewRequest("GET", "/api/grouped", nil)
	w := httptest.NewRecorder()
	// Serve the request using the router that has had skipMiddleware removed.
	r2.ServeHTTP(w, req)

	// Check that the header set by testMiddleware is present.
	if w.Header().Get("X-Test-Middleware") != "yes" {
		t.Errorf("expected X-Test-Middleware header to be set to 'yes'")
	}
	// Check that the header set by skipMiddleware is absent.
	if w.Header().Get("X-Skip-Middleware") != "" {
		t.Errorf("expected X-Skip-Middleware header to be skipped due to group configuration")
	}
}

// TestWithRequest_Success verifies that a valid JSON request
// is correctly decoded by WithRequest and processed by the handler.
func TestWithRequest_Success(t *testing.T) {
	// Define a sample payload structure.
	type RequestPayload struct {
		Message string `json:"message"`
	}

	// Create a new router instance.
	r := routing.New()

	// Register a POST route using WithRequest.
	// The handler simply echoes back the payload message as JSON.
	r.POST("/echo", routing.WithRequest(func(ctx *routing.Context, payload *RequestPayload) {
		ctx.JSON(http.StatusOK, map[string]string{
			"echo": payload.Message,
		})
	}))

	// Build a request with valid JSON.
	validJSON := `{"message": "Hello, world!"}`
	req := httptest.NewRequest("POST", "/echo", strings.NewReader(validJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Serve the request.
	r.ServeHTTP(rr, req)

	// Validate the response.
	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var res map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("error unmarshalling response: %v", err)
	}
	if res["echo"] != "Hello, world!" {
		t.Errorf("expected echo %q, got %q", "Hello, world!", res["echo"])
	}
}

// TestWithRequest_InvalidJSON verifies that an invalid JSON request
// returns an appropriate error response.
func TestWithRequest_InvalidJSON(t *testing.T) {
	// Define a sample payload structure.
	type RequestPayload struct {
		Message string `json:"message"`
	}

	// Create a new router instance.
	r := routing.New()
	// Register a POST route using WithRequest.
	r.POST("/echo", routing.WithRequest(func(ctx *routing.Context, payload *RequestPayload) {
		ctx.JSON(http.StatusOK, map[string]string{"echo": payload.Message})
	}))

	// Create a request with invalid JSON.
	invalidJSON := `{"message": "Hello, world!`
	req := httptest.NewRequest("POST", "/echo", strings.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Serve the request.
	r.ServeHTTP(rr, req)

	// Expect a 400 Bad Request.
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}
	// Verify the error message contains "invalid JSON".
	if !strings.Contains(rr.Body.String(), "invalid JSON") {
		t.Errorf("expected error message containing 'invalid JSON', got %q", rr.Body.String())
	}
}
