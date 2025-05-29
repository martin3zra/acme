package routing

import (
	"net/http"
	"reflect"
	"strings"
)

// HandlerFunc defines the function signature for a route handler.
type HandlerFunc func(*Context)

// Middleware defines a function that wraps a HandlerFunc.
type Middleware func(HandlerFunc) HandlerFunc

// Route represents an HTTP route by associating an HTTP method, a URL pattern,
// a final handler, and any middleware to be explicitly excluded.
type Route struct {
	Method   string
	Pattern  string
	Handler  HandlerFunc
	Excluded []Middleware
}

// WithoutMiddleware excludes the specified middleware from being applied
// to this route.
func (rt *Route) WithoutMiddleware(mw Middleware) *Route {
	rt.Excluded = append(rt.Excluded, mw)
	return rt
}

// Router holds the collection of registered routes, an optional path prefix,
// and the middleware chain.
type Router struct {
	routes      []Route
	prefix      string
	middlewares []Middleware
}

// New creates a new Router instance with an empty route list and middleware chain.
func New() *Router {
	return &Router{
		routes:      []Route{},
		prefix:      "",
		middlewares: []Middleware{},
	}
}

// func (r *Router) WithMiddleware(mwf ...MiddlewareFunc) *Router {
// 	r.middlewares = append(r.middlewares, mwf...)
// 	return r
// }

// WithMiddleware appends a middleware to the router's middleware chain.
func (r *Router) WithMiddleware(mw ...Middleware) *Router {
	r.middlewares = append(r.middlewares, mw...)
	return r
}

// WithoutNextMiddleware returns a new Router based on the parent's settings
// but without the specified middleware in the chain.
func (r *Router) WithoutNextMiddleware(mw Middleware) *Router {
	var newMiddlewares []Middleware
	for _, m := range r.middlewares {
		if getFunctionPointer(m) == getFunctionPointer(mw) {
			continue
		}
		newMiddlewares = append(newMiddlewares, m)
	}
	return &Router{
		routes:      []Route{},
		prefix:      r.prefix,
		middlewares: newMiddlewares,
	}
}

// GET registers a route for HTTP GET requests.
func (r *Router) GET(pattern string, handler HandlerFunc) *Route {
	return r.handle("GET", pattern, handler)
}

// POST registers a route for HTTP POST requests.
func (r *Router) POST(pattern string, handler HandlerFunc) *Route {
	return r.handle("POST", pattern, handler)
}

// PUT registers a route for HTTP PUT requests.
func (r *Router) PUT(pattern string, handler HandlerFunc) *Route {
	return r.handle("PUT", pattern, handler)
}

// DELETE registers a route for HTTP DELETE requests.
func (r *Router) DELETE(pattern string, handler HandlerFunc) *Route {
	return r.handle("DELETE", pattern, handler)
}

// handle is a helper function that registers a route for a specified HTTP method.
// It automatically prefixes the route pattern if a group prefix is set.
func (r *Router) handle(method, pattern string, handler HandlerFunc) *Route {
	fullPattern := r.prefix + pattern
	route := Route{
		Method:  method,
		Pattern: fullPattern,
		Handler: handler,
	}
	r.routes = append(r.routes, route)
	return &r.routes[len(r.routes)-1]
}

// Group creates a new nested router with a path prefix and an independent middleware chain.
// Any middleware added inside the group (i.e. after calling Group) will be applied only to the routes
// registered in that group. In order to accomplish that, this method wraps each route's handler with
// the extra middleware from the group before appending those routes back to the parent router.
func (r *Router) Group(prefix string, fn func(rg *Router)) {
	// Save parent's middleware chain.
	parentMiddleware := append([]Middleware{}, r.middlewares...)

	// Create a new router for this group. Starts with the parent's middleware chain.
	newRouter := &Router{
		routes:      []Route{},
		prefix:      r.prefix + prefix,
		middlewares: append([]Middleware{}, parentMiddleware...),
	}

	// Execute the group callback; this allows new middleware to be added.
	fn(newRouter)

	// Determine the extra middleware added in the group.
	extraMiddleware := newRouter.middlewares[len(parentMiddleware):]

	// For each route in the group, wrap its handler with the group's extra middleware.
	for i, route := range newRouter.routes {
		h := route.Handler
		// Wrap extra middleware in reverse order (outermost middleware should be applied first).
		for j := len(extraMiddleware) - 1; j >= 0; j-- {
			mw := extraMiddleware[j]
			// Only wrap if this middleware is not explicitly excluded for the route.
			if middlewareExcluded(mw, route.Excluded) {
				continue
			}
			h = mw(h)
		}
		newRouter.routes[i].Handler = h
	}

	// Finally, append the group routes (now with extra middleware wrapped) to the parent's routes.
	r.routes = append(r.routes, newRouter.routes...)
}

// ServeHTTP implements the http.Handler interface.
// It matches the incoming request against registered routes, builds the middleware chain,
// and calls the appropriate handler. If a route pattern matches but with a wrong method,
// it sends a 405 response.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := cleanPath(req.URL.Path)
	method := req.Method

	var allowedMethods []string

	// Loop over the routes to find a matching pattern.
	for _, route := range r.routes {
		if ok, _ := MatchRoute(route.Pattern, path); ok {
			// Keep track of allowed methods for this pattern.
			if !contains(allowedMethods, route.Method) {
				allowedMethods = append(allowedMethods, route.Method)
			}
			// Check if the HTTP method matches.
			if route.Method != method {
				continue
			}

			// Pattern matches and method is correct; extract any parameters.
			_, params := MatchRoute(route.Pattern, path)

			// Create a new context to pass HTTP request, response, and parameters.
			ctx := &Context{
				Response: w,
				Request:  req,
				Params:   params,
			}

			// Build the final handler by wrapping it in middlewares.
			finalHandler := route.Handler
			// Apply the middleware chain in reverse order to preserve
			// the registration order.
			for i := len(r.middlewares) - 1; i >= 0; i-- {
				mw := r.middlewares[i]
				if middlewareExcluded(mw, route.Excluded) {
					continue
				}
				finalHandler = mw(finalHandler)
			}
			finalHandler(ctx)
			return
		}
	}

	// If we have matching patterns but the method is not allowed.
	if len(allowedMethods) > 0 {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	// No matching route was found.
	http.NotFound(w, req)
}

// Routes returns a slice of strings that list all registered routes,
// useful for debugging or introspection.
func (r *Router) Routes() []string {
	var list []string
	for _, route := range r.routes {
		list = append(list, route.Method+" "+route.Pattern)
	}
	return list
}

// getFunctionPointer returns the pointer of a middleware function using reflection.
// It is used to compare if two middleware functions are identical.
func getFunctionPointer(mw Middleware) uintptr {
	return reflect.ValueOf(mw).Pointer()
}

// middlewareExcluded checks whether the given middleware is present in the excluded list.
func middlewareExcluded(mw Middleware, excluded []Middleware) bool {
	for _, ex := range excluded {
		if getFunctionPointer(mw) == getFunctionPointer(ex) {
			return true
		}
	}
	return false
}

// contains checks if a slice of strings contains a given string.
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// cleanPath normalizes the URL path, removing trailing slashes (except for the root).
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// matchRoute checks if the given request path matches the defined route pattern.
// It supports dynamic segments specified with a colon (e.g., "/users/:id").
// If the pattern matches the path, it returns true along with a map of the parameters.
// Otherwise, it returns false and a nil map.
//
// The method performs the following steps:
// 1. Normalize the pattern and path by trimming leading/trailing slashes.
// 2. Split both the pattern and the path into segments.
// 3. If the segment counts differ, the route does not match.
// 4. For each segment, it checks:
//   - If the pattern segment starts with a colon, the corresponding path segment
//     is treated as a value to capture into a parameter map.
//   - If the pattern segment is static (i.e., does not start with a colon),
//     then it must exactly match the corresponding path segment.
func MatchRoute(pattern, path string) (bool, map[string]string) {
	// Remove leading and trailing slashes for consistency.
	pattern = strings.Trim(pattern, "/")
	path = strings.Trim(path, "/")

	// Split pattern and path into segments.
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	// If the number of segments is different, it's not a match.
	if len(patternParts) != len(pathParts) {
		return false, nil
	}

	params := make(map[string]string)
	// Compare each segment.
	for i, part := range patternParts {
		// If the segment starts with ":", then capture it as a parameter.
		if strings.HasPrefix(part, ":") {
			// The parameter name is the segment without the colon.
			paramName := part[1:]
			params[paramName] = pathParts[i]
		} else if part != pathParts[i] {
			// If the segment does not match and is not a parameter, return false.
			return false, nil
		}
	}
	return true, params
}
