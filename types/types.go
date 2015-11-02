package types

// HandlerType is an alias for interface{}, but is documented here for clarity.
// wolf will accept a handler of one of the following types, and will convert
// it to the Handler interface that is used internally.
//
//	- types that implement http.Handler
//	- types that implement Handler
//	- func(http.ResponseWriter, *http.Request)
//	- func(context.Context, http.ResponseWriter, *http.Request)
type HandlerType interface{}

// MiddlewareType is an alias for interface{}, but is documented here for
// clarity.  wolf will accept middleware of one of the following types, and
// will convert it to the internal middleware type.
//
//	- func(*context.Context, http.Handler) http.Handler
//	- func(http.Handler) http.Handler
type MiddlewareType interface{}

// TODO: pattern type?
