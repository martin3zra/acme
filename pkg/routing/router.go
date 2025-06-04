package routing

import (
	"net/http"
	"reflect"
	"strings"

	"slices"

	"github.com/romsar/gonertia/v2"
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
	Wrapped  bool // Indicates that the handler has been pre-wrapped.
}

// WithoutMiddleware excludes the specified middleware from being applied
// on this route.
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
	// groupExcluded holds middleware to _exclude_ from this router’s group.
	groupExcluded []Middleware
	inertia       *gonertia.Inertia
	parent        *Router // pointer to parent router (nil if root)
}

// New creates a new Router instance with an empty route list and middleware chain.
func New() *Router {
	return &Router{
		routes:      []Route{},
		prefix:      "",
		middlewares: []Middleware{},
	}
}

func (r *Router) RegisterInertia(i *gonertia.Inertia) *Router {
	r.inertia = i
	return r
}

// WithoutNextMiddleware removes (in place) the first occurrence of the specified middleware
// from the router's middleware slice.
func (r *Router) WithoutNextMiddleware(mw Middleware) *Router {
	for i := 0; i < len(r.middlewares); i++ {
		if getFunctionPointer(r.middlewares[i]) == getFunctionPointer(mw) {
			r.middlewares = append(r.middlewares[:i], r.middlewares[i+1:]...)
			break // remove only one occurrence.
		}
	}
	return r
}

// WithMiddleware appends the given middleware to the router's middleware chain (in place)
// and returns the same router.
func (r *Router) WithMiddleware(mws ...Middleware) *Router {
	r.middlewares = append(r.middlewares, mws...)
	return r
}

// WithoutGroupMiddleware marks the given middleware as excluded for every route in this group.
// (That is, routes registered in a group on this router will have the given middleware skipped.)
func (r *Router) WithoutGroupMiddleware(mw Middleware) *Router {
	r.groupExcluded = append(r.groupExcluded, mw)
	return r
}

// FileServer registers a static file handler under the given URL prefix using the provided file system.
// You can obtain the correct http.FileSystem (via embed.FS or os.DirFS) using build flags before calling this method.
//
// For example, if you pass in a fileSystem that corresponds to the "public/build" directory,
// a URL request for "/build/assets/app.js" will be served from "assets/app.js" relative to that file system.
func (r *Router) FileServer(prefix string, fileSystem http.FileSystem) {
	// Ensure the prefix starts with "/" and ends with "/"
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	// Create a file handler that strips the prefix from requests
	fileHandler := http.StripPrefix(prefix, http.FileServer(fileSystem))

	fileRoutePattern := prefix + ":filepath..."
	fileRoute := r.GET(fileRoutePattern, func(ctx *Context) {
		fileHandler.ServeHTTP(ctx.Response, ctx.Request)
	})

	// Exclude every global middleware for the file route.
	for _, mw := range r.middlewares {
		fileRoute.WithoutMiddleware(mw)
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
// It automatically prefixes the route pattern and inherits group-wide exclusions.
func (r *Router) handle(method, pattern string, handler HandlerFunc) *Route {
	fullPattern := r.prefix + pattern
	route := Route{
		Method:  method,
		Pattern: fullPattern,
		Handler: handler,
		// Each route copies the current group exclusion list.
		Excluded: append([]Middleware{}, r.groupExcluded...),
	}
	r.routes = append(r.routes, route)
	return &r.routes[len(r.routes)-1]
}

// Prefix returns a new router instance with the given prefix appended to the current prefix,
// and inherits the middleware and group exclusion settings.
func (r *Router) Prefix(prefix string) *Router {
	return &Router{
		routes:        []Route{},
		prefix:        r.prefix + prefix,
		middlewares:   append([]Middleware{}, r.middlewares...),
		groupExcluded: append([]Middleware{}, r.groupExcluded...),
		inertia:       r.inertia,
		parent:        r,
	}
}

// Group creates a temporary child router for grouping routes. It first saves the parent's
// middleware and exclusion state, then creates a new child router that inherits those settings.
// The group's callback registers routes on the child. After the callback returns, each route in
// the child is pre-wrapped with the child's middleware chain (skipping those in its exclusion list),
// and then merged back into the parent's route table. Finally, the parent's middleware state is restored.
func (r *Router) Group(fn func(rg *Router)) {
	// Save parent's state.
	savedMiddlewares := append([]Middleware{}, r.middlewares...)
	savedExcluded := append([]Middleware{}, r.groupExcluded...)

	// Create a temporary child router.
	child := &Router{
		routes:        []Route{},
		prefix:        r.prefix,
		middlewares:   append([]Middleware{}, r.middlewares...),
		groupExcluded: append([]Middleware{}, r.groupExcluded...),
		inertia:       r.inertia,
		parent:        r,
	}

	// Execute the group's callback on the child router.
	fn(child)

	// Pre-wrap each route in the child router.
	for i, route := range child.routes {
		effective := route.Handler
		// Apply the child's middleware chain in reverse order.
		for j := len(child.middlewares) - 1; j >= 0; j-- {
			mw := child.middlewares[j]
			// Skip middleware that are excluded for this route.
			if middlewareExcluded(mw, route.Excluded) {
				continue
			}
			effective = mw(effective)
		}
		child.routes[i].Handler = effective
		child.routes[i].Wrapped = true
	}

	// Merge the child's routes back into the parent's route table.
	r.routes = append(r.routes, child.routes...)

	// Restore the parent's middleware and exclusion state.
	r.middlewares = savedMiddlewares
	r.groupExcluded = savedExcluded
}

// ServeHTTP implements the http.Handler interface.
// For routes not pre-wrapped (i.e. registered on the global router), it applies the parent's
// middleware chain dynamically.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := cleanPath(req.URL.Path)
	method := req.Method
	var allowedMethods []string

	for _, route := range r.routes {
		if ok, _ := MatchRoute(route.Pattern, path); ok {
			if !slices.Contains(allowedMethods, route.Method) {
				allowedMethods = append(allowedMethods, route.Method)
			}
			if route.Method != method {
				continue
			}
			_, params := MatchRoute(route.Pattern, path)
			ctx := &Context{
				Response: w,
				Request:  req,
				Params:   params,
				Inertia:  r.inertia,
			}
			finalHandler := route.Handler

			// If the route wasn't pre-wrapped (i.e. is a global route), wrap it now.
			if !route.Wrapped {
				for i := len(r.middlewares) - 1; i >= 0; i-- {
					mw := r.middlewares[i]
					if middlewareExcluded(mw, route.Excluded) {
						continue
					}
					finalHandler = mw(finalHandler)
				}
			}
			finalHandler(ctx)
			return
		}
	}

	if len(allowedMethods) > 0 {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	http.NotFound(w, req)
}

// Routes returns a slice of strings listing all registered routes.
func (r *Router) Routes() []string {
	var list []string
	for _, route := range r.routes {
		list = append(list, route.Method+" "+route.Pattern)
	}
	return list
}

// getFunctionPointer returns the pointer of a middleware function using reflection.
// It is used for comparing function identity.
func getFunctionPointer(mw Middleware) uintptr {
	return reflect.ValueOf(mw).Pointer()
}

// middlewareExcluded checks whether a middleware is in the given excluded list.
func middlewareExcluded(mw Middleware, excluded []Middleware) bool {
	for _, ex := range excluded {
		if getFunctionPointer(mw) == getFunctionPointer(ex) {
			return true
		}
	}
	return false
}

// cleanPath normalizes the URL path, removing trailing slashes (except the root).
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// MatchRoute checks if a request path matches a route pattern and returns any captured parameters.
func MatchRoute(pattern, urlPath string) (bool, map[string]string) {
	pattern = strings.Trim(pattern, "/")
	urlPath = strings.Trim(urlPath, "/")

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(urlPath, "/")
	params := make(map[string]string)

	// Check if last pattern part is a catch-all (ends with '...')
	catchAll := false
	catchAllParam := ""
	if len(patternParts) > 0 {
		last := patternParts[len(patternParts)-1]
		if strings.HasPrefix(last, ":") && strings.HasSuffix(last, "...") {
			catchAll = true
			catchAllParam = last[1 : len(last)-3] // Remove ':' and '...'
			patternParts = patternParts[:len(patternParts)-1]
		}
	}

	// For non catch-all, the segments must match exactly.
	if !catchAll && len(patternParts) != len(pathParts) {
		return false, nil
	}
	// For catch-all, ensure that at least the fixed segments match.
	if catchAll && len(pathParts) < len(patternParts) {
		return false, nil
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:]
			params[paramName] = pathParts[i]
		} else if part != pathParts[i] {
			return false, nil
		}
	}
	if catchAll {
		params[catchAllParam] = strings.Join(pathParts[len(patternParts):], "/")
	}

	return true, params
}
