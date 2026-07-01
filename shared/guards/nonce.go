package guards

import (
	"context"
	"sync"
	"time"
)

// NonceCache tracks recently seen nonces to prevent replay attacks.
// v1: in-memory, suitable for a single collector instance.
// Swap to a Redis-backed implementation for multi-replica deployments.
type NonceCache interface {
	SeenRecently(ctx context.Context, agentID, nonce string, window time.Duration) (bool, error)
	Record(ctx context.Context, agentID, nonce string, ttl time.Duration) error
}

type memNonceCache struct {
	mu      sync.Mutex
	entries map[string]time.Time
}

func NewMemoryNonceCache() NonceCache {
	c := &memNonceCache{entries: make(map[string]time.Time)}
	go c.sweep()
	return c
}

func (c *memNonceCache) cacheKey(agentID, nonce string) string {
	return agentID + ":" + nonce
}

func (c *memNonceCache) SeenRecently(_ context.Context, agentID, nonce string, _ time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	exp, ok := c.entries[c.cacheKey(agentID, nonce)]
	if !ok {
		return false, nil
	}
	return time.Now().Before(exp), nil
}

func (c *memNonceCache) Record(_ context.Context, agentID, nonce string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[c.cacheKey(agentID, nonce)] = time.Now().Add(ttl)
	return nil
}

func (c *memNonceCache) sweep() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		c.mu.Lock()
		for k, exp := range c.entries {
			if now.After(exp) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}
