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
	// Iterate over all the route definitions and save the routes for each
	// method in a map, indexed by HTTP method.
	methods := make(map[string][]route)
	for _, def := range routeDefs {
		// A route contains a parsed pattern and handler.
		r := route{
			pattern: router.ParsePattern(def.Pattern),
			handler: router.MakeHandler(def.Handler),
		}

		// The middleware's "final function" is simply the handler's serve
		// function.
		r.mware = middleware.New(r.handler.ServeHTTPC, def.Middleware)

		// Save this route.
		methods[def.Method] = append(methods[def.Method], r)
	}

	return &SimpleRouter{routes: methods}
}

// This function allows SimpleRouter to implement net/http.Handler
func (s *SimpleRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	found := false

	// Iterate over all routes for this method.
	for _, route := range s.routes[r.Method] {
		stack := route.mware.Get()

		// If the route matches, then we run the matching again in order to
		// capture any variables from dynamic portions of the route, and then
		// run the actual handler.
		//
		// Note: the handler will actually dispatch to the middleware, and then
		// the final handler function.
		if route.pattern.Match(r, &stack.Context) {
			found = true
			route.pattern.Run(r, &stack.Context)
			stack.Handler.ServeHTTP(w, r)
		}

		route.mware.Release(stack)

		if found {
			break
		}
	}

	// If we didn't get a route, then we either run the user-provided not-found
	// handler (if provided), or dispatch to the standard library's NotFound
	// handler.
	if !found {
		if s.NotFound != nil {
			s.NotFound.ServeHTTPC(context.Background(), w, r)
		} else {
			http.NotFound(w, r)
		}
	}
}
