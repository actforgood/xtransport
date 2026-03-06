package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/actforgood/xtransport"
	httpTransport "github.com/actforgood/xtransport/http"
)

// AccessLevel is the log level used for access logs.
var AccessLevel slog.Level = slog.LevelError + 4

// http.ResponseWriter.
type statusAwareResponseWriter struct {
	origW      http.ResponseWriter // original response writer
	statusCode int                 // captures response status code
	bodySize   int                 // captures response body length
}

func (w *statusAwareResponseWriter) Header() http.Header {
	return w.origW.Header()
}

func (w *statusAwareResponseWriter) Write(data []byte) (int, error) {
	n, err := w.origW.Write(data)
	w.bodySize += n

	return n, err
}

func (w *statusAwareResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.origW.WriteHeader(code)
}

func (w *statusAwareResponseWriter) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}

	return w.statusCode
}

func (w *statusAwareResponseWriter) BodySize() int {
	return w.bodySize
}

// AccessRequestCallback is a function type through which you can specify a callback
// to to indicate whether a request should be logged or not, and to modify the request before logging.
//
// Usage example:
//
//	func myAccessRequestCallback(r *http.Request) bool {
//	    if r.URL.Path == "/health" {
//	        return false // skip logging for health endpoint
//	    }
//
//	    if r.PathValue("secretToken") != "" {
//	        // obfuscate secret token in the request path
//	        newPath := strings.ReplaceAll(r.URL.Path, r.PathValue("secretToken"), r.PathValue("secretToken")[:4]+"****")
//	        r.URL.Path = newPath
//	    }
//
//	    return true // log the request
//	}
type AccessRequestCallbeck func(r *http.Request) bool

// AccessLog is a decorator/middleware that extracts/ads a correlation id
// from/to request/response.
func AccessLog(next http.Handler, logger *slog.Logger, callback AccessRequestCallbeck) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, origReq *http.Request) {
		now := time.Now().UTC()
		newW := &statusAwareResponseWriter{origW: w}
		next.ServeHTTP(newW, origReq)

		r := origReq
		if callback != nil {
			r = origReq.Clone(r.Context())
			if !callback(r) {
				return // skip logging for this request
			}
		}

		logParams := make([]any, 0, 12*2)
		logParams = append(logParams,
			[]any{
				"lvl", "ACCESS",
				"method", r.Method,
				"path", r.URL.Path,
				"took", time.Since(now).String(),
				"userAgent", r.Header.Get("User-Agent"),
				"ip", httpTransport.GetClientIP(r).String(),
				"statusCode", newW.StatusCode(),
			}...,
		)
		if r.URL.RawQuery != "" {
			logParams = append(logParams, "query", r.URL.RawQuery)
		}
		correlationID := xtransport.CorrelationIDFromContext(r.Context())
		if correlationID != "" {
			logParams = append(logParams, "correlationId", correlationID)
		}
		if r.URL.User.Username() != "" {
			logParams = append(logParams, "authUsername", r.URL.User.Username())
		}
		if r.Header.Get("Content-Length") != "" {
			reqContentLen, _ := strconv.Atoi(r.Header.Get("Content-Length"))
			logParams = append(logParams, "reqContentLength", reqContentLen)
		}
		if w.Header().Get("Content-Length") != "" {
			respContentLen, _ := strconv.Atoi(w.Header().Get("Content-Length"))
			logParams = append(logParams, "respContentLength", respContentLen)
		} else {
			logParams = append(logParams, "respBodyLength", newW.BodySize())
		}

		logger.Log(r.Context(), AccessLevel, "access log", logParams...)
	})
}
