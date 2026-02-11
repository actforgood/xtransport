package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
			time.Sleep(2 * time.Millisecond)
		})
		logger = xlog.NewMockLogger()
		req    = httptest.NewRequest(http.MethodGet, "http://example.com/foo/bar", nil)
		w      = httptest.NewRecorder()
	)
	req.Header.Set("User-Agent", "TestAccessLog/1.1")

	// act
	middleware.AccessLog(nextHandler, logger).ServeHTTP(w, req)

	// assert
	assert.Equal(t, 1, nextHandlerCallsCnt)
	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	respBody, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, t.Name(), string(respBody))
	if assert.Equal(t, 1, logger.LogCallsCount(xlog.LevelNone)) {
		assert.Equal(t, "ACCESS", logger.ValueAt(1, "lvl"))
		assert.Equal(t, http.MethodGet, logger.ValueAt(1, "method"))
		assert.Equal(t, "/foo/bar", logger.ValueAt(1, "path"))
		assert.Equal(t, http.StatusForbidden, logger.ValueAt(1, "statusCode"))
		assert.Equal(t, len(t.Name()), logger.ValueAt(1, "respBodyLength"))
		assert.Equal(t, "192.0.2.1", logger.ValueAt(1, "ip"))
		assert.Equal(t, "TestAccessLog/1.1", logger.ValueAt(1, "userAgent"))
		if took, ok := logger.ValueAt(1, "took").(string); assert.True(t, ok) {
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
			time.Sleep(2 * time.Millisecond)
		})
		logger = xlog.NewMockLogger()
		req    = httptest.NewRequest(http.MethodDelete, "http://example.com/foo/bar?id=123", nil)
		w      = httptest.NewRecorder()
	)
	req.Header.Set("User-Agent", "TestAccessLog/1.2")
	req.Header.Set("Content-Length", "100")
	req = req.WithContext(xtransport.ContextWithCorrelationID(req.Context(), "abcd-1234"))
	req.URL.User = url.User("john-doe")
	w.Header().Set("Content-Length", "200")

	// act
	middleware.AccessLog(nextHandler, logger).ServeHTTP(w, req)

	// assert
	assert.Equal(t, 1, nextHandlerCallsCnt)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	respBody, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, t.Name(), string(respBody))
	if assert.Equal(t, 1, logger.LogCallsCount(xlog.LevelNone)) {
		assert.Equal(t, "ACCESS", logger.ValueAt(1, "lvl"))
		assert.Equal(t, http.MethodDelete, logger.ValueAt(1, "method"))
		assert.Equal(t, "/foo/bar", logger.ValueAt(1, "path"))
		assert.Equal(t, "id=123", logger.ValueAt(1, "query"))
		assert.Equal(t, http.StatusOK, logger.ValueAt(1, "statusCode"))
		assert.Equal(t, 100, logger.ValueAt(1, "reqContentLength"))
		assert.Equal(t, 200, logger.ValueAt(1, "respContentLength"))
		assert.Equal(t, "192.0.2.1", logger.ValueAt(1, "ip"))
		assert.Equal(t, "TestAccessLog/1.2", logger.ValueAt(1, "userAgent"))
		assert.Equal(t, "john-doe", logger.ValueAt(1, "authUsername"))
		if took, ok := logger.ValueAt(1, "took").(string); assert.True(t, ok) {
			tookDuration, err := time.ParseDuration(took)
			assert.Nil(t, err)
			assert.True(t, tookDuration > 0)
			assert.True(t, tookDuration < 5*time.Second)
		}
		assert.Equal(t, "abcd-1234", logger.ValueAt(1, "correlationId"))
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
			time.Sleep(2 * time.Millisecond)
		})
		logger = xlog.NewMockLogger()
		// suppose url would be registered as http://example.com/auth/{authToken}/customer/{customerId}
		req = httptest.NewRequest(http.MethodDelete, "http://example.com/auth/abc-xyz-very-secret/customer/1234", nil)
		w   = httptest.NewRecorder()
	)
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
		assert.Equal(t, "ACCESS", logger.ValueAt(1, "lvl"))
		assert.Equal(t, http.MethodDelete, logger.ValueAt(1, "method"))
		assert.Equal(t, "/auth/abc-xyz-ver********/customer/1234", logger.ValueAt(1, "path"))
		assert.Equal(t, http.StatusNoContent, logger.ValueAt(1, "statusCode"))
		assert.Equal(t, 0, logger.ValueAt(1, "respBodyLength"))
		assert.Equal(t, "192.0.2.1", logger.ValueAt(1, "ip"))
		assert.Equal(t, "TestAccessLog/1.4", logger.ValueAt(1, "userAgent"))
		if took, ok := logger.ValueAt(1, "took").(string); assert.True(t, ok) {
			tookDuration, err := time.ParseDuration(took)
			assert.Nil(t, err)
			assert.True(t, tookDuration > 0)
			assert.True(t, tookDuration < 5*time.Second)
		}
	}
}
