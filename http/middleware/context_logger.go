package middleware

import (
	"net/http"

	"github.com/actforgood/xlog"
)

// ContextWithLogger is a decorator/middleware that sets the logger on request context.
func ContextWithLogger(next http.Handler, logger xlog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newCtx := xlog.ContextWithLogger(r.Context(), logger)
		newR := r.WithContext(newCtx)

		next.ServeHTTP(w, newR)
	})
}
