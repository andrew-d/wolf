package router

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/andrew-d/wolf/types"
)

// Handler is similar to net/http's http.Handler, but accepts a Context from
// x/net/context as the first parameter.
type Handler interface {
	ServeHTTPC(context.Context, http.ResponseWriter, *http.Request)
}

// HandlerFunc is similar to net/http's http.HandlerFunc, but accepts a Context
// object.  It implements both the Handler interface and http.Handler.
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTP implements http.Handler, allowing HandlerFuncs to be used with
// net/http and other routers.  When used this way, the underlying function
// will be passed a Background context.
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(context.Background(), w, r)
}

// ServeHTTPC implements Handler.
func (f HandlerFunc) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

// netHTTPWrap is a helper to turn a http.Handler into our Handler
type netHTTPWrap struct {
	http.Handler
}

func (h netHTTPWrap) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.ServeHTTP(w, r)
}

// MakeHandler turns a HandlerType into something that implements our Handler
// interface.  It will panic if the input is not a valid HandlerType.
func MakeHandler(h types.HandlerType) Handler {
	switch f := h.(type) {
	case Handler:
		return f
	case http.Handler:
		return netHTTPWrap{f}
	case func(context.Context, http.ResponseWriter, *http.Request):
		return HandlerFunc(f)
	case func(http.ResponseWriter, *http.Request):
		return netHTTPWrap{http.HandlerFunc(f)}
	default:
		msg := fmt.Sprintf(`Invalid handler type '%T'.  See `+
			`https://godoc.org/github.com/andrew-d/wolf/types#HandlerType `+
			`for a list of valid handler types`, h)
		panic(msg)
	}
}
