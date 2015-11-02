package middleware

import (
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/net/context"

	"github.com/andrew-d/wolf2/types"
)

// canonicalMiddleware is the 'canonical' middleware type - we coerce all other
// middlewares to this type.
type canonicalMiddleware func(ctx *context.Context, h http.Handler) http.Handler

// FinalFunc is the type of the function that is called at the end of a
// middleware chain.  Generally, this will dispatch to the user-provided
// handler function.
type FinalFunc func(context.Context, http.ResponseWriter, *http.Request)

// MiddlewareStack represents a collection of 'stacks' of middleware - a set of
// middleware to apply in order, followed by a final handler function.  It
// maintains an internal cache of stacks in order to improve performance.
//
// MiddlewareStack is safe for use in multiple goroutines concurrently.
type MiddlewareStack struct {
	// Cache of pre-built middleware stacks
	cache *sync.Pool

	// The final handler that we call after applying all middleware.
	final FinalFunc

	// The base context for all middleware (i.e. this is passed to the first
	// middleware).  By default, it is set to `context.Background()`.
	BaseCtx context.Context

	// List of middleware functions
	funcs []canonicalMiddleware
	mu    sync.Mutex

	// List of input middleware values (i.e. not coerced to `canonicalMiddleware`).
	orig []types.MiddlewareType
}

// New creates and returns a new middleware stack with some initial set of
// middleware.
func New(handler FinalFunc, middleware []types.MiddlewareType) *MiddlewareStack {
	m := &MiddlewareStack{
		final:   handler,
		BaseCtx: context.Background(),
	}

	// Push all existing.  We can use the 'unlocked' version since we're the
	// only thing that owns this stack right now.
	for _, mw := range middleware {
		m.push(mw)
	}

	m.resetPool()
	return m
}

// Push adds a new middleware to the current stack.  This invalidates any
// existing cached stacks.
func (m *MiddlewareStack) Push(mw types.MiddlewareType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Call 'real' push function
	m.push(mw)

	// Invalidate any existing cache
	m.resetPool()
}

// Convert a middleware into our canonical type.  Panics on error.
func makeCanonical(mw types.MiddlewareType) canonicalMiddleware {
	var resolvedFn canonicalMiddleware

	switch f := mw.(type) {
	case func(http.Handler) http.Handler:
		resolvedFn = func(ctx *context.Context, h http.Handler) http.Handler {
			return f(h)
		}
	case func(*context.Context, http.Handler) http.Handler:
		resolvedFn = f
	default:
		msg := fmt.Sprintf(`Invalid middleware type '%T'.  See `+
			`https://godoc.org/github.com/andrew-d/wolf/types#MiddlewareType `+
			`for a list of valid middleware types`, mw)
		panic(msg)
	}

	return resolvedFn
}

// Add a new middleware to the current stack, without locking or resetting the
// cache.  Panics on error.
func (m *MiddlewareStack) push(mw types.MiddlewareType) {
	// We store both the original and canonical functions, so we can remove a
	// middleware
	m.orig = append(m.orig, mw)
	m.funcs = append(m.funcs, makeCanonical(mw))
}

// Remove a middleware from the stack.  Does nothing if the given middleware is
// not in this stack.
func (m *MiddlewareStack) Remove(mw types.MiddlewareType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find the index of this middleware
	idx := -1
	for i, test := range m.orig {
		if funcEqual(test, mw) {
			idx = i
			break
		}
	}

	// Found it?
	if idx < 0 {
		return
	}

	// Remove from the array
	m.orig = append(m.orig[:idx], m.orig[idx+1:]...)
	m.funcs = append(m.funcs[:idx], m.funcs[idx+1:]...)

	// Invalidate the middleware cache, since we've changed things
	m.resetPool()
}

// Reset (invalidate) any cached stacks.
func (m *MiddlewareStack) resetPool() {
	// Create an entirely new pool (the old one gets garbage-collected)
	m.cache = &sync.Pool{
		New: m.newResolved,
	}
}

// Get obtains a new middleware stack from the cache.
func (m *MiddlewareStack) Get() http.Handler {
	return m.cache.Get().(http.Handler)
}

// Release puts a previously-obtained middleware stack back into the cache.
func (m *MiddlewareStack) Release(h http.Handler) {
	m.cache.Put(h)
}

// Constructor function that is used to create new middleware stacks when the
// cache does not have any available values.
//
// This is where the actual middlewares are applied.
func (m *MiddlewareStack) newResolved() interface{} {
	// Copy variables in 'm' that are used in the function below.
	ctx := m.BaseCtx
	final := m.final

	var finalFunc http.Handler
	finalFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Dispatch to our final handler.
		final(ctx, w, r)
	})

	// Apply all middleware.
	for i := len(m.funcs) - 1; i >= 0; i-- {
		finalFunc = m.funcs[i](&ctx, finalFunc)
	}

	return finalFunc
}
