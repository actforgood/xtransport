package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/actforgood/xtransport"
	"github.com/actforgood/xtransport/http/middleware"
	"github.com/actforgood/xtransport/testing/assert"
)

func TestCorrelationID(t *testing.T) {
	t.Parallel()

	t.Run("new correlation id is added on context, if not present", testCorrelationIDMissing)
	t.Run("existing correlation id is taken over, if present", testCorrelationIDFound)
}

func testCorrelationIDMissing(t *testing.T) {
	t.Parallel()

	// arrange
	var correlationID string
	var nextHandlerCallsCnt int
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCallsCnt++
		w.Write([]byte(t.Name()))
		correlationID = xtransport.CorrelationIDFromContext(r.Context())
	})
	correlationIDFactory := func() string { return "test-correlation-id-1234" }
	req := httptest.NewRequest(http.MethodGet, "http://example.com/noCorrelationId", nil)
	w := httptest.NewRecorder()

	// act
	middleware.CorrelationID(nextHandler, correlationIDFactory).ServeHTTP(w, req)

	// assert
	assert.Equal(t, 1, nextHandlerCallsCnt)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	respBody, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, t.Name(), string(respBody))
	assert.Equal(t, "test-correlation-id-1234", correlationID)
	assert.Equal(t, correlationID, w.Result().Header.Get(xtransport.CorrelationIDHeaderKey))
}

func testCorrelationIDFound(t *testing.T) {
	t.Parallel()

	// arrange
	var correlationID string
	var nextHandlerCallsCnt int
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCallsCnt++
		w.Write([]byte(t.Name()))
		correlationID = xtransport.CorrelationIDFromContext(r.Context())
	})
	correlationIDFactory := func() string { return "test-correlation-id-abcd" }
	req := httptest.NewRequest(http.MethodGet, "http://example.com/withCorrelationId", nil)
	req.Header.Set(xtransport.CorrelationIDHeaderKey, "test-correlation-id-wxyz")
	w := httptest.NewRecorder()

	// act
	middleware.CorrelationID(nextHandler, correlationIDFactory).ServeHTTP(w, req)

	// assert
	assert.Equal(t, 1, nextHandlerCallsCnt)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	respBody, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, t.Name(), string(respBody))
	assert.Equal(t, "test-correlation-id-wxyz", correlationID)
	assert.Equal(t, correlationID, w.Result().Header.Get(xtransport.CorrelationIDHeaderKey))
}
