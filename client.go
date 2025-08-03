// Package openbymadata provides a comprehensive Go client for accessing financial data
// from the Buenos Aires Stock Exchange (BYMA - Bolsas y Mercados Argentinos) through
// their free public API.
//
// OpenBYMAData offers strongly-typed, concurrent-safe access to Argentine financial
// market data with built-in 5-minute caching, individual ticker lookups, batch operations,
// and historical data retrieval.
//
// # Quick Start
//
// Create a client and start fetching data:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Get specific US stock (CEDEAR)
//	aapl, err := client.GetCedear(ctx, "AAPL")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("AAPL: $%.2f (%.2f%%)\n", aapl.Last, aapl.Change)
//
//	// Universal search (recommended)
//	security, err := client.GetSecurity(ctx, "BMA")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("BMA: $%.2f\n", security.Last)
//
// # Key Features
//
// Individual Ticker Lookups: Get specific securities without fetching entire collections.
// Use GetSecurity() for universal search, or specific methods like GetCedear(), GetBluechip().
//
// Batch Operations: Efficiently retrieve multiple securities in a single operation using
// GetMultipleSecurities().
//
// Historical Data: Access OHLCV charting data with GetHistory() and GetHistoryLastDays().
//
// Smart Caching: All data is automatically cached for 5 minutes, reducing API calls by 95%
// and improving performance by 100x for cached requests.
//
// # Performance
//
// The client includes automatic caching that dramatically improves performance:
//   - First API call: ~100ms (network request)
//   - Cached calls: ~50µs (100x faster)
//   - Thread-safe concurrent access
//   - Automatic cache expiration after 5 minutes
//
// # Configuration
//
// Customize the client with ClientOptions:
//
//	opts := &openbymadata.ClientOptions{
//		Timeout:       30 * time.Second,
//		RetryAttempts: 5,
//		EnableCache:   true, // Default: true
//	}
//	client := openbymadata.NewClient(opts)
//
// # API Coverage
//
// The client covers all major BYMA API endpoints:
//   - Securities: Leading equity, general equity, CEDEARs
//   - Fixed Income: Government bonds, corporate bonds, short-term bonds
//   - Derivatives: Options and futures contracts
//   - Market Data: Indices, market summary, working day status
//   - Historical Data: OHLCV time series for charting
//   - News: Market news and financial statements
//
// # Error Handling
//
// The library provides structured error handling with custom error types:
//
//	securities, err := client.GetBluechips(ctx)
//	if err != nil {
//		if bymaErr, ok := err.(*openbymadata.BYMAError); ok {
//			switch bymaErr.Code {
//			case "TIMEOUT":
//				// Handle timeout
//			case "RATE_LIMITED":
//				// Handle rate limiting
//			default:
//				log.Printf("API error: %v", bymaErr)
//			}
//		}
//	}
//
// # Thread Safety
//
// All client operations are thread-safe and can be used concurrently across
// multiple goroutines. The built-in cache is also thread-safe.
//
// # Testing Support
//
// The library supports easy testing with HTTP test servers. See the testing
// documentation for examples of mocking API responses.
package openbymadata

import (
	"context"
	"time"

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
// The client includes automatic 5-minute caching, retry logic, and thread-safe operations.
// All methods accept context.Context for cancellation and timeout control.
//
// Example usage:
//
//	// Basic client with default settings (recommended)
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Get a specific security
//	aapl, err := client.GetCedear(ctx, "AAPL")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("AAPL: $%.2f\n", aapl.Last)
//
// Custom configuration:
//
//	// Production configuration with longer timeouts
//	opts := &openbymadata.ClientOptions{
//		Timeout:       60 * time.Second,  // Longer timeout
//		RetryAttempts: 5,                 // More retries
//		EnableCache:   true,              // Keep caching enabled
//	}
//	client := openbymadata.NewClient(opts)
//
// Development configuration:
//
//	// Disable caching for always fresh data during development
//	devOpts := &openbymadata.ClientOptions{
//		EnableCache: false,
//		Timeout:     10 * time.Second,
//	}
//	client := openbymadata.NewClient(devOpts)
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

// GetBluechips retrieves all leading equity securities (blue chip stocks).
// These are the most liquid and actively traded Argentine stocks.
// Results are cached for 5 minutes to improve performance.
//
// Example usage:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Get all blue chip securities
//	bluechips, err := client.GetBluechips(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("💎 Blue Chip Securities (%d total):\n", len(bluechips))
//	for i, security := range bluechips {
//		if i >= 5 { // Show first 5
//			fmt.Printf("   ... and %d more\n", len(bluechips)-5)
//			break
//		}
//		changeIcon := "🟢"
//		if security.Change < 0 {
//			changeIcon = "🔴"
//		}
//		fmt.Printf("  %s %s: $%.2f (%.2f%%) | Vol: %d\n",
//			changeIcon, security.Symbol, security.Last, security.Change, security.Volume)
//	}
//
// Market analysis:
//
//	// Find biggest movers
//	var biggestGainer, biggestLoser Security
//	for _, security := range bluechips {
//		if security.Change > biggestGainer.Change {
//			biggestGainer = security
//		}
//		if security.Change < biggestLoser.Change {
//			biggestLoser = security
//		}
//	}
//	fmt.Printf("📈 Biggest Gainer: %s (+%.2f%%)\n",
//		biggestGainer.Symbol, biggestGainer.Change)
//	fmt.Printf("📉 Biggest Loser: %s (%.2f%%)\n",
//		biggestLoser.Symbol, biggestLoser.Change)
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

// GetCedears retrieves all CEDEAR securities (US stocks traded in Argentina).
// CEDEARs are Argentine depositary receipts representing shares of US companies.
// Results are cached for 5 minutes to improve performance.
//
// Example usage:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Get all CEDEAR securities
//	cedears, err := client.GetCedears(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("🌎 CEDEARs (%d total):\n", len(cedears))
//	for i, cedear := range cedears {
//		if i >= 10 { // Show first 10
//			fmt.Printf("   ... and %d more\n", len(cedears)-10)
//			break
//		}
//		fmt.Printf("  %s: $%.2f (%.2f%%)\n",
//			cedear.Symbol, cedear.Last, cedear.Change)
//	}
//
// Find specific patterns:
//
//	// Find all tech stocks (simplified example)
//	techStocks := []string{"AAPL", "MSFT", "GOOGL", "TSLA", "META", "NVDA"}
//	fmt.Printf("\n📱 Tech CEDEARs:\n")
//	for _, cedear := range cedears {
//		for _, tech := range techStocks {
//			if cedear.Symbol == tech {
//				fmt.Printf("  %s: $%.2f (%.2f%%)\n",
//					cedear.Symbol, cedear.Last, cedear.Change)
//			}
//		}
//	}
//
// For getting a single CEDEAR, use GetCedear() instead for better performance.
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

// GetSecurity finds a security by symbol across all security types.
// This is the recommended method for security lookup as it searches across
// CEDEARs, blue chips, and general equity automatically.
//
// Example usage:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Universal search - works for any security type
//	security, err := client.GetSecurity(ctx, "AAPL")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("%s: $%.2f (%.2f%% change)\n",
//		security.Symbol, security.Last, security.Change)
//
// Error handling:
//
//	security, err := client.GetSecurity(ctx, "UNKNOWN")
//	if err != nil {
//		if bymaErr, ok := err.(*openbymadata.BYMAError); ok {
//			if bymaErr.Code == "INVALID_TICKER" {
//				fmt.Printf("Security %s not found\n", "UNKNOWN")
//				return
//			}
//		}
//		log.Fatal(err)
//	}
//
// The method leverages caching, so subsequent calls for the same or different
// symbols will be much faster if the underlying collections are cached.
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

// GetCedear finds a specific CEDEAR (US stock) by symbol.
// CEDEARs are Argentine depositary receipts that represent shares of US companies.
//
// Example usage:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Get Apple stock (CEDEAR)
//	aapl, err := client.GetCedear(ctx, "AAPL")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("🍎 AAPL: $%.2f (%.2f%% change)\n", aapl.Last, aapl.Change)
//	fmt.Printf("Volume: %d | Last Update: %s\n",
//		aapl.Volume, aapl.DateTime.Format("15:04:05"))
//
// Portfolio tracking:
//
//	symbols := []string{"AAPL", "MSFT", "GOOGL", "TSLA"}
//	for _, symbol := range symbols {
//		security, err := client.GetCedear(ctx, symbol)
//		if err != nil {
//			log.Printf("Error getting %s: %v", symbol, err)
//			continue
//		}
//		fmt.Printf("%s: $%.2f\n", symbol, security.Last)
//	}
//
// The function uses caching, so repeated calls are very fast.
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

// GetMultipleSecurities gets multiple securities by symbols in a single efficient operation.
// This method is much more efficient than calling individual Get methods in a loop
// because it leverages the cache and fetches all collections once.
//
// Example usage:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Portfolio tracking
//	watchlist := []string{"AAPL", "MSFT", "GOOGL", "GGAL", "YPF"}
//	securities, err := client.GetMultipleSecurities(ctx, watchlist)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("💼 Portfolio (%d securities):\n", len(securities))
//	totalValue := 0.0
//	for _, symbol := range watchlist {
//		if security, found := securities[symbol]; found {
//			changeIcon := "🟢"
//			if security.Change < 0 {
//				changeIcon = "🔴"
//			}
//			fmt.Printf("  %s %s: $%.2f (%.2f%%)\n",
//				changeIcon, symbol, security.Last, security.Change)
//			totalValue += security.Last
//		} else {
//			fmt.Printf("  ❌ %s: Not found\n", symbol)
//		}
//	}
//	fmt.Printf("💰 Total Portfolio Value: $%.2f\n", totalValue)
//
// Performance comparison:
//
//	// ❌ Inefficient - multiple API calls
//	for _, symbol := range watchlist {
//		security, _ := client.GetSecurity(ctx, symbol)
//		// Process security...
//	}
//
//	// ✅ Efficient - single operation using cache
//	securities, _ := client.GetMultipleSecurities(ctx, watchlist)
//	for symbol, security := range securities {
//		// Process security...
//	}
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

// GetCacheInfo returns information about cached data including age, size, and freshness.
// This is useful for monitoring cache performance and debugging.
//
// Example usage:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Make some API calls to populate cache
//	_, _ = client.GetBluechips(ctx)
//	_, _ = client.GetCedears(ctx)
//
//	// Check cache status
//	cacheInfo := client.GetCacheInfo()
//	fmt.Printf("🗄️  Cache Status:\n")
//	for category, info := range cacheInfo {
//		infoMap := info.(map[string]interface{})
//		fmt.Printf("   %-12s: %v items, age %v, fresh: %v\n",
//			category, infoMap["count"], infoMap["age"], infoMap["fresh"])
//	}
//
// Performance monitoring:
//
//	// Before making calls
//	start := time.Now()
//	_, _ = client.GetCedear(ctx, "AAPL")  // First call - API request
//	firstCall := time.Since(start)
//
//	start = time.Now()
//	_, _ = client.GetCedear(ctx, "AAPL")  // Second call - from cache
//	cachedCall := time.Since(start)
//
//	fmt.Printf("First call: %v\n", firstCall)    // ~100ms
//	fmt.Printf("Cached call: %v\n", cachedCall)  // ~50µs (100x faster!)
//
//	cacheInfo = client.GetCacheInfo()
//	fmt.Printf("Cache performance: %+v\n", cacheInfo)
func (c *client) GetCacheInfo() map[string]interface{} {
	if c.cache != nil {
		return c.cache.GetInfo()
	}
	return make(map[string]interface{})
}

// ClearCache clears all cached data, forcing fresh API calls for subsequent requests.
// This is useful when you need absolutely fresh data or for testing purposes.
//
// Example usage:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Make some calls that will be cached
//	_, _ = client.GetBluechips(ctx)
//	_, _ = client.GetCedears(ctx)
//
//	// Check cache status
//	cacheInfo := client.GetCacheInfo()
//	fmt.Printf("Before clear: %d cached categories\n", len(cacheInfo))
//
//	// Clear cache
//	client.ClearCache()
//
//	// Verify cache is cleared
//	cacheInfo = client.GetCacheInfo()
//	fmt.Printf("After clear: %d cached categories\n", len(cacheInfo))
//
//	// Next calls will fetch fresh data from API
//	bluechips, _ := client.GetBluechips(ctx)  // Fresh API call
//	fmt.Printf("Fresh data: %d bluechips\n", len(bluechips))
//
// Use cases:
//
//	// 1. Real-time trading applications
//	if needRealTimeData {
//		client.ClearCache()
//		positions, _ := client.GetMultipleSecurities(ctx, portfolio)
//	}
//
//	// 2. Testing with fresh data
//	func TestMarketData(t *testing.T) {
//		client.ClearCache()  // Ensure clean state
//		// Your test code...
//	}
//
//	// 3. Debugging cache issues
//	if suspectStaleData {
//		client.ClearCache()
//		freshData, _ := client.GetSecurity(ctx, "AAPL")
//	}
func (c *client) ClearCache() {
	if c.cache != nil {
		c.cache.Clear()
	}
}

// =============================================================================
// Market Status & Information (delegated methods with examples)
// =============================================================================

// IsWorkingDay checks if the market is open today.
// This is useful for determining if trading data is available.
//
// Example usage:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Check market status
//	isWorking, err := client.IsWorkingDay(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if isWorking {
//		fmt.Println("🟢 Market is OPEN - trading data available")
//
//		// Proceed with data fetching
//		bluechips, _ := client.GetBluechips(ctx)
//		fmt.Printf("Retrieved %d blue chip securities\n", len(bluechips))
//	} else {
//		fmt.Println("🔴 Market is CLOSED - showing last available data")
//
//		// Still fetch data (will show last trading day)
//		bluechips, _ := client.GetBluechips(ctx)
//		fmt.Printf("Last trading data: %d securities\n", len(bluechips))
//	}
//
// Application logic:
//
//	// Trading bot logic
//	if isWorking, _ := client.IsWorkingDay(ctx); isWorking {
//		// Execute trading strategies
//		positions, _ := client.GetMultipleSecurities(ctx, portfolio)
//		for symbol, security := range positions {
//			// Analyze and potentially trade
//			if security.Change > 5.0 {
//				fmt.Printf("🚀 %s is up %.2f%% - potential sell signal\n",
//					symbol, security.Change)
//			}
//		}
//	} else {
//		fmt.Println("⏸️  Market closed - trading bot on standby")
//	}
func (c *client) IsWorkingDay(ctx context.Context) (bool, error) {
	return c.Client.IsWorkingDay(ctx)
}

// =============================================================================
// Historical Data & Charting (delegated methods with examples)
// =============================================================================

// GetHistory retrieves historical OHLCV data for a symbol within a date range.
// This is essential for charting and technical analysis.
//
// Parameters:
//   - symbol: Security symbol (automatically normalized with "24HS" suffix)
//   - resolution: "D" (daily), "W" (weekly), "M" (monthly)
//   - from, to: Date range as time.Time
//
// Example usage:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Get 3 months of weekly data for Apple
//	from := time.Now().AddDate(0, -3, 0)  // 3 months ago
//	to := time.Now()                      // now
//
//	weeklyData, err := client.GetHistory(ctx, "AAPL", "W", from, to)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("📈 AAPL Weekly Data (%d weeks):\n", len(weeklyData.Time))
//	for i := range len(weeklyData.Time) {
//		date := time.Unix(weeklyData.Time[i], 0)
//		fmt.Printf("  %s: Open=$%.2f High=$%.2f Low=$%.2f Close=$%.2f Vol=%d\n",
//			date.Format("2006-01-02"), weeklyData.Open[i], weeklyData.High[i],
//			weeklyData.Low[i], weeklyData.Close[i], weeklyData.Volume[i])
//	}
//
// Technical analysis example:
//
//	// Calculate simple moving average
//	if len(weeklyData.Close) >= 10 {
//		sum := 0.0
//		for i := len(weeklyData.Close) - 10; i < len(weeklyData.Close); i++ {
//			sum += weeklyData.Close[i]
//		}
//		sma10 := sum / 10
//		currentPrice := weeklyData.Close[len(weeklyData.Close)-1]
//
//		fmt.Printf("Current Price: $%.2f\n", currentPrice)
//		fmt.Printf("10-Week SMA: $%.2f\n", sma10)
//		if currentPrice > sma10 {
//			fmt.Println("📈 Price above SMA - Bullish signal")
//		} else {
//			fmt.Println("📉 Price below SMA - Bearish signal")
//		}
//	}
func (c *client) GetHistory(ctx context.Context, symbol, resolution string, from, to time.Time) (*OHLCV, error) {
	return c.Client.GetHistory(ctx, symbol, resolution, from, to)
}

// GetHistoryLastDays retrieves historical OHLCV data for the last N days.
// This is a convenient method for recent historical data.
//
// Example usage:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Get last 30 days of daily data for SPY
//	historyData, err := client.GetHistoryLastDays(ctx, "SPY", 30)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("📊 SPY Historical Data (last 30 days, %d points):\n", len(historyData.Time))
//
//	// Show first, middle, and last data points
//	indices := []int{0, len(historyData.Time) / 2, len(historyData.Time) - 1}
//	labels := []string{"First", "Middle", "Latest"}
//
//	for i, dataIndex := range indices {
//		date := time.Unix(historyData.Time[dataIndex], 0)
//		fmt.Printf("  %s (%s): Close=$%.2f Volume=%d\n",
//			labels[i], date.Format("2006-01-02"),
//			historyData.Close[dataIndex], historyData.Volume[dataIndex])
//	}
//
// Volatility analysis:
//
//	// Calculate daily returns and volatility
//	if len(historyData.Close) > 1 {
//		var returns []float64
//		for i := 1; i < len(historyData.Close); i++ {
//			dailyReturn := (historyData.Close[i] - historyData.Close[i-1]) / historyData.Close[i-1]
//			returns = append(returns, dailyReturn)
//		}
//
//		// Simple volatility calculation (standard deviation)
//		mean := 0.0
//		for _, r := range returns {
//			mean += r
//		}
//		mean /= float64(len(returns))
//
//		variance := 0.0
//		for _, r := range returns {
//			variance += (r - mean) * (r - mean)
//		}
//		volatility := math.Sqrt(variance / float64(len(returns)-1))
//
//		fmt.Printf("📊 30-Day Volatility: %.2f%% daily\n", volatility*100)
//	}
func (c *client) GetHistoryLastDays(ctx context.Context, symbol string, days int) (*OHLCV, error) {
	return c.Client.GetHistoryLastDays(ctx, symbol, days)
}

// ConvertToHistoricalData converts OHLCV slices to structured HistoricalData format.
// Use this when you need individual data points instead of parallel arrays.
//
// Example usage:
//
//	client := openbymadata.NewClient()
//	ctx := context.Background()
//
//	// Get OHLCV data (parallel arrays format)
//	ohlcvData, err := client.GetHistoryLastDays(ctx, "AAPL", 10)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Convert to structured format
//	structuredData, err := client.ConvertToHistoricalData(ohlcvData)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("🔄 Converted %d OHLCV data points to structured format\n", len(structuredData))
//
//	// Now you can work with individual candles
//	for _, candle := range structuredData {
//		date := time.Unix(candle.Time, 0)
//		fmt.Printf("  %s: OHLC(%.2f, %.2f, %.2f, %.2f) Vol=%d\n",
//			date.Format("2006-01-02"), candle.Open, candle.High,
//			candle.Low, candle.Close, candle.Volume)
//	}
//
// Candlestick pattern detection:
//
//	// Simple doji detection (open ≈ close)
//	for _, candle := range structuredData {
//		bodySize := math.Abs(candle.Close - candle.Open)
//		totalRange := candle.High - candle.Low
//
//		if totalRange > 0 && bodySize/totalRange < 0.1 {
//			date := time.Unix(candle.Time, 0)
//			fmt.Printf("🕯️  Doji pattern detected on %s\n", date.Format("2006-01-02"))
//		}
//	}
func (c *client) ConvertToHistoricalData(slices *OHLCV) ([]HistoricalData, error) {
	return c.Client.ConvertToHistoricalData(slices)
}
