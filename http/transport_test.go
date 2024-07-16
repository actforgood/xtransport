package http_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/actforgood/xlog"
	"github.com/actforgood/xtransport"
	httpTransport "github.com/actforgood/xtransport/http"
	"github.com/actforgood/xtransport/testing/assert"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	t.Run("is healthy", restHealthIsHealthy)
	t.Run("is not healthy", restHealthIsNotHealthy)
}

func restHealthIsHealthy(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		req   = httptest.NewRequest(http.MethodGet, "http://example.com/health", nil)
		w     = httptest.NewRecorder()
		probe = new(xtransport.Probe)
	)
	probe.SetReady(true)

	// act
	httpTransport.Health(probe)(w, req)

	// assert
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	respBody, err := io.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	assert.Equal(t, "OK", string(respBody))
}

func restHealthIsNotHealthy(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		req   = httptest.NewRequest(http.MethodGet, "http://example.com/health", nil)
		w     = httptest.NewRecorder()
		probe = new(xtransport.Probe)
	)

	// act
	httpTransport.Health(probe)(w, req)

	// assert
	assert.Equal(t, http.StatusServiceUnavailable, w.Result().StatusCode)
	respBody, err := io.ReadAll(w.Result().Body)
	assert.Nil(t, err)
	assert.Equal(t, "NOTOK", string(respBody))
}

func TestHTTPTransport(t *testing.T) {
	t.Parallel()

	var (
		probe   = new(xtransport.Probe)
		mux     = http.NewServeMux()
		httpSrv = &http.Server{
			Addr:         "127.0.0.1:5568",
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  10 * time.Second,
			IdleTimeout:  2 * time.Minute,
			Handler:      mux,
		}
		subject = httpTransport.NewHTTPTransport(httpSrv, xlog.NopLogger{}, probe)
		errChan = make(chan error, 1)
		ctx     = context.Background()
	)
	mux.HandleFunc("/health", httpTransport.Health(probe))
	go func() {
		err := <-errChan
		assert.Nil(t, err)
	}()
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// act
	subject.StartAsync(ctx, errChan)

	// assert
	for {
		if probe.IsReady() {
			cancel()
		}
		if timeoutCtx.Err() != nil {
			break
		}
	}
	assert.True(t, probe.IsReady())

	// act
	resp, err := http.Get("http://" + httpSrv.Addr + "/health") // nolint:noctx

	// assert
	if assert.Nil(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()
	}

	// act
	err = subject.Shutdown(ctx)

	// assert
	assert.Nil(t, err)
	assert.Equal(t, false, probe.IsReady())
}
