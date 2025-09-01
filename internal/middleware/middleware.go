package middleware

import "net/http"

func TextPlainContentType(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		handler.ServeHTTP(w, r)
	})
}

func JSONContentType(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		handler.ServeHTTP(w, r)
	})
}
