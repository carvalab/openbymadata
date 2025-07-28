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

	// Collections cache
	bluechips      *cachedSecurities
	cedears        *cachedSecurities
	galpones       *cachedSecurities
	bonds          *cachedBonds
	shortBonds     *cachedBonds
	corporateBonds *cachedBonds
	options        *cachedOptions
	futures        *cachedFutures
	indices        *cachedIndices
	marketSummary  *cachedMarketSummary
	news           *cachedNews

	// Income statements cache (per symbol)
	incomeStatements map[string]*cachedIncomeStatements
}

// Cached data structures
type cachedSecurities struct {
	data      []api.Security
	timestamp time.Time
}

type cachedBonds struct {
	data      []api.Bond
	timestamp time.Time
}

type cachedOptions struct {
	data      []api.Option
	timestamp time.Time
}

type cachedFutures struct {
	data      []api.Future
	timestamp time.Time
}

type cachedIndices struct {
	data      []api.Index
	timestamp time.Time
}

type cachedMarketSummary struct {
	data      []api.MarketSummary
	timestamp time.Time
}

type cachedNews struct {
	data      []api.News
	timestamp time.Time
}

type cachedIncomeStatements struct {
	data      []api.IncomeStatement
	timestamp time.Time
}

// New creates a new cache with 5-minute duration
func New() *Cache {
	return &Cache{
		duration:         5 * time.Minute,
		incomeStatements: make(map[string]*cachedIncomeStatements),
	}
}

// isFresh checks if cached data is still valid
func (c *Cache) isFresh(timestamp time.Time) bool {
	return time.Since(timestamp) < c.duration
}

// GetBluechips returns cached data or nil if not available/expired
func (c *Cache) GetBluechips() ([]api.Security, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.bluechips != nil && c.isFresh(c.bluechips.timestamp) {
		return c.bluechips.data, true
	}
	return nil, false
}

// SetBluechips stores data in cache
func (c *Cache) SetBluechips(data []api.Security) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bluechips = &cachedSecurities{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetCedears returns cached data or nil if not available/expired
func (c *Cache) GetCedears() ([]api.Security, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cedears != nil && c.isFresh(c.cedears.timestamp) {
		return c.cedears.data, true
	}
	return nil, false
}

// SetCedears stores data in cache
func (c *Cache) SetCedears(data []api.Security) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cedears = &cachedSecurities{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetGalpones returns cached data or nil if not available/expired
func (c *Cache) GetGalpones() ([]api.Security, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.galpones != nil && c.isFresh(c.galpones.timestamp) {
		return c.galpones.data, true
	}
	return nil, false
}

// SetGalpones stores data in cache
func (c *Cache) SetGalpones(data []api.Security) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.galpones = &cachedSecurities{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetBonds returns cached data or nil if not available/expired
func (c *Cache) GetBonds() ([]api.Bond, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.bonds != nil && c.isFresh(c.bonds.timestamp) {
		return c.bonds.data, true
	}
	return nil, false
}

// SetBonds stores data in cache
func (c *Cache) SetBonds(data []api.Bond) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bonds = &cachedBonds{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetShortTermBonds returns cached data or nil if not available/expired
func (c *Cache) GetShortTermBonds() ([]api.Bond, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.shortBonds != nil && c.isFresh(c.shortBonds.timestamp) {
		return c.shortBonds.data, true
	}
	return nil, false
}

// SetShortTermBonds stores data in cache
func (c *Cache) SetShortTermBonds(data []api.Bond) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.shortBonds = &cachedBonds{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetCorporateBonds returns cached data or nil if not available/expired
func (c *Cache) GetCorporateBonds() ([]api.Bond, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.corporateBonds != nil && c.isFresh(c.corporateBonds.timestamp) {
		return c.corporateBonds.data, true
	}
	return nil, false
}

// SetCorporateBonds stores data in cache
func (c *Cache) SetCorporateBonds(data []api.Bond) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.corporateBonds = &cachedBonds{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetOptions returns cached data or nil if not available/expired
func (c *Cache) GetOptions() ([]api.Option, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.options != nil && c.isFresh(c.options.timestamp) {
		return c.options.data, true
	}
	return nil, false
}

// SetOptions stores data in cache
func (c *Cache) SetOptions(data []api.Option) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.options = &cachedOptions{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetFutures returns cached data or nil if not available/expired
func (c *Cache) GetFutures() ([]api.Future, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.futures != nil && c.isFresh(c.futures.timestamp) {
		return c.futures.data, true
	}
	return nil, false
}

// SetFutures stores data in cache
func (c *Cache) SetFutures(data []api.Future) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.futures = &cachedFutures{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetIndices returns cached data or nil if not available/expired
func (c *Cache) GetIndices() ([]api.Index, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.indices != nil && c.isFresh(c.indices.timestamp) {
		return c.indices.data, true
	}
	return nil, false
}

// SetIndices stores data in cache
func (c *Cache) SetIndices(data []api.Index) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.indices = &cachedIndices{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetMarketSummary returns cached data or nil if not available/expired
func (c *Cache) GetMarketSummary() ([]api.MarketSummary, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.marketSummary != nil && c.isFresh(c.marketSummary.timestamp) {
		return c.marketSummary.data, true
	}
	return nil, false
}

// SetMarketSummary stores data in cache
func (c *Cache) SetMarketSummary(data []api.MarketSummary) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.marketSummary = &cachedMarketSummary{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetNews returns cached data or nil if not available/expired
func (c *Cache) GetNews() ([]api.News, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.news != nil && c.isFresh(c.news.timestamp) {
		return c.news.data, true
	}
	return nil, false
}

// SetNews stores data in cache
func (c *Cache) SetNews(data []api.News) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.news = &cachedNews{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetIncomeStatement returns cached data or nil if not available/expired
func (c *Cache) GetIncomeStatement(ticker string) ([]api.IncomeStatement, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if cached, exists := c.incomeStatements[ticker]; exists && c.isFresh(cached.timestamp) {
		return cached.data, true
	}
	return nil, false
}

// SetIncomeStatement stores data in cache
func (c *Cache) SetIncomeStatement(ticker string, data []api.IncomeStatement) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.incomeStatements[ticker] = &cachedIncomeStatements{
		data:      data,
		timestamp: time.Now(),
	}
}

// GetInfo returns information about cached data
func (c *Cache) GetInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info := make(map[string]interface{})

	if c.bluechips != nil {
		info["bluechips"] = map[string]interface{}{
			"count":     len(c.bluechips.data),
			"timestamp": c.bluechips.timestamp,
			"age":       time.Since(c.bluechips.timestamp),
			"fresh":     c.isFresh(c.bluechips.timestamp),
		}
	}

	if c.cedears != nil {
		info["cedears"] = map[string]interface{}{
			"count":     len(c.cedears.data),
			"timestamp": c.cedears.timestamp,
			"age":       time.Since(c.cedears.timestamp),
			"fresh":     c.isFresh(c.cedears.timestamp),
		}
	}

	if c.galpones != nil {
		info["galpones"] = map[string]interface{}{
			"count":     len(c.galpones.data),
			"timestamp": c.galpones.timestamp,
			"age":       time.Since(c.galpones.timestamp),
			"fresh":     c.isFresh(c.galpones.timestamp),
		}
	}

	return info
}

// Clear clears all cached data
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bluechips = nil
	c.cedears = nil
	c.galpones = nil
	c.bonds = nil
	c.shortBonds = nil
	c.corporateBonds = nil
	c.options = nil
	c.futures = nil
	c.indices = nil
	c.marketSummary = nil
	c.news = nil
	c.incomeStatements = make(map[string]*cachedIncomeStatements)
}
