package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recover перехватывает панику в обработчике: сервер не падает,
// пользователь получает страницу 500 (её рисует onPanic).
func Recover(onPanic func(w http.ResponseWriter, r *http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					if rec == http.ErrAbortHandler {
						panic(rec) // штатное прерывание — отдаём серверу
					}
					log.Printf("PANIC %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
					w.Header().Set("Connection", "close")
					onPanic(w, r)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
