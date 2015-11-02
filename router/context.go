package router

import (
	"golang.org/x/net/context"
)

type private struct{}

var urlParamKey private

// AddURLParams will add the given URL parameters to the given context.
func SetURLParams(ctx context.Context, matches map[string]string) context.Context {
	return context.WithValue(ctx, urlParamKey, matches)
}

// GetURLParams will retrieve the URL parameters map from the given context.
func GetURLParams(ctx context.Context) map[string]string {
	val := ctx.Value(urlParamKey)
	if val == nil {
		return nil
	}

	return val.(map[string]string)
}
