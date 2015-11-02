package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	//"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestMiddlewareTypes(t *testing.T) {
	makeCanonical(func(ctx *context.Context, h http.Handler) http.Handler { return nil })
	makeCanonical(func(h http.Handler) http.Handler { return nil })

	assert.Panics(t, func() {
		makeCanonical(func(i int) int { return i + 1 })
	})
}

func TestMiddlewareOrder(t *testing.T) {
	var run bool
	final := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		run = true
		assert.NotNil(t, ctx)
	}

	stack := New(final, nil)

	var calls []string

	// Verify that our middleware type works
	stack.Push(func(ctx *context.Context, h http.Handler) http.Handler {
		wrap := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "one")
			h.ServeHTTP(w, r)
		})
		return wrap
	})

	// Standard library-ish middleware type should work too
	stack.Push(func(h http.Handler) http.Handler {
		wrap := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "two")
			h.ServeHTTP(w, r)
		})
		return wrap
	})

	// Get a handler
	handler := stack.Get()
	defer stack.Release(handler)

	var w http.ResponseWriter = httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/foo", nil)
	assert.NoError(t, err)
	handler.ServeHTTP(w, r)

	assert.True(t, run)
	assert.Equal(t, []string{"one", "two"}, calls)
}
