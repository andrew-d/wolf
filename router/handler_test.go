package router

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type dummyHandler struct{}

func (d dummyHandler) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {}

type dummyStdHandler struct{}

func (d dummyStdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func TestMakeHandler(t *testing.T) {
	// Our handler type
	assert.NotNil(t, MakeHandler(dummyHandler{}))

	// http.Handler
	var stdHandler http.Handler = dummyStdHandler{}
	assert.NotNil(t, MakeHandler(stdHandler))

	// net/http HandlerFunc style
	stdFn := func(w http.ResponseWriter, r *http.Request) {}
	assert.NotNil(t, MakeHandler(stdFn))

	// Our HandlerFunc type
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	assert.NotNil(t, MakeHandler(fn))

	// Another, incompatible type
	assert.Panics(t, func() {
		MakeHandler(func(i int) int {
			return i + 1
		})
	})
}
