package broker

import "context"

const (
	// ConsumeResultAck indicates that the message was successfully consumed
	// and can be acknowledged.
	ConsumeResultAck byte = iota + 1
	// ConsumeResultNack indicates that the message was not successfully consumed
	// and should be negatively acknowledged.
	ConsumeResultNack
	// ConsumeResultNackRequeue indicates that the message was not successfully consumed
	// and should be negatively acknowledged and requeued for later processing.
	ConsumeResultNackRequeue
)

// Consumer is the interface that wraps the basic Consume method.
type Consumer interface {
	// Props returns the properties of the consumer.
	Props() Props
	// Consume processes the given message and returns the result of the consumption.
	Consume(context.Context, Message) byte
}
