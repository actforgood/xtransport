package broker

import "context"

type Publisher interface {
	Publish(context.Context, Message) error
	Close() error
}
