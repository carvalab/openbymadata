# Caching

The client keeps market-wide lists in memory for **5 minutes**. Repeated
calls inside that window, from any goroutine, are answered without a request.

## What is cached

| Cached for 5 minutes | Always fetched |
|---|---|
| `GetBluechips`, `GetGalpones`, `GetCedears`, `GetETFs`, `GetCorporateGovernance`, `GetCorporateGovernancePlus` | `GetCurrentQuote` |
| `GetBonds`, `GetShortTermBonds`, `GetCorporateBonds`, the labelled and SENEBI bond lists, `GetFinancialTrusts`, `GetRepos`, `GetSecuritiesLending`, `GetBondIndices` | `GetHistory`, `GetHistoryLastDays`, `GetIntradayHistory`, `GetIndexHistory` |
| `GetOptions`, `GetFutures` | `IsMarketOpen`, `IsWorkingDay`, `GetMarketTime`, `GetServerTime` |
| `GetIndices`, `GetMEPIndices`, `MarketResume`, `GetTopMovers` | `GetCompanyInfo`, `GetEquityProfile`, `GetCompanyManagement`, `GetCompanyBalance`, `GetMarketResponsibles` |
| `GetNews`, `GetIncomeStatement` (per ticker), `GetInstrumentCatalog` | `GetBondCoupons`, `SearchInstruments` |

Lookups read from the cached lists, so they cost at most one request per
list every 5 minutes:

- `GetBluechip`, `GetCedear`, `GetGalpone`: their panel
- `GetSecurity`, `GetMultipleSecurities`, `SearchSecurities`: the three
  equity panels
- `GetBond`: government bonds, Letras and corporate bonds
- `GetOption`, `GetFuture`: options or futures

A failed request is not cached; the next call tries again.

## Turning it on and off

```go
client := openbymadata.NewClient()        // cache on
opts := openbymadata.DefaultClientOptions() // cache on
opts.EnableCache = false                    // off: every call hits the API
```

If you build `ClientOptions` yourself, the zero value of `EnableCache` is
`false`, so set it to `true` to keep caching.

## Inspecting and clearing

```go
for key, v := range client.GetCacheInfo() {
	info := v.(map[string]any)
	fmt.Println(key, info["count"], info["age"], info["fresh"])
}

client.ClearCache() // next calls fetch fresh data
```

## Freshness

Five minutes is fine for screens and reports. For trading decisions, use
`GetCurrentQuote` (never cached) or clear the cache first. Each `Security`
and `Bond` also has `DateTime`: today's time of the last trade, or the fetch
time when BYMA sends none.
