package routing

import (
	"encoding/json"
	"io"
	"net/http"
)

// WithRequest allows defining a handler with a JSON-decoded request body.
func WithRequest[T any](handler func(ctx *Context, body *T)) HandlerFunc {
	return func(ctx *Context) {
		var body T
		raw, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
			return
		}

		if err := json.Unmarshal(raw, &body); err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON", "details": err.Error()})
			return
		}

		handler(ctx, &body)
	}
}
