package openbymadata

import (
	"context"

	"github.com/carvalab/openbymadata/internal/api"
	"github.com/carvalab/openbymadata/internal/cache"
	"github.com/carvalab/openbymadata/internal/helpers"
)

// client wraps the internal client and implements the public interface
type client struct {
	*api.Client
	cache *cache.Cache
}

// NewClient creates a new BYMA data client with the provided options.
// This is the main entry point for creating a client instance.
//
// Example usage:
//
//	client := openbymadata.NewClient()
//
// Or with custom options:
//
//	opts := &openbymadata.ClientOptions{
//		Timeout:       30 * time.Second,
//		RetryAttempts: 5,
//	}
//	client := openbymadata.NewClient(opts)
func NewClient(opts ...*ClientOptions) Client {
	options := DefaultClientOptions()
	if len(opts) > 0 && opts[0] != nil {
		if opts[0].BaseURL != "" {
			options.BaseURL = opts[0].BaseURL
		}
		if opts[0].Timeout > 0 {
			options.Timeout = opts[0].Timeout
		}
		if opts[0].RetryAttempts > 0 {
			options.RetryAttempts = opts[0].RetryAttempts
		}
		if opts[0].Logger != nil {
			options.Logger = opts[0].Logger
		}
		// EnableCache is handled below
	}

	// Convert to internal options
	internalOpts := &api.ClientOptions{
		BaseURL:       options.BaseURL,
		Timeout:       options.Timeout,
		RetryAttempts: options.RetryAttempts,
		Logger:        &loggerAdapter{logger: options.Logger},
	}

	c := &client{
		Client: api.New(internalOpts),
	}

	// Initialize cache if enabled
	if options.EnableCache {
		c.cache = cache.New()
	}

	return c
}

// loggerAdapter adapts the public logger interface to the internal one
type loggerAdapter struct {
	logger Logger
}

func (l *loggerAdapter) Debug(msg string, fields ...api.LogField) {
	publicFields := make([]LogField, len(fields))
	for i, f := range fields {
		publicFields[i] = LogField{Key: f.Key, Value: f.Value}
	}
	l.logger.Debug(msg, publicFields...)
}

func (l *loggerAdapter) Info(msg string, fields ...api.LogField) {
	publicFields := make([]LogField, len(fields))
	for i, f := range fields {
		publicFields[i] = LogField{Key: f.Key, Value: f.Value}
	}
	l.logger.Info(msg, publicFields...)
}

func (l *loggerAdapter) Warn(msg string, fields ...api.LogField) {
	publicFields := make([]LogField, len(fields))
	for i, f := range fields {
		publicFields[i] = LogField{Key: f.Key, Value: f.Value}
	}
	l.logger.Warn(msg, publicFields...)
}

func (l *loggerAdapter) Error(msg string, fields ...api.LogField) {
	publicFields := make([]LogField, len(fields))
	for i, f := range fields {
		publicFields[i] = LogField{Key: f.Key, Value: f.Value}
	}
	l.logger.Error(msg, publicFields...)
}

// =============================================================================
// Cache-enabled methods override the base Client methods
// =============================================================================

// GetBluechips with caching support
func (c *client) GetBluechips(ctx context.Context) ([]Security, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetBluechips(); found {
			return cached, nil
		}
	}

	data, err := c.Client.GetBluechips(ctx)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetBluechips(data)
	}

	return data, nil
}

// GetCedears with caching support
func (c *client) GetCedears(ctx context.Context) ([]Security, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetCedears(); found {
			return cached, nil
		}
	}

	data, err := c.Client.GetCedears(ctx)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetCedears(data)
	}

	return data, nil
}

// GetGalpones with caching support
func (c *client) GetGalpones(ctx context.Context) ([]Security, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetGalpones(); found {
			return cached, nil
		}
	}

	data, err := c.Client.GetGalpones(ctx)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetGalpones(data)
	}

	return data, nil
}

// GetBonds with caching support
func (c *client) GetBonds(ctx context.Context) ([]Bond, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetBonds(); found {
			return cached, nil
		}
	}

	data, err := c.Client.GetBonds(ctx)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetBonds(data)
	}

	return data, nil
}

// GetShortTermBonds with caching support
func (c *client) GetShortTermBonds(ctx context.Context) ([]Bond, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetShortTermBonds(); found {
			return cached, nil
		}
	}

	data, err := c.Client.GetShortTermBonds(ctx)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetShortTermBonds(data)
	}

	return data, nil
}

// GetCorporateBonds with caching support
func (c *client) GetCorporateBonds(ctx context.Context) ([]Bond, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetCorporateBonds(); found {
			return cached, nil
		}
	}

	data, err := c.Client.GetCorporateBonds(ctx)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetCorporateBonds(data)
	}

	return data, nil
}

// GetOptions with caching support
func (c *client) GetOptions(ctx context.Context) ([]Option, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetOptions(); found {
			return cached, nil
		}
	}

	data, err := c.Client.GetOptions(ctx)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetOptions(data)
	}

	return data, nil
}

// GetFutures with caching support
func (c *client) GetFutures(ctx context.Context) ([]Future, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetFutures(); found {
			return cached, nil
		}
	}

	data, err := c.Client.GetFutures(ctx)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetFutures(data)
	}

	return data, nil
}

// GetIndices with caching support
func (c *client) GetIndices(ctx context.Context) ([]Index, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetIndices(); found {
			return cached, nil
		}
	}

	data, err := c.Client.GetIndices(ctx)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetIndices(data)
	}

	return data, nil
}

// MarketResume with caching support
func (c *client) MarketResume(ctx context.Context) ([]MarketSummary, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetMarketSummary(); found {
			return cached, nil
		}
	}

	data, err := c.Client.MarketResume(ctx)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetMarketSummary(data)
	}

	return data, nil
}

// GetNews with caching support
func (c *client) GetNews(ctx context.Context) ([]News, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetNews(); found {
			return cached, nil
		}
	}

	data, err := c.Client.GetNews(ctx)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetNews(data)
	}

	return data, nil
}

// GetIncomeStatement with caching support (per ticker)
func (c *client) GetIncomeStatement(ctx context.Context, ticker string) ([]IncomeStatement, error) {
	if c.cache != nil {
		if cached, found := c.cache.GetIncomeStatement(ticker); found {
			return cached, nil
		}
	}

	data, err := c.Client.GetIncomeStatement(ctx, ticker)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		c.cache.SetIncomeStatement(ticker, data)
	}

	return data, nil
}

// =============================================================================
// Individual security lookup methods
// =============================================================================

// GetSecurity finds a security by symbol across all security types
func (c *client) GetSecurity(ctx context.Context, symbol string) (*Security, error) {
	// Get all security collections (use cache when available)
	bluechips, err := c.GetBluechips(ctx)
	if err != nil {
		return nil, err
	}

	cedears, err := c.GetCedears(ctx)
	if err != nil {
		return nil, err
	}

	galpones, err := c.GetGalpones(ctx)
	if err != nil {
		return nil, err
	}

	return helpers.FindSecurityBySymbol(symbol, bluechips, cedears, galpones)
}

// GetBluechip finds a specific blue chip security by symbol
func (c *client) GetBluechip(ctx context.Context, symbol string) (*Security, error) {
	bluechips, err := c.GetBluechips(ctx)
	if err != nil {
		return nil, err
	}

	return helpers.FindSecurityInCollection(symbol, bluechips)
}

// GetCedear finds a specific CEDEAR by symbol
func (c *client) GetCedear(ctx context.Context, symbol string) (*Security, error) {
	cedears, err := c.GetCedears(ctx)
	if err != nil {
		return nil, err
	}

	return helpers.FindSecurityInCollection(symbol, cedears)
}

// GetGalpone finds a specific general equity security by symbol
func (c *client) GetGalpone(ctx context.Context, symbol string) (*Security, error) {
	galpones, err := c.GetGalpones(ctx)
	if err != nil {
		return nil, err
	}

	return helpers.FindSecurityInCollection(symbol, galpones)
}

// GetBond finds a specific bond by symbol across all bond types
func (c *client) GetBond(ctx context.Context, symbol string) (*Bond, error) {
	bonds, err := c.GetBonds(ctx)
	if err != nil {
		return nil, err
	}

	shortBonds, err := c.GetShortTermBonds(ctx)
	if err != nil {
		return nil, err
	}

	corporateBonds, err := c.GetCorporateBonds(ctx)
	if err != nil {
		return nil, err
	}

	return helpers.FindBondBySymbol(symbol, bonds, shortBonds, corporateBonds)
}

// GetOption finds a specific option by symbol
func (c *client) GetOption(ctx context.Context, symbol string) (*Option, error) {
	options, err := c.GetOptions(ctx)
	if err != nil {
		return nil, err
	}

	return helpers.FindOptionBySymbol(symbol, options)
}

// GetFuture finds a specific future by symbol
func (c *client) GetFuture(ctx context.Context, symbol string) (*Future, error) {
	futures, err := c.GetFutures(ctx)
	if err != nil {
		return nil, err
	}

	return helpers.FindFutureBySymbol(symbol, futures)
}

// =============================================================================
// Batch operations
// =============================================================================

// GetMultipleSecurities gets multiple securities by symbols in a single operation
func (c *client) GetMultipleSecurities(ctx context.Context, symbols []string) (map[string]*Security, error) {
	// Pre-load all security collections to use the cache efficiently
	bluechips, err := c.GetBluechips(ctx)
	if err != nil {
		return nil, err
	}

	cedears, err := c.GetCedears(ctx)
	if err != nil {
		return nil, err
	}

	galpones, err := c.GetGalpones(ctx)
	if err != nil {
		return nil, err
	}

	return helpers.GetMultipleSecurities(symbols, bluechips, cedears, galpones), nil
}

// SearchSecurities searches for securities containing the given text in their symbol
func (c *client) SearchSecurities(ctx context.Context, searchText string) ([]Security, error) {
	bluechips, err := c.GetBluechips(ctx)
	if err != nil {
		return nil, err
	}

	cedears, err := c.GetCedears(ctx)
	if err != nil {
		return nil, err
	}

	galpones, err := c.GetGalpones(ctx)
	if err != nil {
		return nil, err
	}

	return helpers.SearchSecurities(searchText, bluechips, cedears, galpones), nil
}

// =============================================================================
// Cache management
// =============================================================================

// GetCacheInfo returns information about cached data
func (c *client) GetCacheInfo() map[string]interface{} {
	if c.cache != nil {
		return c.cache.GetInfo()
	}
	return make(map[string]interface{})
}

// ClearCache clears all cached data
func (c *client) ClearCache() {
	if c.cache != nil {
		c.cache.Clear()
	}
}
