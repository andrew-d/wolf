package simple

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/andrew-d/wolf2/builder"
	"github.com/andrew-d/wolf2/middleware"
	"github.com/andrew-d/wolf2/router"
)

// A combination of a route's pattern, handler, and the middleware stack.
type route struct {
	pattern router.Pattern
	handler router.Handler
	mware   *middleware.MiddlewareStack
}

// SimpleRouter is the simplest-possible router - it checks each route in
// sequence for a match, and dispatches to the first one.
type SimpleRouter struct {
	// Map of HTTP method --> route array
	routes map[string][]route

	// NotFound will be run whenever no route is matched (if non-nil).
	NotFound router.Handler
}

// New takes a list of route definitions (generally created by using the
// builder package) and returns a router instance.
func New(routeDefs []builder.RouteDef) *SimpleRouter {
	methods := make(map[string][]route)

	for _, def := range routeDefs {
		r := route{
			pattern: router.ParsePattern(def.Pattern),
			handler: router.MakeHandler(def.Handler),
		}

		// Point the middleware at the handler's serve function.
		r.mware = middleware.New(r.handler.ServeHTTPC, def.Middleware)

		methods[def.Method] = append(methods[def.Method], r)
	}

	return &SimpleRouter{routes: methods}
}

// This function allows SimpleRouter to implement net/http.Handler
func (s *SimpleRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	found := false

	for _, route := range s.routes[r.Method] {
		stack := route.mware.Get()

		if route.pattern.Match(r, &stack.Context) {
			found = true
			route.pattern.Run(r, &stack.Context)
			stack.Handler.ServeHTTP(w, r)
		}

		route.mware.Release(stack)
	}

	// Support NotFound handler
	if !found && s.NotFound != nil {
		s.NotFound.ServeHTTPC(context.Background(), w, r)
	}
}
