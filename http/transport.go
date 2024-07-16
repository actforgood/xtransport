// Package http provides a HTTP transport and some utitities.
package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/actforgood/xerr"
	"github.com/actforgood/xlog"

	"github.com/actforgood/xtransport"
)

type httpTransport struct {
	httpSrv *http.Server
	logger  xlog.Logger
	probe   *xtransport.Probe
}

// NewHTTPTransport instantiates a new HTTP transport.
func NewHTTPTransport(
	httpSrv *http.Server,
	logger xlog.Logger,
	probe *xtransport.Probe,
) xtransport.Transport {
	return httpTransport{
		httpSrv: httpSrv,
		logger:  logger,
		probe:   probe,
	}
}

// StartAsync starts the HTTP server. It listens for new connections and messages.
func (ht httpTransport) StartAsync(_ context.Context, errorsChan chan<- error) {
	go func() {
		if ht.probe != nil {
			ht.probe.SetReady(true)
		}
		ht.logger.Info(xlog.MessageKey, "HTTP server starting", "address", ht.httpSrv.Addr)
		if err := ht.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errorsChan <- xerr.Wrap(err, "HTTP server could not listen for connections")
		}
	}()
}

// Shutdown shuts down the HTTP server.
func (ht httpTransport) Shutdown(ctx context.Context) error {
	if ht.probe != nil {
		ht.probe.SetReady(false)
	}
	if err := ht.httpSrv.Shutdown(ctx); err != nil {
		return xerr.Wrap(err, "could not shutdown HTTP server")
	}

	return nil
}

// Health can be registered as health probe endpoint.
func Health(probe *xtransport.Probe) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if probe.IsReady() {
			w.Write([]byte("OK"))

			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("NOTOK"))
	}
}
