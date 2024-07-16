package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/actforgood/xlog"

	"github.com/actforgood/xtransport"
	"github.com/actforgood/xtransport/http/middleware"
	"github.com/actforgood/xtransport/testing/assert"
)

func TestAccessLog(t *testing.T) {
	t.Parallel()

	t.Run("check basic log information", testAccessLogBasic)
	t.Run("check extended log information", testAccessLogExtended)
	t.Run("check certain method is skipped", testAccessLogSkipMethod)
	t.Run("check path value is obfuscated", testAccessLogObfuscatePathValue)
}

func testAccessLogBasic(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		nextHandlerCallsCnt int
		nextHandler         = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextHandlerCallsCnt++
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(t.Name()))
		})
		logger       = xlog.NewMockLogger()
		logKeyValues []any
		req          = httptest.NewRequest(http.MethodGet, "http://example.com/foo/bar", nil)
		w            = httptest.NewRecorder()
	)
	logger.SetLogCallback(xlog.LevelNone, func(keyValues ...any) {
		logKeyValues = keyValues
	})
	req.Header.Set("User-Agent", "TestAccessLog/1.1")

	// act
	middleware.AccessLog(nextHandler, logger).ServeHTTP(w, req)

	// assert
	assert.Equal(t, 1, nextHandlerCallsCnt)
	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	respBody, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, t.Name(), string(respBody))
	if assert.Equal(t, 1, logger.LogCallsCount(xlog.LevelNone)) {
		assert.Equal(t, 16, len(logKeyValues))
		assert.Equal(t, "ACCESS", getLogValue("lvl", logKeyValues...))
		assert.Equal(t, http.MethodGet, getLogValue("method", logKeyValues...))
		assert.Equal(t, "/foo/bar", getLogValue("path", logKeyValues...))
		assert.Equal(t, http.StatusForbidden, getLogValue("statusCode", logKeyValues...))
		assert.Equal(t, len(t.Name()), getLogValue("respBodyLength", logKeyValues...))
		assert.Equal(t, "192.0.2.1", getLogValue("ip", logKeyValues...))
		assert.Equal(t, "TestAccessLog/1.1", getLogValue("userAgent", logKeyValues...))
		if took, ok := getLogValue("took", logKeyValues...).(string); assert.True(t, ok) {
			tookDuration, err := time.ParseDuration(took)
			assert.Nil(t, err)
			assert.True(t, tookDuration > 0)
			assert.True(t, tookDuration < 5*time.Second)
		}
	}
}

func testAccessLogExtended(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		nextHandlerCallsCnt int
		nextHandler         = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextHandlerCallsCnt++
			w.Write([]byte(t.Name()))
		})
		logger       = xlog.NewMockLogger()
		logKeyValues []any
		req          = httptest.NewRequest(http.MethodDelete, "http://example.com/foo/bar?id=123", nil)
		w            = httptest.NewRecorder()
	)
	logger.SetLogCallback(xlog.LevelNone, func(keyValues ...any) {
		logKeyValues = keyValues
	})
	req.Header.Set("User-Agent", "TestAccessLog/1.2")
	req.Header.Set("Content-Length", "100")
	req = req.WithContext(xtransport.ContextWithCorrelationID(req.Context(), "abcd-1234"))
	w.Header().Set("Content-Length", "200")

	// act
	middleware.AccessLog(nextHandler, logger).ServeHTTP(w, req)

	// assert
	assert.Equal(t, 1, nextHandlerCallsCnt)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	respBody, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, t.Name(), string(respBody))
	if assert.Equal(t, 1, logger.LogCallsCount(xlog.LevelNone)) {
		assert.Equal(t, 22, len(logKeyValues))
		assert.Equal(t, "ACCESS", getLogValue("lvl", logKeyValues...))
		assert.Equal(t, http.MethodDelete, getLogValue("method", logKeyValues...))
		assert.Equal(t, "/foo/bar", getLogValue("path", logKeyValues...))
		assert.Equal(t, "id=123", getLogValue("query", logKeyValues...))
		assert.Equal(t, http.StatusOK, getLogValue("statusCode", logKeyValues...))
		assert.Equal(t, 100, getLogValue("reqContentLength", logKeyValues...))
		assert.Equal(t, 200, getLogValue("respContentLength", logKeyValues...))
		assert.Equal(t, "192.0.2.1", getLogValue("ip", logKeyValues...))
		assert.Equal(t, "TestAccessLog/1.2", getLogValue("userAgent", logKeyValues...))
		if took, ok := getLogValue("took", logKeyValues...).(string); assert.True(t, ok) {
			tookDuration, err := time.ParseDuration(took)
			assert.Nil(t, err)
			assert.True(t, tookDuration > 0)
			assert.True(t, tookDuration < 5*time.Second)
		}
		assert.Equal(t, "abcd-1234", getLogValue("correlationId", logKeyValues...))
	}
}

func testAccessLogSkipMethod(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		nextHandlerCallsCnt int
		nextHandler         = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextHandlerCallsCnt++
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(t.Name()))
		})
		logger = xlog.NewMockLogger()
		req    = httptest.NewRequest(http.MethodGet, "http://example.com/skip/method/get", nil)
		w      = httptest.NewRecorder()
	)
	opts := middleware.AccessLogOpts{SkipMethods: []string{http.MethodGet}}

	// act
	middleware.AccessLog(nextHandler, logger, opts).ServeHTTP(w, req)

	// assert
	assert.Equal(t, 1, nextHandlerCallsCnt)
	assert.Equal(t, http.StatusConflict, w.Result().StatusCode)
	respBody, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, t.Name(), string(respBody))
	assert.Equal(t, 0, logger.LogCallsCount(xlog.LevelNone))
}

func testAccessLogObfuscatePathValue(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		nextHandlerCallsCnt int
		nextHandler         = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextHandlerCallsCnt++
			w.WriteHeader(http.StatusNoContent)
		})
		logger       = xlog.NewMockLogger()
		logKeyValues []any
		// suppose url would be registered as http://example.com/auth/{authToken}/customer/{customerId}
		req = httptest.NewRequest(http.MethodDelete, "http://example.com/auth/abc-xyz-very-secret/customer/1234", nil)
		w   = httptest.NewRecorder()
	)
	logger.SetLogCallback(xlog.LevelNone, func(keyValues ...any) {
		logKeyValues = keyValues
	})
	req.Header.Set("User-Agent", "TestAccessLog/1.4")
	req.SetPathValue("authToken", "abc-xyz-very-secret")
	req.SetPathValue("customerId", "1234")
	opts := middleware.AccessLogOpts{ObfuscatePathValues: []string{"authToken"}}

	// act
	middleware.AccessLog(nextHandler, logger, opts).ServeHTTP(w, req)

	// assert
	assert.Equal(t, 1, nextHandlerCallsCnt)
	assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
	respBody, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, "", string(respBody))
	if assert.Equal(t, 1, logger.LogCallsCount(xlog.LevelNone)) {
		assert.Equal(t, 16, len(logKeyValues))
		assert.Equal(t, "ACCESS", getLogValue("lvl", logKeyValues...))
		assert.Equal(t, http.MethodDelete, getLogValue("method", logKeyValues...))
		assert.Equal(t, "/auth/abc-xyz-ver********/customer/1234", getLogValue("path", logKeyValues...))
		assert.Equal(t, http.StatusNoContent, getLogValue("statusCode", logKeyValues...))
		assert.Equal(t, 0, getLogValue("respBodyLength", logKeyValues...))
		assert.Equal(t, "192.0.2.1", getLogValue("ip", logKeyValues...))
		assert.Equal(t, "TestAccessLog/1.4", getLogValue("userAgent", logKeyValues...))
		if took, ok := getLogValue("took", logKeyValues...).(string); assert.True(t, ok) {
			tookDuration, err := time.ParseDuration(took)
			assert.Nil(t, err)
			assert.True(t, tookDuration > 0)
			assert.True(t, tookDuration < 5*time.Second)
		}
	}
}

// getLogValue finds a value for a log key, if any.
func getLogValue(key string, keyValues ...any) any {
	for i := 0; i <= len(keyValues)-2; i += 2 {
		if keyValues[i] == key && len(keyValues) > i {
			return keyValues[i+1]
		}
	}

	return "key-not-found"
}
