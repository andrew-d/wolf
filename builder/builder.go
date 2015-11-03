package builder

import (
	"github.com/andrew-d/wolf2/types"
)

// Builder is an interface that allows you to define handlers for given routes,
// along with their associated middleware.
//
// It contains functions to manage middleware (Use, Group), methods to create
// and mount subbuilders (Route, Mount), the main method that is used to define
// routes (Handle), and a set of convenience methods for common HTTP verbs.
type Builder interface {
	// Add a middleware function to this builder.  All routes registered on
	// this builder (and all subbuilders, e.g. those created with Group, Route,
	// etc.) will be wrapped in this middleware.
	Use(m types.MiddlewareType)

	// Create a new middleware group.  The given function is called with a new
	// builder that is exactly the same as this builder (i.e. with no path
	// changes), except that middleware registered on the new builder are not
	// registered on the parent.
	Group(fn func(r Builder))

	// Create a subbuilder with a given prefix.  The given function is called
	// with a new builder that registers routes with the given prefix.  Note
	// that this does minimal parsing of the given pattern - it essentially
	// adds the given prefix to all routes underneath it.
	//
	// Middleware is handled similar to the Group function - a middleware added
	// in a subbuilder will not affect the parent.
	Route(pattern string, fn func(r Builder))

	// Mount another builder as a subbuilder.  This copies all route
	// definitions from the given Builder to this one (including all
	// middleware).
	Mount(pattern string, sr Builder)

	// Main handler method
	Handle(method string, pattern types.PatternType, handler types.HandlerType)

	// Helper functions
	Delete(pattern types.PatternType, handler types.HandlerType)
	Get(pattern types.PatternType, handler types.HandlerType)
	Head(pattern types.PatternType, handler types.HandlerType)
	Options(pattern types.PatternType, handler types.HandlerType)
	Patch(pattern types.PatternType, handler types.HandlerType)
	Post(pattern types.PatternType, handler types.HandlerType)
	Put(pattern types.PatternType, handler types.HandlerType)

	// Returns a list of all route definitions on this builder (note: this
	// includes all definitions from attached subbuilders, groups, etc.)
	RouteDefs() []RouteDef
}

// This type represents a single route definition.
type RouteDef struct {
	Method     string
	Pattern    types.PatternType
	Handler    types.HandlerType
	Middleware []types.MiddlewareType
}

// New creates a new builder with no existing middleware or routes.
func New() Builder {
	return newBuilder()
}
