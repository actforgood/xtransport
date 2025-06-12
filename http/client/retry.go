package client

import (
	"net/http"
	"time"

	"github.com/actforgood/xerr"
	"github.com/actforgood/xlog"
	"github.com/actforgood/xrand"
	"github.com/actforgood/xtransport"
)

const defaultMaxRetries byte = 3

// Retry is a decorator for a client, to retry request in case of errors.
type Retry struct {
	client     Contract // original client
	maxRetries byte
}

// NewRetry instantiates a new retryable client.
// A maximum number of retries can be specified,
// otherwise a default of 3 is used.
// Retry strategy is an incremental one with a jitter applied,
// meaning first request will happen after ~1s, 2nd after another ~2s,
// 3rd after another ~3s, etc.
func NewRetry(client Contract, maxRetries ...byte) Contract {
	retries := defaultMaxRetries
	if len(maxRetries) > 0 {
		retries = maxRetries[0]
	}

	return Retry{
		client:     client,
		maxRetries: retries,
	}
}

// Do calls original client's Do and retries the request in case of error.
func (retry Retry) Do(r *http.Request) (*http.Response, error) {
	var (
		mErr    xerr.MultiError
		retryNo byte
		resp    *http.Response
		err     error
	)
	for retryNo = range retry.maxRetries + 1 {
		resp, err = retry.client.Do(r)
		if err == nil {
			break
		}
		mErr.AddOnce(err)
		backoff := xrand.Jitter(time.Duration(retryNo+1) * time.Second)
		select {
		case <-time.After(backoff):
			// continue
		case <-r.Context().Done():
			mErr.AddOnce(r.Context().Err())
		}
		if r.Context().Err() != nil {
			break
		}
	}

	if err == nil {
		if mErr.ErrOrNil() != nil {
			xlog.LoggerFromContext(r.Context()).Warn(
				xlog.MsgKey, "client request recovered, but had failed request(s)",
				xlog.ErrorKey, mErr.ErrOrNil(),
				"correlationId", xtransport.CorrelationIDFromContext(r.Context()),
				"endpoint", r.URL.Scheme+"://"+r.URL.Host+r.URL.Path,
				"retriesCount", retryNo,
			)
		}

		return resp, nil
	}

	return resp, mErr.ErrOrNil()
}
