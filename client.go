// Package openbymadata is a Go client for BYMA Open Data
// (https://open.bymadata.com.ar), the free market data API of the Buenos
// Aires stock exchange: stocks, CEDEARs, ETFs, bonds, repos, options,
// futures, indices, the MEP dollar, live quotes and OHLCV history. No API key
// is needed.
//
// # Quick Start
//
//	client := openbymadata.NewClient() // 5-minute cache on
//	defer client.Close()
//	ctx := context.Background()
//
//	ggal, err := client.GetBluechip(ctx, "GGAL")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("GGAL %.2f (%+.2f%%)\n", ggal.Last, ggal.Change*100)
//
// # Things to know
//
// History symbols need a settlement suffix: GetHistory(ctx, "GGAL 24HS", ...).
// A bare "GGAL" returns an empty result without an error.
//
// Quotes use Settlement24HS or SettlementCI; 48HS no longer trades since
// Argentina moved to T+1.
//
// IsMarketOpen reports whether BYMA is trading right now (false on holidays
// and outside hours); IsWorkingDay only covers the day.
//
// Change fields are fractions: 0.0123 means +1.23%.
//
// # Caching, retries and cancellation
//
// Market-wide lists are cached for 5 minutes and shared across goroutines;
// quotes, history and market status are always fetched. Every method takes a
// context: cancelling it aborts the request and any pending retry. Network
// errors, HTTP 400, 408, 429 and 5xx are retried RetryAttempts times.
//
//	opts := openbymadata.DefaultClientOptions()
//	opts.Timeout = 10 * time.Second
//	client := openbymadata.NewClient(opts)
package openbymadata

import (
	"context"
	"fmt"
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
// Call Close() when finished to release the underlying HTTP transport.
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
		options.EnableCache = opts[0].EnableCache
	}

	internalOpts := &api.ClientOptions{
		BaseURL:       options.BaseURL,
		Timeout:       options.Timeout,
		RetryAttempts: options.RetryAttempts,
		Logger:        options.Logger,
	}

	c := &client{
		Client: api.New(internalOpts),
	}

	if options.EnableCache {
		c.cache = cache.New()
	}

	return c
}

// cached returns the cached value for key or fetches it from the upstream client,
// storing the result in cache on success.
func cached[T any](ctx context.Context, c *client, key string, fetch func(context.Context) (T, error)) (T, error) {
	if c.cache != nil {
		if v, ok := cache.Get[T](c.cache, key); ok {
			return v, nil
		}
	}

	data, err := fetch(ctx)
	if err != nil {
		var zero T
		return zero, err
	}

	if c.cache != nil {
		cache.Set[T](c.cache, key, data)
	}
	return data, nil
}

// Close releases the HTTP transport owned by the client. It is safe to call
// multiple times.
func (c *client) Close() {
	c.Client.Close()
}

// =============================================================================
// Cache-enabled collection methods
// =============================================================================

// GetBluechips retrieves all leading equity securities (blue chip stocks).
func (c *client) GetBluechips(ctx context.Context) ([]Security, error) {
	return cached(ctx, c, "bluechips", func(ctx context.Context) ([]Security, error) {
		return c.Client.GetBluechips(ctx)
	})
}

// GetCedears retrieves all CEDEAR securities (US stocks traded in Argentina).
func (c *client) GetCedears(ctx context.Context) ([]Security, error) {
	return cached(ctx, c, "cedears", func(ctx context.Context) ([]Security, error) {
		return c.Client.GetCedears(ctx)
	})
}

// GetGalpones retrieves all general equity securities.
func (c *client) GetGalpones(ctx context.Context) ([]Security, error) {
	return cached(ctx, c, "galpones", func(ctx context.Context) ([]Security, error) {
		return c.Client.GetGalpones(ctx)
	})
}

// GetBonds retrieves all government bonds.
func (c *client) GetBonds(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "bonds", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetBonds(ctx)
	})
}

// GetShortTermBonds retrieves all short-term government bonds.
func (c *client) GetShortTermBonds(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "short_bonds", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetShortTermBonds(ctx)
	})
}

// GetCorporateBonds retrieves all corporate bonds.
func (c *client) GetCorporateBonds(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "corporate_bonds", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetCorporateBonds(ctx)
	})
}

// GetOptions retrieves all options contracts.
func (c *client) GetOptions(ctx context.Context) ([]Option, error) {
	return cached(ctx, c, "options", func(ctx context.Context) ([]Option, error) {
		return c.Client.GetOptions(ctx)
	})
}

// GetFutures retrieves all futures contracts.
func (c *client) GetFutures(ctx context.Context) ([]Future, error) {
	return cached(ctx, c, "futures", func(ctx context.Context) ([]Future, error) {
		return c.Client.GetFutures(ctx)
	})
}

// GetETFs retrieves the ETFs listed on BYMA.
func (c *client) GetETFs(ctx context.Context) ([]Security, error) {
	return cached(ctx, c, "etfs", func(ctx context.Context) ([]Security, error) {
		return c.Client.GetETFs(ctx)
	})
}

// GetGreenBonds retrieves bonds labelled green.
func (c *client) GetGreenBonds(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "green_bonds", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetGreenBonds(ctx)
	})
}

// GetSocialBonds retrieves bonds labelled social.
func (c *client) GetSocialBonds(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "social_bonds", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetSocialBonds(ctx)
	})
}

// GetSustainabilityBonds retrieves bonds labelled sustainable.
func (c *client) GetSustainabilityBonds(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "sustainability_bonds", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetSustainabilityBonds(ctx)
	})
}

// GetLinkedBonds retrieves sustainability-linked bonds.
func (c *client) GetLinkedBonds(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "linked_bonds", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetLinkedBonds(ctx)
	})
}

// GetSenebiBonds retrieves government bonds traded in SENEBI (symbols end in ".SB").
func (c *client) GetSenebiBonds(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "senebi_bonds", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetSenebiBonds(ctx)
	})
}

// GetSenebiLetters retrieves Letras traded in SENEBI.
func (c *client) GetSenebiLetters(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "senebi_letters", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetSenebiLetters(ctx)
	})
}

// GetSenebiCorporateBonds retrieves corporate bonds traded in SENEBI.
func (c *client) GetSenebiCorporateBonds(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "senebi_corporate_bonds", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetSenebiCorporateBonds(ctx)
	})
}

// GetBondIndices retrieves BYMA's fixed-income indices.
func (c *client) GetBondIndices(ctx context.Context) ([]BondIndex, error) {
	return cached(ctx, c, "bond_indices", func(ctx context.Context) ([]BondIndex, error) {
		return c.Client.GetBondIndices(ctx)
	})
}

// GetMEPIndices retrieves BYMA's MEP dollar indices.
func (c *client) GetMEPIndices(ctx context.Context) ([]Index, error) {
	return cached(ctx, c, "mep_indices", func(ctx context.Context) ([]Index, error) {
		return c.Client.GetMEPIndices(ctx)
	})
}

// GetTopMovers retrieves BYMA's "main ups and downs" panel.
func (c *client) GetTopMovers(ctx context.Context) (*TopMovers, error) {
	return cached(ctx, c, "top_movers", func(ctx context.Context) (*TopMovers, error) {
		return c.Client.GetTopMovers(ctx)
	})
}

// GetCorporateGovernance retrieves the stocks in BYMA's Corporate Governance panel.
func (c *client) GetCorporateGovernance(ctx context.Context) ([]Security, error) {
	return cached(ctx, c, "corporate_governance", func(ctx context.Context) ([]Security, error) {
		return c.Client.GetCorporateGovernance(ctx)
	})
}

// GetCorporateGovernancePlus retrieves the stocks in BYMA's Corporate Governance Plus panel.
func (c *client) GetCorporateGovernancePlus(ctx context.Context) ([]Security, error) {
	return cached(ctx, c, "corporate_governance_plus", func(ctx context.Context) ([]Security, error) {
		return c.Client.GetCorporateGovernancePlus(ctx)
	})
}

// GetFinancialTrusts retrieves financial trusts (fideicomisos financieros).
func (c *client) GetFinancialTrusts(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "financial_trusts", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetFinancialTrusts(ctx)
	})
}

// GetRepos retrieves cauciones (repo loans). Prices are rates as BYMA publishes them.
func (c *client) GetRepos(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "repos", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetRepos(ctx)
	})
}

// GetSecuritiesLending retrieves securities lending contracts. Prices are rates as BYMA publishes them.
func (c *client) GetSecuritiesLending(ctx context.Context) ([]Bond, error) {
	return cached(ctx, c, "securities_lending", func(ctx context.Context) ([]Bond, error) {
		return c.Client.GetSecuritiesLending(ctx)
	})
}

// GetInstrumentCatalog returns every instrument BYMA lists (about 11,000).
func (c *client) GetInstrumentCatalog(ctx context.Context) ([]CatalogInstrument, error) {
	return cached(ctx, c, "instrument_catalog", func(ctx context.Context) ([]CatalogInstrument, error) {
		return c.Client.GetInstrumentCatalog(ctx)
	})
}

// GetIndices retrieves market indices.
func (c *client) GetIndices(ctx context.Context) ([]Index, error) {
	return cached(ctx, c, "indices", func(ctx context.Context) ([]Index, error) {
		return c.Client.GetIndices(ctx)
	})
}

// MarketResume retrieves market summary.
func (c *client) MarketResume(ctx context.Context) ([]MarketSummary, error) {
	return cached(ctx, c, "market_summary", func(ctx context.Context) ([]MarketSummary, error) {
		return c.Client.MarketResume(ctx)
	})
}

// GetNews retrieves market news.
func (c *client) GetNews(ctx context.Context) ([]News, error) {
	return cached(ctx, c, "news", func(ctx context.Context) ([]News, error) {
		return c.Client.GetNews(ctx)
	})
}

// GetIncomeStatement retrieves income statements for a ticker (cached per ticker).
func (c *client) GetIncomeStatement(ctx context.Context, ticker string) ([]IncomeStatement, error) {
	return cached(ctx, c, cache.IncomeKey(ticker), func(ctx context.Context) ([]api.IncomeStatement, error) {
		return c.Client.GetIncomeStatement(ctx, ticker)
	})
}

// =============================================================================
// Market status and quote lookups
// =============================================================================

// GetMarketTime returns the full market-time payload including opening and
// closing instants.
func (c *client) GetMarketTime(ctx context.Context) (MarketTime, error) {
	mt, err := c.Client.GetMarketTime(ctx)
	if err != nil {
		return MarketTime{}, err
	}
	return MarketTime{
		IsWorkingDay: mt.IsWorkingDay,
		Opens:        mt.Opens,
		Closes:       mt.Closes,
		Timezone:     mt.Timezone,
		OpeningText:  mt.OpeningText,
		ClosingText:  mt.ClosingText,
	}, nil
}

// IsWorkingDay reports whether the BYMA market is open today.
func (c *client) IsWorkingDay(ctx context.Context) (bool, error) {
	return c.Client.IsWorkingDay(ctx)
}

// GetCurrentQuote returns the latest quote for a symbol at the requested
// settlement window. The settlement accepts numeric ("1"/"2"/"3") or
// textual ("CI"/"24HS"/"48HS") codes via Settlement* constants.
func (c *client) GetCurrentQuote(ctx context.Context, symbol string, settlement Settlement) (CurrentQuote, error) {
	return c.Client.GetCurrentQuote(ctx, symbol, settlement)
}

// =============================================================================
// Individual security lookup methods
// =============================================================================

// GetSecurity finds a security by symbol across all security types.
func (c *client) GetSecurity(ctx context.Context, symbol string) (*Security, error) {
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
func (c *client) GetMultipleSecurities(ctx context.Context, symbols []string) (map[string]*Security, error) {
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
func (c *client) GetCacheInfo() map[string]interface{} {
	if c.cache != nil {
		return c.cache.GetInfo()
	}
	return make(map[string]interface{})
}

// ClearCache clears all cached data, forcing fresh API calls for subsequent requests.
func (c *client) ClearCache() {
	if c.cache != nil {
		c.cache.Clear()
	}
}

// =============================================================================
// Historical data
// =============================================================================

// GetHistory retrieves historical OHLCV data for a symbol within a date range
// at the requested resolution.
func (c *client) GetHistory(ctx context.Context, symbol string, resolution Resolution, from, to time.Time) (*OHLCV, error) {
	return c.Client.GetHistory(ctx, symbol, resolution, from, to)
}

// GetHistoryLastDays retrieves historical OHLCV data for the last N days (daily).
func (c *client) GetHistoryLastDays(ctx context.Context, symbol string, days int) (*OHLCV, error) {
	return c.Client.GetHistoryLastDays(ctx, symbol, days)
}

// GetIntradayHistory retrieves intraday OHLCV bars. Default resolution is 1
// minute; pass Resolution5Min or Resolution60Min for coarser bars.
func (c *client) GetIntradayHistory(ctx context.Context, symbol string, resolution Resolution, from, to time.Time) (*OHLCV, error) {
	if resolution == "" {
		resolution = Resolution1Min
	}
	return c.Client.GetIntradayHistory(ctx, symbol, resolution, from, to)
}

// =============================================================================
// Company metadata
// =============================================================================

// GetCompanyInfo returns the public profile (society/company data) for a symbol.
func (c *client) GetCompanyInfo(ctx context.Context, symbol string) (*CompanyInfo, error) {
	return c.Client.GetCompanyInfo(ctx, symbol)
}

// GetEquityProfile returns the technical profile for a symbol.
func (c *client) GetEquityProfile(ctx context.Context, symbol string) (*EquityProfile, error) {
	return c.Client.GetEquityProfile(ctx, symbol)
}

// GetCompanyManagement returns the directors/management roster for a symbol.
func (c *client) GetCompanyManagement(ctx context.Context, symbol string) ([]CompanyDirector, error) {
	return c.Client.GetCompanyManagement(ctx, symbol)
}

// GetCompanyBalance returns the latest balance sheet rows for a symbol.
// BYMA now puts this endpoint behind a login, so it fails on the free API;
// use GetIncomeStatement for the filed statement PDFs instead.
func (c *client) GetCompanyBalance(ctx context.Context, symbol string) ([]CompanyBalance, error) {
	return c.Client.GetCompanyBalance(ctx, symbol)
}

// =============================================================================
// Deprecated aliases for migration from pyhomebroker-style naming.
// =============================================================================

// GetLeadingEquity is a deprecated alias for GetBluechips.
func (c *client) GetLeadingEquity(ctx context.Context) ([]Security, error) {
	return c.GetBluechips(ctx)
}

// GetGeneralEquity is a deprecated alias for GetGalpones.
func (c *client) GetGeneralEquity(ctx context.Context) ([]Security, error) {
	return c.GetGalpones(ctx)
}

// GetPublicBonds is a deprecated alias for GetBonds.
func (c *client) GetPublicBonds(ctx context.Context) ([]Bond, error) {
	return c.GetBonds(ctx)
}

// GetLetters is a deprecated alias for GetShortTermBonds.
func (c *client) GetLetters(ctx context.Context) ([]Bond, error) {
	return c.GetShortTermBonds(ctx)
}

// Ensure fmt remains in scope for any future added use without re-importing it.
var _ = fmt.Sprintf
