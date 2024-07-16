package middleware_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/actforgood/xlog"

	"github.com/actforgood/xtransport/http/middleware"
	"github.com/actforgood/xtransport/testing/assert"
)

func TestRecover(t *testing.T) {
	t.Parallel()

	t.Run("next handler is successfully triggered", testRecoverNoPanic)
	t.Run("panic is catched and 500 response returned", testRecoverWithPanic)
}

func testRecoverNoPanic(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		nextHandlerCallsCnt int
		nextHandler         = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextHandlerCallsCnt++
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(t.Name()))
		})
		logger = xlog.NewMockLogger()
		req    = httptest.NewRequest(http.MethodGet, "http://example.com/noPanic", nil)
		w      = httptest.NewRecorder()
	)

	// act
	middleware.Recover(nextHandler, logger).ServeHTTP(w, req)

	// assert
	assert.Equal(t, 1, nextHandlerCallsCnt)
	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	respBody, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, t.Name(), string(respBody))
	assert.Equal(t, 0, logger.LogCallsCount(xlog.LevelError))
}

func testRecoverWithPanic(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		nextHandlerCallsCnt int
		nextHandler         = http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			nextHandlerCallsCnt++
			panic(errors.New("intentionally triggered panic"))
		})
		logger = xlog.NewMockLogger()
		req    = httptest.NewRequest(http.MethodGet, "http://example.com/panicRoute", nil)
		w      = httptest.NewRecorder()
	)

	// act
	middleware.Recover(nextHandler, logger).ServeHTTP(w, req)

	// assert
	assert.Equal(t, 1, nextHandlerCallsCnt)
	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	respBody, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, "an unexpected error occurred, please try again later", string(respBody))
	assert.Equal(t, 1, logger.LogCallsCount(xlog.LevelError))
}
