package http_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpTransport "github.com/actforgood/xtransport/http"
	"github.com/actforgood/xtransport/testing/assert"
)

func TestGetRequestBody(t *testing.T) {
	t.Parallel()

	t.Run("limit is reached", testGetRequestBodyLimitReached)
	t.Run("limit is not reached", testGetRequestBodyLimitNotReached)
}

func testGetRequestBodyLimitReached(t *testing.T) {
	t.Parallel()

	// arrange
	req := httptest.NewRequest(
		http.MethodPost,
		"http://example.com/baz",
		strings.NewReader("foo"),
	)

	// act
	reader := httpTransport.GetRequestBody(nil, req, 1)
	_, err := io.ReadAll(reader)

	// assert
	if assert.NotNil(t, err) {
		var mbErr *http.MaxBytesError
		assert.True(t, errors.As(err, &mbErr))
	}
}

func testGetRequestBodyLimitNotReached(t *testing.T) {
	t.Parallel()

	// arrange
	req := httptest.NewRequest(
		http.MethodPost,
		"http://example.com/baz",
		strings.NewReader("foo"),
	)

	// act
	reader := httpTransport.GetRequestBody(nil, req)
	_, err := io.ReadAll(reader)

	// assert
	assert.Nil(t, err)
}

func TestGetClientIP(t *testing.T) {
	t.Parallel()

	t.Run("returns real ip", testGetClientIPWithRealIP)
	t.Run("returns first ip from XFF", testGetClientIPWithXFF)
	t.Run("returns remote address", testGetClientIPWithRemoteAddr)
}

func testGetClientIPWithRealIP(t *testing.T) {
	t.Parallel()

	// arrange
	req := httptest.NewRequest(http.MethodGet, "http://example.com/realip", nil)
	expectedIP := "1.2.3.4"
	req.Header.Set("X-Real-Ip", expectedIP)
	req.Header.Set("X-Forwarded-For", "8.8.8.8, 9.9.9.9")

	// act
	actualIP := httpTransport.GetClientIP(req)

	// assert
	assert.Equal(t, expectedIP, actualIP.String())
}

func testGetClientIPWithXFF(t *testing.T) {
	t.Parallel()

	// arrange
	req := httptest.NewRequest(http.MethodGet, "http://example.com/xff", nil)
	expectedIP := "8.8.8.8"
	req.Header.Set("X-Forwarded-For", expectedIP+", 9.9.9.9")

	// act
	actualIP := httpTransport.GetClientIP(req)

	// assert
	assert.Equal(t, expectedIP, actualIP.String())
}

func testGetClientIPWithRemoteAddr(t *testing.T) {
	t.Parallel()

	// arrange
	req := httptest.NewRequest(http.MethodGet, "http://example.com/remoteAddr", nil)
	expectedIP := "192.0.2.1"

	// act
	actualIP := httpTransport.GetClientIP(req)

	// assert
	assert.Equal(t, expectedIP, actualIP.String())
}
