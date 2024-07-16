package xtransport

import (
	"context"

	"github.com/actforgood/xrand"
	"github.com/google/uuid"
)

// CorrelationIDHeaderKey is the header to look for correlation id.
var CorrelationIDHeaderKey = "X-Correlation-Id"

// correlationIDCtxKey is the context key where correlation id is stored.
type correlationIDCtxKey struct{}

// ContextWithCorrelationID returns a new context enriched with correlation id information.
func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDCtxKey{}, correlationID)
}

// CorrelationIDFromContext returns the correlation id stored in the context,
// or an empty value if no correlation id is present in the context.
func CorrelationIDFromContext(ctx context.Context) string {
	if correlationID, found := ctx.Value(correlationIDCtxKey{}).(string); found {
		return correlationID
	}

	return ""
}

// CorrelationIDFactory generates correlation ids.
type CorrelationIDFactory func() string

// UUIDCorrelationIDFactory is a [CorrelationIDFactory] that produces
// UUID based correlation ids.
var UUIDCorrelationIDFactory CorrelationIDFactory = uuid.NewString

// XRandCorrelationIDFactory is a [CorrelationIDFactory] that produces
// random alfanumeric strings of length 32.
var XRandCorrelationIDFactory CorrelationIDFactory = func() string { return xrand.String(32) }
