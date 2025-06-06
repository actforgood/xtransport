package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/actforgood/xlog"
	"github.com/actforgood/xtransport"
	httpTransport "github.com/actforgood/xtransport/http"
)

// Recover is a decorator/middleware that gracefully logs
// any panic occurred while serving a request.
func Recover(next http.Handler, logger xlog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("an unexpected error occurred, please try again later"))
				logger.Error(
					xlog.MsgKey, "handler panic catched",
					xlog.ErrorKey, err,
					"path", r.URL.Path,
					"method", r.Method,
					"ip", httpTransport.GetClientIP(r).String(),
					"agent", r.Header.Get("User-Agent"),
					"stack", string(debug.Stack()),
					"correlationId", xtransport.CorrelationIDFromContext(r.Context()),
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
