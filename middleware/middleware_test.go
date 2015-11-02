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
	t.Parallel()

	makeCanonical(func(ctx *context.Context, h http.Handler) http.Handler { return nil })
	makeCanonical(func(h http.Handler) http.Handler { return nil })

	assert.Panics(t, func() {
		makeCanonical(func(i int) int { return i + 1 })
	})
}

func TestMiddlewareOrder(t *testing.T) {
	t.Parallel()

	final, run := makeFinalFunc()
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

	si := stack.Get()
	defer stack.Release(si)

	// Both middleware should run
	sendRequest(si.Handler)
	assert.True(t, *run)
	assert.Equal(t, []string{"one", "two"}, calls)
}

func TestRemove(t *testing.T) {
	t.Parallel()

	final, run := makeFinalFunc()
	stack := New(final, nil)

	var calls []string

	// Make three middleware functions
	middlewareMaker := func(name string) func(http.Handler) http.Handler {
		return func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, name)
				h.ServeHTTP(w, r)
			})
		}
	}
	mw1 := middlewareMaker("one")
	mw2 := middlewareMaker("two")
	mw3 := middlewareMaker("three")

	// Push them both on the stack.
	stack.Push(mw1)
	stack.Push(mw2)
	stack.Push(mw3)

	// Assert that we run all middleware.
	si := stack.Get()
	sendRequest(si.Handler)
	assert.True(t, *run)
	assert.Equal(t, []string{"one", "two", "three"}, calls)
	stack.Release(si)

	// Reset our state ...
	*run = false
	calls = []string{}

	// ... remove the second middleware ...
	stack.Remove(mw2)

	// ... and assert that it's actually gone
	si = stack.Get()
	sendRequest(si.Handler)
	assert.True(t, *run)
	assert.Equal(t, []string{"one", "three"}, calls)
	stack.Release(si)
}

func sendRequest(h http.Handler) error {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return err
	}

	h.ServeHTTP(w, r)
	return nil
}

func makeFinalFunc() (FinalFunc, *bool) {
	run := new(bool)
	final := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		*run = true
	}

	return final, run
}
