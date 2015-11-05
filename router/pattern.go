package router

import (
	"fmt"
	"net/http"
	"regexp"

	"golang.org/x/net/context"

	"github.com/andrew-d/wolf2/types"
)

// A Pattern determines whether or not a given request matches some criteria.
// They are often used in routes, which are essentially (pattern, methodSet,
// handler) tuples. If the method and pattern match, the given handler is used.
//
// Built-in implementations of this interface are used to implement regular
// expression and string matching.
type Pattern interface {
	// In practice, most real-world routes have a string prefix that can be
	// used to quickly determine if a pattern is an eligible match. The
	// router uses the result of this function to optimize away calls to the
	// full Match function, which is likely much more expensive to compute.
	// If your Pattern does not support prefixes, this function should
	// return the empty string.
	Prefix() string

	// Returns true if the request satisfies the pattern. This function should
	// only examine the request to make this decision, and should be idempotent
	// (i.e. it should be a pure function). Match should not modify its
	// arguments, and since it will potentially be called several times over
	// the course of matching a request, it should be reasonably efficient.
	Match(r *http.Request) bool

	// Run the pattern on the request and context, modifying the context as
	// necessary to bind URL parameters or other parsed state.
	Run(r *http.Request, ctx *context.Context)
}

// ParsePattern is used internally by Goji to parse route patterns. It is
// exposed publicly to make it easier to write thin wrappers around the
// built-in Pattern implementations.
//
// ParsePattern panics if it is passed a value of an unexpected type (see the
// documentation for PatternType for a list of which types are accepted). It is
// the caller's responsibility to ensure that ParsePattern is called in a
// type-safe manner.
func ParsePattern(raw types.PatternType) Pattern {
	switch v := raw.(type) {
	case Pattern:
		return v
	case *regexp.Regexp:
		return ParseRegexpPattern(v)
	case string:
		return ParseStringPattern(v)
	default:
		msg := fmt.Sprintf(`Unknown pattern type %T. See `+
			`https://godoc.org/github.com/andrew-d/wolf/types#PatternType `+
			`for a list of acceptable types.`, v)
		panic(msg)
	}
}
