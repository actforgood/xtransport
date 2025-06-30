package middleware

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/actforgood/xlog"
	"github.com/actforgood/xtransport"
	httpTransport "github.com/actforgood/xtransport/http"
)

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

// AccessLogOpts holds some configuration for access log.
type AccessLogOpts struct {
	// SkipMethods specifies the http methods to skip logging for.
	SkipMethods []string
	// ObfuscatePathValues specifies the request url parts that should be obscure.
	// Last maximum 8 chars from it will be replaced with "*".
	ObfuscatePathValues []string
}

// AccessLog is a decorator/middleware that extracts/ads a correlation id
// from/to request/response.
func AccessLog(next http.Handler, logger xlog.Logger, opts ...AccessLogOpts) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opt AccessLogOpts
		if len(opts) > 0 {
			opt = opts[0]
		}
		for _, skipMethod := range opt.SkipMethods {
			if r.Method == skipMethod {
				next.ServeHTTP(w, r)

				return
			}
		}
		now := time.Now().UTC()

		newW := &statusAwareResponseWriter{origW: w}
		next.ServeHTTP(newW, r)

		urlPath := r.URL.Path
		for _, obscurePathValue := range opt.ObfuscatePathValues {
			if val := r.PathValue(obscurePathValue); val != "" {
				charsNo := int(math.Min(float64(len(val)), 8.0))
				obscureVal := val[:len(val)-charsNo] + strings.Repeat("*", charsNo)
				urlPath = strings.ReplaceAll(urlPath, val, obscureVal)
			}
		}

		logParams := make([]any, 0, 12*2)
		logParams = append(logParams,
			[]any{
				"lvl", "ACCESS",
				"method", r.Method,
				"path", urlPath,
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

		logger.Log(logParams...)
	})
}
