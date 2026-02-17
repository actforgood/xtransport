// Package mock provides a mock implementations needed for testing.
package mock

import (
	"context"
	"log/slog"
	"sync"
)

// KeyNotFound is the type of value returned by [SlogHandler.ValueAt]
// in case the searched key is not found.
type KeyNotFound struct{}

// Any can be passed to [MockHandler.ValueAt] as call index,
// in case the order is not known/important/multiple logs happen concurrently.
const Any uint = 0

// SlogHandler is a mock for [slog.Handler],
// which allows us to intercept the log calls and assert on the log content.
type SlogHandler struct {
	loggedKeyValues [][]slog.Attr
	logCallsCnt     map[slog.Level]uint32
	attrs           []slog.Attr
	mu              sync.RWMutex
}

// NewSlogHandler instantiates a new SlogHandler object.
func NewSlogHandler() *SlogHandler {
	return &SlogHandler{
		logCallsCnt: make(map[slog.Level]uint32, 4),
		attrs:       make([]slog.Attr, 0, 4),
	}
}

// Handle intercepts the log calls and stores the log content for later assertions.
func (mock *SlogHandler) Handle(_ context.Context, record slog.Record) error {
	mock.mu.Lock()
	defer mock.mu.Unlock()

	lvl := record.Level
	mock.logCallsCnt[lvl]++

	attrs := make([]slog.Attr, 0, len(mock.attrs)+record.NumAttrs()+1)
	attrs = append(attrs, slog.Attr{Key: "msg", Value: slog.StringValue(record.Message)})
	attrs = append(attrs, mock.attrs...)
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)

		return true
	})
	mock.loggedKeyValues = append(mock.loggedKeyValues, attrs)

	return nil
}

// Enabled returns true, as we want to intercept all log calls.
func (mock *SlogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// WithAttrs adds the given attributes to the handler, so that they are included in all subsequent log calls.
func (mock *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	mock.mu.Lock()
	defer mock.mu.Unlock()

	mock.attrs = append(mock.attrs, attrs...)

	return mock
}

// WithGroup is a no-op for SlogHandler, as groups are not relevant for our tests.
func (mock *SlogHandler) WithGroup(_ string) slog.Handler {
	return mock
}

// ValueAt returns the value for a key at given call, in case no callback was set.
// Calls are positive numbers (starting with 1).
// If the order of the calls is not known/important/multiple logs happen concurrently,
// you can use [Any].
// If the key is not found,  [KeyNotFound] is returned.
func (mock *SlogHandler) ValueAt(callNo uint, forKey string) any {
	mock.mu.RLock()
	defer mock.mu.RUnlock()

	if callNo == Any {
		for call := range len(mock.loggedKeyValues) {
			value := mock.valueAt(call+1, forKey)
			if _, isNotFound := value.(KeyNotFound); !isNotFound {
				return value
			}
		}

		return KeyNotFound{}
	}

	return mock.valueAt(int(callNo), forKey)
}

func (mock *SlogHandler) valueAt(callNo int, forKey any) any {
	if len(mock.loggedKeyValues) >= callNo {
		for _, attr := range mock.loggedKeyValues[callNo-1] {
			if attr.Key == forKey {
				return attr.Value.Any()
			}
		}
	}

	return KeyNotFound{}
}

// LogCallsCount returns the no. of times Error/Warn/Info/Debug/Log was called.
// Differentiate methods calls count by passing appropriate level.
func (mock *SlogHandler) LogCallsCount(lvl slog.Level) int {
	mock.mu.RLock()
	defer mock.mu.RUnlock()

	return int(mock.logCallsCnt[lvl])
}
