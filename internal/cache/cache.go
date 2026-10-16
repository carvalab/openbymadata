package cache

import (
	"sync"
	"time"

	"github.com/carvalab/openbymadata/internal/api"
)

// Cache provides 5-minute caching for BYMA data
type Cache struct {
	mu       sync.RWMutex
	duration time.Duration
	entries  map[string]*entry
}

type entry struct {
	data      any
	timestamp time.Time
}

// Income statements are keyed per ticker using the "income:" prefix.
const incomePrefix = "income:"

// New creates a new cache with 5-minute duration
func New() *Cache {
	return &Cache{
		duration: 5 * time.Minute,
		entries:  make(map[string]*entry),
	}
}

func (c *Cache) isFresh(timestamp time.Time) bool {
	return time.Since(timestamp) < c.duration
}

// Get returns the cached value for key, or the zero value and false if absent,
// expired, or stored under a different type.
func Get[T any](c *Cache, key string) (T, bool) {
	var zero T
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[key]
	if !ok || !c.isFresh(e.timestamp) {
		return zero, false
	}
	data, ok := e.data.(T)
	if !ok {
		return zero, false
	}
	return data, true
}

// Set stores data under key with the current timestamp.
func Set[T any](c *Cache, key string, data T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &entry{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetInfo returns information about cached data, keyed by cache key.
func (c *Cache) GetInfo() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info := make(map[string]any, len(c.entries))
	for key, e := range c.entries {
		count := 0
		switch v := e.data.(type) {
		case []api.Security:
			count = len(v)
		case []api.Bond:
			count = len(v)
		case []api.Option:
			count = len(v)
		case []api.Future:
			count = len(v)
		case []api.Index:
			count = len(v)
		case []api.MarketSummary:
			count = len(v)
		case []api.News:
			count = len(v)
		case []api.IncomeStatement:
			count = len(v)
		case []api.BondIndex:
			count = len(v)
		case []api.CatalogInstrument:
			count = len(v)
		}
		info[key] = map[string]any{
			"count":     count,
			"timestamp": e.timestamp,
			"age":       time.Since(e.timestamp),
			"fresh":     c.isFresh(e.timestamp),
		}
	}
	return info
}

// Clear clears all cached data
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*entry)
}

// IncomeKey returns the cache key for a ticker's income statements.
func IncomeKey(ticker string) string {
	return incomePrefix + ticker
}
