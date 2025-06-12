package client_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/actforgood/xtransport/http/client"
	"github.com/actforgood/xtransport/testing/assert"
)

func TestMock(t *testing.T) {
	t.Parallel()

	// arrange
	var _ client.Contract = (*client.Mock)(nil) // test it implements its contract
	var (
		subject       = new(client.Mock)
		req1, _       = http.NewRequestWithContext(context.Background(), "https://example.com/foo", http.MethodGet, nil)
		req2, _       = http.NewRequestWithContext(context.Background(), "https://example.com/bar/123", http.MethodDelete, nil)
		expectedResp1 = &http.Response{StatusCode: http.StatusNotFound}
		expectedErr2  = errors.New("intentionally triggered error")
	)
	subject.SetDoCallback(func(r *http.Request) (*http.Response, error) {
		switch subject.DoCallsCount() {
		case 1:
			assert.Equal(t, req1, r)

			return expectedResp1, nil
		case 2:
			assert.Equal(t, req2, r)

			return nil, expectedErr2
		}

		t.Fatal("should not get here")

		return nil, nil
	})

	// act
	actualResp1, actualErr1 := subject.Do(req1)
	actualResp2, actualErr2 := subject.Do(req1)

	// assert
	assert.Equal(t, 2, subject.DoCallsCount())
	assert.Nil(t, actualErr1)
	assert.Equal(t, expectedResp1, actualResp1)
	assert.True(t, errors.Is(actualErr2, expectedErr2))
	assert.Nil(t, actualResp2)
}
