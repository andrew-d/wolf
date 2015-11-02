package builder

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func noopHandler(c context.Context, w http.ResponseWriter, r *http.Request) {}

// Test that the Handle() function adds to the RouteDefs array.
func TestHandle(t *testing.T) {
	b := New()
	b.Handle("GET", "/", noopHandler)

	rd := b.RouteDefs()
	if assert.Len(t, rd, 1) {
		assert.Equal(t, rd[0].Method, "GET")
		assert.Equal(t, rd[0].Pattern, "/")
		assert.Len(t, rd[0].Middleware, 0)

		// TODO: assert handler function equality?
		assert.NotNil(t, rd[0].Handler)
	}
}

// Test that the various verbs add to the RouteDefs array.
func TestVerbShorthand(t *testing.T) {
	b := New()

	b.Delete("/", noopHandler)
	b.Get("/", noopHandler)
	b.Head("/", noopHandler)
	b.Options("/", noopHandler)
	b.Patch("/", noopHandler)
	b.Post("/", noopHandler)
	b.Put("/", noopHandler)

	verbs := []string{}
	for _, def := range b.RouteDefs() {
		verbs = append(verbs, def.Method)
		assert.Equal(t, def.Pattern, "/")
		assert.Len(t, def.Middleware, 0)

		// TODO: assert handler function equality?
		assert.NotNil(t, def.Handler)
	}

	assert.Equal(t, verbs, []string{
		"DELETE",
		"GET",
		"HEAD",
		"OPTIONS",
		"PATCH",
		"POST",
		"PUT",
	})
}

// Test that a middleware function is applied both before and after route definitions.
func TestUse(t *testing.T) {
	b := New()

	// Note: these aren't valid middleware, but we don't actually type-check
	// them in the builder.
	var mw1 interface{} = 1234
	var mw2 interface{} = 5678

	b.Use(mw1)
	b.Handle("GET", "/", noopHandler)
	b.Use(mw2)

	rd := b.RouteDefs()
	if assert.Len(t, rd, 1) && assert.Len(t, rd[0].Middleware, 2) {
		assert.Equal(t, rd[0].Middleware[0], mw1)
		assert.Equal(t, rd[0].Middleware[1], mw2)
	}
}

// Test that we can create a new middleware group without affecting the parent.
func TestGroup(t *testing.T) {
	b := New()

	// Note: these aren't valid middleware, but we don't actually type-check
	// them in the builder.
	var mw1 interface{} = 1234
	var mw2 interface{} = 5678

	b.Use(mw1)
	b.Handle("GET", "/", noopHandler)
	b.Group(func(b Builder) {
		b.Use(mw2)
		b.Handle("GET", "/hello", noopHandler)
	})
	b.Handle("GET", "/foobar", noopHandler)

	rd := b.RouteDefs()
	if assert.Len(t, rd, 3) {
		assert.Len(t, rd[0].Middleware, 1)
		assert.Len(t, rd[1].Middleware, 2)
		assert.Len(t, rd[2].Middleware, 1)
	}
}
