package broker

import "context"

// Publisher is the interface that wraps the basic Publish method.
type Publisher interface {
	// Publish sends the given message to the broker.
	Publish(context.Context, Message) error
	// Close closes the publisher and releases any resources associated with it.
	Close() error
}
