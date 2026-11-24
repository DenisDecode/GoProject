// Package middleware — обёртки над http.Handler.
package middleware

import "net/http"

// Chain(h, A, B, C) == A(B(C(h))): первый middleware — самый внешний,
// он первым получает запрос и последним видит ответ.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
