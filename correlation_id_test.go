package xtransport_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/google/uuid"

	"github.com/actforgood/xtransport"
	"github.com/actforgood/xtransport/testing/assert"
)

func TestContextWithCorrelationID(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		expectedCorrelationID = "test-correlation-foo-bar-123-abc"
		ctx                   = context.Background()
	)

	// act
	defaultCorrelationID := xtransport.CorrelationIDFromContext(ctx)
	newCtx := xtransport.ContextWithCorrelationID(ctx, expectedCorrelationID)
	actualCorrelationID := xtransport.CorrelationIDFromContext(newCtx)

	// assert
	assert.Equal(t, "", defaultCorrelationID)
	assert.Equal(t, expectedCorrelationID, actualCorrelationID)
}

func TestUUIDCorrelationIDFactory(t *testing.T) {
	t.Parallel()

	// act
	correlationID1 := xtransport.UUIDCorrelationIDFactory()
	correlationID2 := xtransport.UUIDCorrelationIDFactory()

	// assert
	_, err1 := uuid.Parse(correlationID1)
	assert.Nil(t, err1)
	_, err2 := uuid.Parse(correlationID2)
	assert.Nil(t, err2)
	assert.True(t, correlationID1 != correlationID2)
}

func TestXRandCorrelationIDFactory(t *testing.T) {
	t.Parallel()

	// arrange
	reg := regexp.MustCompile("^[a-zA-Z0-9]{32}$")

	// act
	correlationID1 := xtransport.XRandCorrelationIDFactory()
	correlationID2 := xtransport.XRandCorrelationIDFactory()

	// assert
	assert.True(t, reg.MatchString(correlationID1))
	assert.True(t, reg.MatchString(correlationID2))
	assert.True(t, correlationID1 != correlationID2)
}

func BenchmarkUUIDCorrelationIDFactory(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = xtransport.UUIDCorrelationIDFactory()
	}
}

func BenchmarkXRandCorrelationIDFactory(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = xtransport.XRandCorrelationIDFactory()
	}
}
