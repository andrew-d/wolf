package builder

import (
	"fmt"

	"github.com/andrew-d/wolf2/types"
)

var _ = fmt.Println

type routeSpec struct {
	method  string
	handler types.HandlerType

	// TODO: future support for per-route middleware would go here
}

type builderSpec struct {
	// True if this builder should inherit the middleware from the parent.
	inherit bool

	builder Builder
}

type routeOrBuilderSpec struct {
	pattern string

	// Only one of these will be given
	subBuilder *builderSpec
	route      *routeSpec
}

type builder struct {
	specs      []routeOrBuilderSpec
	middleware []types.MiddlewareType
}

func newBuilder() *builder {
	return &builder{}
}

func (r *builder) Handle(method, path string, handler types.HandlerType) {
	r.specs = append(r.specs, routeOrBuilderSpec{
		pattern: path,
		route: &routeSpec{
			method:  method,
			handler: handler,
		},
	})
}

func (r *builder) Use(m types.MiddlewareType) {
	r.middleware = append(r.middleware, m)
}

func (r *builder) Group(fn func(r Builder)) {
	r.Route("", fn)
}

func (r *builder) Route(pattern string, fn func(r Builder)) {
	// Create a new builder.
	sub := newBuilder()

	// Call the function in order to register things.
	fn(sub)

	// Append this builder to our specifications array.
	r.specs = append(r.specs, routeOrBuilderSpec{
		pattern: pattern,
		subBuilder: &builderSpec{
			inherit: true,
			builder: sub,
		},
	})
}

func (r *builder) Mount(pattern string, sr Builder) {
	// Append this builder to our specifications array, but explicitly mark it
	// as 'not inheriting'.
	r.specs = append(r.specs, routeOrBuilderSpec{
		pattern: pattern,
		subBuilder: &builderSpec{
			inherit: false,
			builder: sr,
		},
	})
}

// TODO: Need to refactor way we store route definitions.  Should:
//   - Store either a route definition or a sub-builder
//   - When getting definitions, we recurse into subbuilders
//     - Sub-builders should be 'inherit' or not.  Inherit builders inherit
//       our middleware specification, whereas others do not.
//     - We actually build the final route defs only when the RouteDefs()
//       call is made
//
// Group, Route inherit, 'Mount' does not

func (r *builder) RouteDefs() []RouteDef {
	defs := []RouteDef{}
	seen := map[*builder]struct{}{}

	// Recursively traverse the routes array.
	var walk func(*builder, []types.MiddlewareType)
	walk = func(b *builder, middleware []types.MiddlewareType) {
		// If we've seen this builder before, then we've hit a cycle.
		if _, ok := seen[b]; ok {
			msg := fmt.Sprintf(`Cycle detected while traversing router: saw `+
				`the builder %+v more than once`, b)
			panic(msg)
		}
		seen[b] = struct{}{}

		// Walk the specs in this builder.
		for _, spec := range b.specs {
			mware := make([]types.MiddlewareType, 0, len(middleware)+len(b.middleware))

			// Simple case - this is a route specification.  Copy the spec.
			if spec.route != nil {
				mware = append(mware, middleware...)
				mware = append(mware, b.middleware...)

				defs = append(defs, RouteDef{
					Method:     spec.route.method,
					Pattern:    spec.pattern,
					Handler:    spec.route.handler,
					Middleware: mware,
				})
			} else if spec.subBuilder != nil {
				// If this builder inherits, then we copy the middleware -
				// otherwise, we do nothing in order to pass the empty array
				// through.
				if spec.subBuilder.inherit {
					mware = append(mware, middleware...)
					mware = append(mware, b.middleware...)
				}

				// TODO: do we always have the same builder type?
				sb := spec.subBuilder.builder.(*builder)

				// Recurse into the sub-builder.
				walk(sb, mware)
			} else {
				panic("BUG: neither route or builder")
			}
		}
	}

	walk(r, nil)

	return defs
}

// Helper functions below here

func (r *builder) Delete(path string, handler types.HandlerType) {
	r.Handle("DELETE", path, handler)
}

func (r *builder) Get(path string, handler types.HandlerType) {
	r.Handle("GET", path, handler)
}

func (r *builder) Head(path string, handler types.HandlerType) {
	r.Handle("HEAD", path, handler)
}

func (r *builder) Options(path string, handler types.HandlerType) {
	r.Handle("OPTIONS", path, handler)
}

func (r *builder) Patch(path string, handler types.HandlerType) {
	r.Handle("PATCH", path, handler)
}

func (r *builder) Post(path string, handler types.HandlerType) {
	r.Handle("POST", path, handler)
}

func (r *builder) Put(path string, handler types.HandlerType) {
	r.Handle("PUT", path, handler)
}

var _ Builder = &builder{}