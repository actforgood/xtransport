package broker

import (
	"context"
	"sync/atomic"
)

// PublisherMock is a mock for [Publisher]. It can be used in UT.
type PublisherMock struct {
	publishCallsCnt uint32
	publishCallback func(context.Context, Message) error
	closeCallsCnt   uint32
	closeCallback   func() error
}

// Publish mock logic...
func (mock *PublisherMock) Publish(ctx context.Context, msg Message) error {
	atomic.AddUint32(&mock.publishCallsCnt, 1)
	if mock.publishCallback != nil {
		return mock.publishCallback(ctx, msg)
	}

	return nil
}

// Close mock logic...
func (mock *PublisherMock) Close() error {
	atomic.AddUint32(&mock.closeCallsCnt, 1)
	if mock.closeCallback != nil {
		return mock.closeCallback()
	}

	return nil
}

// SetPublishCallback sets the callback to be executed on Publish call.
func (mock *PublisherMock) SetPublishCallback(cb func(context.Context, Message) error) {
	mock.publishCallback = cb
}

// PublishCallsCount returns the no. of times Publish was called.
func (mock *PublisherMock) PublishCallsCount() int {
	return int(atomic.LoadUint32(&mock.publishCallsCnt))
}

// SetCloseCallback sets the callback to be executed on Close call.
func (mock *PublisherMock) SetCloseCallback(cb func() error) {
	mock.closeCallback = cb
}

// CloseCallsCount returns the no. of times Close was called.
func (mock *PublisherMock) CloseCallsCount() int {
	return int(atomic.LoadUint32(&mock.closeCallsCnt))
}
