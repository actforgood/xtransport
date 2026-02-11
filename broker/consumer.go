package broker

import "context"

const (
	ConsumeResultAck byte = iota + 1
	ConsumeResultNack
	ConsumeResultNackRequeue
)

type Consumer interface {
	Props() Props
	Consume(context.Context, Message) byte
}
