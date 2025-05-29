package routing

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// Context wraps request and response and provides helper methods.
type Context struct {
	Response http.ResponseWriter
	Request  *http.Request
	Params   map[string]string
}

// Text writes plain text response with status code.
func (c *Context) Text(status int, text string) {
	c.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Response.WriteHeader(status)
	_, _ = c.Response.Write([]byte(text))
}

// JSON sends a JSON response with the given status code.
func (ctx *Context) JSON(status int, data any) {
	ctx.Response.Header().Set("Content-Type", "application/json")
	ctx.Response.WriteHeader(status)
	_ = json.NewEncoder(ctx.Response).Encode(data)
}

// Query retrieves the query value for a key or returns the fallback value.
func (ctx *Context) Query(key, fallback string) string {
	val := ctx.Request.URL.Query().Get(key)
	if val == "" {
		return fallback
	}
	return val
}

// QueryValues returns all query parameters.
func (ctx *Context) QueryValues() url.Values {
	return ctx.Request.URL.Query()
}

// Param returns a path parameter by key or empty string if not present.
func (ctx *Context) Param(key string) string {
	if ctx.Params == nil {
		return ""
	}
	return ctx.Params[key]
}
