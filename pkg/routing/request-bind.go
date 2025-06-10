package routing

import (
	"net/http"

	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/support"
)

// WithRequest allows defining a handler with a JSON-decoded request body.
func WithRequest[T any](handler func(ctx *Context, body *T)) HandlerFunc {
	return func(ctx *Context) {
		var body T

		err := support.ParseRequest(ctx.Request, &body)
		if err != nil {

			if e, ok := err.(foundation.ErrorFormatter); ok {
				if e.Status() == http.StatusForbidden {
					ctx.Error(err, e.Status())
					return
				}
			}
			ctx.Back()
			return
		}

		handler(ctx, &body)
	}
}
