package xtransport

import (
	"context"
)

// Transport is the contract for a transport.
type Transport interface {
	// StartAsync starts the transport asynchronous.
	// Any error received is passed to the error channel passed as second parameter.
	StartAsync(context.Context, chan<- error)

	// Shutdown stops the transport.
	Shutdown(context.Context) error
}
