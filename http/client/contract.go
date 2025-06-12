// Package client provides http client contract and utitlities.
package client

import (
	"net/http"
)

// Contract is a contract for [http.Client].
type Contract interface {
	// Do makes a request.
	// See [http.Client.Do] for original documentation.
	Do(*http.Request) (*http.Response, error)
}
