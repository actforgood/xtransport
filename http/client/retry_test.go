package client_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/actforgood/xlog"

	"github.com/actforgood/xtransport"
	"github.com/actforgood/xtransport/http/client"
	"github.com/actforgood/xtransport/testing/assert"
)

func TestRetry(t *testing.T) {
	t.Parallel()

	t.Run("client succeeds without need of retring", testRetrySucceeds)
	t.Run("client recovers", testRetryRecovers)
	t.Run("client fails", testRetryFails)
}

func testRetrySucceeds(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		logger     = xlog.NewMockLogger()
		ctx        = xlog.ContextWithLogger(context.Background(), logger)
		req, _     = http.NewRequestWithContext(ctx, http.MethodGet, "https://bar.baz/foo", nil)
		mockClient = new(client.Mock)
		subject    = client.NewRetry(mockClient)
	)

	// act
	actualResp, actualErr := subject.Do(req)

	// assert
	assert.Equal(t, 1, mockClient.DoCallsCount())
	assert.Nil(t, actualErr)
	assert.NotNil(t, actualResp)
	assert.Equal(t, 0, logger.LogCallsCount(xlog.LevelWarning))
}

func testRetryRecovers(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		logger = xlog.NewMockLogger()
		ctx    = xtransport.ContextWithCorrelationID(
			xlog.ContextWithLogger(context.Background(), logger),
			"cid-123-xyz",
		)
		req, _       = http.NewRequestWithContext(ctx, http.MethodGet, "https://foo.bar/baz?secretKey=abc", nil)
		expectedResp = &http.Response{StatusCode: http.StatusOK}
		mockClient   = new(client.Mock)
		subject      = client.NewRetry(mockClient, 4)
	)

	mockClient.SetDoCallback(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, req, r) // make expectations upon params

		switch mockClient.DoCallsCount() {
		case 1:
			return nil, errors.New("intentionally triggered Do error (initial error)")
		case 2:
			return nil, errors.New("intentionally triggered Do error (retry #1)")
		case 3:
			return nil, errors.New("intentionally triggered Do error (retry #2)")
		case 4:
			return nil, errors.New("intentionally triggered Do error (retry #3)")
		case 5: // on retry no. 4 we recover.
			return &http.Response{StatusCode: http.StatusOK}, nil
		}

		t.Fatal("should not get here")

		return nil, nil
	})
	logger.SetLogCallback(xlog.LevelWarning, func(keyValues ...any) {
		var foundKeys byte
		for i := 0; i < len(keyValues); i += 2 {
			switch keyValues[i] {
			case xlog.MsgKey:
				assert.Equal(t, "client request recovered, but had failed request(s)", keyValues[i+1])
				foundKeys++
			case xlog.ErrorKey:
				assert.NotNil(t, keyValues[i+1])
				foundKeys++
			case "correlationId":
				assert.Equal(t, "cid-123-xyz", keyValues[i+1])
				foundKeys++
			case "endpoint":
				assert.Equal(t, "https://foo.bar/baz", keyValues[i+1])
				foundKeys++
			case "retriesCount":
				assert.Equal(t, byte(4), keyValues[i+1])
				foundKeys++
			}
		}
		assert.Equal(t, byte(5), foundKeys)
	})

	// act
	actualResp, actualErr := subject.Do(req)

	// assert
	assert.Equal(t, 5, mockClient.DoCallsCount())
	assert.Nil(t, actualErr)
	assert.Equal(t, expectedResp, actualResp)
	assert.Equal(t, 1, logger.LogCallsCount(xlog.LevelWarning))
}

func testRetryFails(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		logger      = xlog.NewMockLogger()
		ctx         = xlog.ContextWithLogger(context.Background(), logger)
		req, _      = http.NewRequestWithContext(ctx, http.MethodGet, "https://foo.bar", nil)
		expectedErr = errors.New("intentionally triggered Do error")
		mockClient  = new(client.Mock)
		subject     = client.NewRetry(mockClient)
	)
	mockClient.SetDoCallback(func(_ *http.Request) (*http.Response, error) {
		return nil, expectedErr
	})

	// act
	actualResp, actualErr := subject.Do(req)

	// assert
	assert.Equal(t, 4, mockClient.DoCallsCount())
	assert.True(t, errors.Is(actualErr, expectedErr))
	assert.Nil(t, actualResp)
	assert.Equal(t, 0, logger.LogCallsCount(xlog.LevelWarning))
}
