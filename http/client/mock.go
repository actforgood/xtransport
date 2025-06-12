package client

import (
	"net/http"
	"sync/atomic"
)

// Mock is a mock for [Contract], to be used in UT.
type Mock struct {
	doCallsCnt uint32
	doCallback func(*http.Request) (*http.Response, error)
}

// Do mock logic.
func (mock *Mock) Do(r *http.Request) (*http.Response, error) {
	atomic.AddUint32(&mock.doCallsCnt, 1)
	if mock.doCallback != nil {
		return mock.doCallback(r)
	}

	return &http.Response{
		Status:        "200 OK",
		StatusCode:    http.StatusOK,
		Proto:         "HTTP/1.1",
		ContentLength: 0,
		ProtoMajor:    1,
		ProtoMinor:    1,
		Request:       r,
	}, nil
}

// SetDoCallback sets the callback to be executed inside Do.
// You can make assertions upon passed parameter(s) this way,
// and control returned values.
func (mock *Mock) SetDoCallback(callback func(*http.Request) (*http.Response, error)) {
	mock.doCallback = callback
}

// DoCallsCount returns the no. of times Do() method was called.
func (mock *Mock) DoCallsCount() int {
	return int(atomic.LoadUint32(&mock.doCallsCnt))
}
