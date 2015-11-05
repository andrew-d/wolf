package types

// HandlerType is an alias for interface{}, but is documented here for clarity.
// Packages that use portions of wolf are free to accept any type, but all
// components of wolf will accept a handler of one of the following types, and
// will convert it to the Handler interface that is used internally.
//
//	- types that implement http.Handler
//	- types that implement Handler
//	- func(http.ResponseWriter, *http.Request)
//	- func(context.Context, http.ResponseWriter, *http.Request)
type HandlerType interface{}

// MiddlewareType is an alias for interface{}, but is documented here for
// clarity.  Packages that use portions of wolf are free to accept any type,
// but all components of wolf will accept middleware of one of the following
// types, and will convert it to the internal middleware type.
//
//	- func(*context.Context, http.Handler) http.Handler
//	- func(http.Handler) http.Handler
type MiddlewareType interface{}

// PatternType is an alias for interface{}, but is documented here for clarity.
// Packages that use portions of wolf are free to accept any type, but all
// components of wolf will accept a pattern of one of the following types, and
// will convert it to an underlying 'Pattern' interface.
//
//   - types that implement the Pattern interface
//   - string, which is interpreted as a Sinatra-like URL pattern. In
//     particular, the following syntax is recognized:
//   	- a path segment starting with a colon will match any
//   	  string placed at that position. e.g., "/:name" will match
//   	  "/carl", binding "name" to "carl".
//   	- a pattern ending with "/*" will match any route with that
//   	  prefix. For instance, the pattern "/u/:name/*" will match
//   	  "/u/carl/" and "/u/carl/projects/123", but not "/u/carl"
//   	  (because there is no trailing slash). In addition to any names
//   	  bound in the pattern, the special key "*" is bound to the
//   	  unmatched tail of the match, but including the leading "/". So
//   	  for the two matching examples above, "*" would be bound to "/"
//   	  and "/projects/123" respectively.
//     Unlike http.ServeMux's patterns, string patterns support neither the
//     "rooted subtree" behavior nor Host-specific routes. Users who require
//     either of these features are encouraged to compose package http's mux
//     with the mux provided by this package.
//   - regexp.Regexp, which is assumed to be a Perl-style regular expression
//     that is anchored on the left (i.e., the beginning of the string). If
//     your regular expression is not anchored on the left, a
//     hopefully-identical left-anchored regular expression will be created
//     and used instead.
//
//     Capturing groups will be converted into bound URL parameters in
//     URLParams. If the capturing group is named, that name will be used;
//     otherwise the special identifiers "$1", "$2", etc. will be used.
type PatternType interface{}
