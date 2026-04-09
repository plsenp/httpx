package httpx

import "net/http"

type Middleware func(http.Handler) http.Handler

func chainMiddleware(middleware ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			next = middleware[i](next)
		}
		return next
	}
}
