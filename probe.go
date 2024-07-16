package xtransport

import "sync"

// Probe can be used as a ready/alive/health flag.
// It is concurrent safe to use.
type Probe struct {
	isReady bool
	mu      sync.RWMutex
}

// SetReady sets the readiness state.
func (p *Probe) SetReady(isReady bool) {
	p.mu.Lock()
	p.isReady = isReady
	p.mu.Unlock()
}

// IsReady returns probe readiness.
func (p *Probe) IsReady() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.isReady
}
