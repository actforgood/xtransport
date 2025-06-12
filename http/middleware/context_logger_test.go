package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/actforgood/xlog"

	"github.com/actforgood/xtransport/http/middleware"
	"github.com/actforgood/xtransport/testing/assert"
)

func TestRetry(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		nextHandlerCallsCnt int
		nextHandler         = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextHandlerCallsCnt++
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(t.Name()))

			// next handler should have the log on request context
			xlog.LoggerFromContext(r.Context()).Info(
				"testName", t.Name(),
			)
		})
		logger       = xlog.NewMockLogger()
		logKeyValues []any
		req          = httptest.NewRequest(http.MethodGet, "http://example.com/foo/bar", nil)
		w            = httptest.NewRecorder()
	)
	logger.SetLogCallback(xlog.LevelInfo, func(keyValues ...any) {
		logKeyValues = keyValues
	})
	req.Header.Set("User-Agent", t.Name()+"/1.1")

	// act
	middleware.ContextWithLogger(nextHandler, logger).ServeHTTP(w, req)

	// assert
	assert.Equal(t, 1, nextHandlerCallsCnt)
	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	respBody, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, t.Name(), string(respBody))
	if assert.Equal(t, 1, logger.LogCallsCount(xlog.LevelInfo)) {
		assert.Equal(t, 2, len(logKeyValues))
		assert.Equal(t, t.Name(), getLogValue("testName", logKeyValues...))
	}
}
