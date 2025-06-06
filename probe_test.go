package xtransport_test

import (
	"sync"
	"testing"

	"github.com/actforgood/xrand"
	"github.com/actforgood/xtransport"

	"github.com/actforgood/xtransport/testing/assert"
)

func TestProbe(t *testing.T) {
	t.Parallel()

	// arrange
	subject := new(xtransport.Probe)

	// act
	defaultState := subject.IsReady()
	subject.SetReady(true)
	actualState := subject.IsReady()

	// assert
	assert.Equal(t, false, defaultState)
	assert.True(t, actualState)
}

// TestProbe_concurrency accesses ready state of Probe
// in a concurrent environment. This test does not assert
// something in particular, it is aimed to be run with '--race'
// flag and see that no data race occurs.
func TestProbe_concurrency(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject = new(xtransport.Probe)
		wg      sync.WaitGroup
	)
	wg.Add(20)

	// act
	for range 10 {
		go func() {
			rand := xrand.Intn(2)
			if rand == 0 {
				subject.SetReady(false)
			} else {
				subject.SetReady(true)
			}
			wg.Done()
		}()
	}
	for range 10 {
		go func() {
			_ = subject.IsReady()
			wg.Done()
		}()
	}
	wg.Wait()
}
