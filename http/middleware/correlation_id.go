package middleware

import (
	"net/http"

	"github.com/actforgood/xtransport"
)

// CorrelationID is a decorator/middleware that extracts/ads a correlation id
// from/to request/response.
func CorrelationID(next http.Handler, makeCorrelationID xtransport.CorrelationIDFactory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// extract the correlationd id from request or generate a new one
		correlationID := r.Header.Get(xtransport.CorrelationIDHeaderKey)
		if correlationID == "" {
			correlationID = makeCorrelationID()
		}
		// set the correlation id on context
		ctx := xtransport.ContextWithCorrelationID(r.Context(), correlationID)
		newR := r.WithContext(ctx)
		// send back the correlation id
		w.Header().Add(xtransport.CorrelationIDHeaderKey, correlationID)

		next.ServeHTTP(w, newR)
	})
}
