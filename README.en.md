# openbymadata

[![Go Reference](https://pkg.go.dev/badge/github.com/carvalab/openbymadata.svg)](https://pkg.go.dev/github.com/carvalab/openbymadata)
[![CI](https://github.com/carvalab/openbymadata/actions/workflows/go.yml/badge.svg)](https://github.com/carvalab/openbymadata/actions/workflows/go.yml)
[![Live API](https://github.com/carvalab/openbymadata/actions/workflows/integration.yml/badge.svg)](https://github.com/carvalab/openbymadata/actions/workflows/integration.yml)
[![Coverage](https://codecov.io/gh/carvalab/openbymadata/branch/main/graph/badge.svg)](https://app.codecov.io/gh/carvalab/openbymadata)
[![Release](https://badgen.net/github/release/carvalab/openbymadata)](https://github.com/carvalab/openbymadata/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> 🇦🇷 **Español:** [Leer el README en español →](README.md)

Go client for [BYMA Open Data](https://open.bymadata.com.ar), the free market
data API of the Buenos Aires stock exchange. Stocks, CEDEARs, ETFs, bonds,
repos, options, futures, indices, the MEP dollar, live quotes and OHLCV
history, as typed Go structs. No API key, no dependencies.

The **Live API** badge is a test suite that runs against the real API every
weekday, so you know whether BYMA changed something before your users do.

## Install

```bash
go get github.com/carvalab/openbymadata
```

Requires Go 1.26 or newer.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/carvalab/openbymadata"
)

func main() {
	client := openbymadata.NewClient() // 5-minute cache on by default
	defer client.Close()
	ctx := context.Background()

	open, err := client.IsMarketOpen(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("market open:", open)

	ggal, err := client.GetBluechip(ctx, "GGAL")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GGAL %.2f (%+.2f%%)\n", ggal.Last, ggal.Change*100)

	// History needs a settlement suffix: "GGAL 24HS", not "GGAL".
	bars, err := client.GetHistoryLastDays(ctx, "GGAL 24HS", 30)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("daily bars:", len(bars.Time))

	mep, err := client.GetMEPIndices(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("MEP dollar: %.2f\n", mep[0].Last)
}
```

Runnable examples for each method are on
[pkg.go.dev](https://pkg.go.dev/github.com/carvalab/openbymadata#pkg-examples),
and `go run ./cmd/example` walks through most of the API.

## Things to know

- **History symbols need a settlement suffix.** `GetHistory`,
  `GetHistoryLastDays` and `GetIntradayHistory` take `"GGAL 24HS"`. A bare
  `"GGAL"` returns an empty result, not an error.
- **Use `Settlement24HS` or `SettlementCI` for quotes.** Argentina settles
  T+1, so `Settlement48HS` no longer returns prices.
- **Open vs working day.** `IsMarketOpen` is true only while BYMA is trading
  right now. `IsWorkingDay` says whether today is a trading day.
  `GetMarketTime` gives today's opening and closing times (UTC-3).
- **`Change` is a fraction.** `0.0123` means +1.23%.
- **Repos and securities lending prices are rates** as BYMA publishes them,
  and the scale differs between fields. Compare like with like.
- **Intraday data is delayed** on the free API: a few bars per day.
- **`GetCompanyBalance` doesn't work** on the free API anymore; BYMA moved it
  behind a login. Use `GetIncomeStatement` for the filed statement PDFs.

## API at a glance

Market-wide lists are cached for 5 minutes (see [Caching](#caching)).
Per-symbol calls, quotes, history and market status always hit the API.

**Market status and indices**

| Method | Returns |
|---|---|
| `IsMarketOpen(ctx)` | whether BYMA is trading right now |
| `IsWorkingDay(ctx)` | whether today is a trading day |
| `GetMarketTime(ctx)` | today's opening and closing times |
| `GetServerTime(ctx)` | BYMA's clock (UTC-3) |
| `GetIndices(ctx)` | index levels (Merval is `"M"`) |
| `GetIndexHistory(ctx, symbol, res, from, to)` | index OHLCV |
| `GetMEPIndices(ctx)` | MEP dollar indices |
| `MarketResume(ctx)` | traded totals |
| `GetTopMovers(ctx)` | stocks, bonds and options from BYMA's ups/downs panel |

**Equities**

| Method | Returns |
|---|---|
| `GetBluechips(ctx)` | leading panel (panel líder) |
| `GetGalpones(ctx)` | general panel |
| `GetCedears(ctx)` | CEDEARs |
| `GetETFs(ctx)` | ETFs |
| `GetCorporateGovernance(ctx)`, `GetCorporateGovernancePlus(ctx)` | governance panels |

**Fixed income and money market**

| Method | Returns |
|---|---|
| `GetBonds(ctx)` | government bonds |
| `GetShortTermBonds(ctx)` | Letras |
| `GetCorporateBonds(ctx)` | obligaciones negociables |
| `GetGreenBonds`, `GetSocialBonds`, `GetSustainabilityBonds`, `GetLinkedBonds` | labelled bonds |
| `GetSenebiBonds`, `GetSenebiLetters`, `GetSenebiCorporateBonds` | SENEBI segment (`.SB` symbols) |
| `GetFinancialTrusts(ctx)` | fideicomisos financieros |
| `GetRepos(ctx)` | cauciones (repos) |
| `GetSecuritiesLending(ctx)` | securities lending |
| `GetBondIndices(ctx)` | fixed-income indices |
| `GetBondCoupons(ctx, symbol)` | a bond's interest and amortization schedule |

**Derivatives**

| Method | Returns |
|---|---|
| `GetOptions(ctx)` | options, with `OptionType` CALL/PUT |
| `GetFutures(ctx)` | futures (often empty on the free API) |

**Quotes and history**

| Method | Returns |
|---|---|
| `GetCurrentQuote(ctx, symbol, settlement)` | live bid/ask/last |
| `GetHistory(ctx, symbol, res, from, to)` | OHLCV (daily, weekly, monthly or intraday resolutions) |
| `GetHistoryLastDays(ctx, symbol, days)` | daily OHLCV up to now |
| `GetIntradayHistory(ctx, symbol, res, from, to)` | 1, 5 or 60-minute bars |

**Lookups and search**

| Method | Returns |
|---|---|
| `GetSecurity(ctx, symbol)` | a stock or CEDEAR from any equity panel |
| `GetBluechip`, `GetCedear`, `GetGalpone`, `GetBond`, `GetOption`, `GetFuture` | one instrument by symbol |
| `GetMultipleSecurities(ctx, symbols)` | several equities at once |
| `SearchSecurities(ctx, text)` | equities whose symbol contains text |
| `SearchInstruments(ctx, query)` | any instrument by symbol or name, up to 10 |
| `GetInstrumentCatalog(ctx)` | every listed instrument (about 11,000) |

**Companies and filings**

| Method | Returns |
|---|---|
| `GetCompanyInfo(ctx, symbol)` | company profile fields |
| `GetEquityProfile(ctx, symbol)` | security profile (ISIN, par value, ...) |
| `GetCompanyManagement(ctx, symbol)` | directors |
| `GetMarketResponsibles(ctx, symbol)` | market representatives |
| `GetIncomeStatement(ctx, symbol)` | filed statements with PDF links |
| `GetNews(ctx)` | BYMA notices with PDF links |

Field-by-field docs for every type are on
[pkg.go.dev](https://pkg.go.dev/github.com/carvalab/openbymadata).

## Configuration

```go
opts := openbymadata.DefaultClientOptions() // cache on, 30s timeout, 3 retries
opts.Timeout = 10 * time.Second
opts.Logger = myLogger // anything with Debug/Info/Warn/Error
client := openbymadata.NewClient(opts)
```

When you build `ClientOptions` yourself, set `EnableCache: true` if you want
caching; the zero value turns it off.

### Caching

Market-wide lists (panels, bonds, options, indices, news, MEP, the catalog)
are kept in memory for 5 minutes, shared across goroutines. Lookups such as
`GetBluechip` or `GetMultipleSecurities` read from those cached lists.

```go
client.GetCacheInfo() // per-key count, age and freshness
client.ClearCache()   // force fresh data
```

More in [CACHING_GUIDE.md](CACHING_GUIDE.md).

### Errors, retries and cancellation

- Every method takes a `context.Context`. Cancelling it or hitting its
  deadline aborts the request in flight and any pending retry; check with
  `errors.Is(err, context.DeadlineExceeded)`.
- Network errors, HTTP 400, 408, 429 and 5xx are retried (`RetryAttempts`,
  default 3) with a 1s, 2s, 3s… wait. BYMA answers 400 at random under load,
  which is why it's retried. Other 4xx fail at once.
- Errors are wrapped Go errors, for example
  `request failed after 2 attempts: HTTP error 404`.

To see each request, pass a `Logger`; see [DEBUG.md](DEBUG.md).

## Testing

```bash
go test ./...                                                    # unit tests, recorded BYMA responses
INTEGRATION_TEST=true go test -tags integration -run Integration ./...  # live API
```

Unit tests replay responses recorded from the real API
(`internal/api/testdata`), so they run offline. The live suite runs on CI
every weekday; price checks only run while `IsMarketOpen` is true, so
holidays and after-hours runs don't fail.

## Keeping up with BYMA

BYMA changes the API without notice. Three things catch it:

- the weekday **Live API** run,
- a drift check that fails when BYMA stops sending a field we parse,
- a weekly **Discover** run that diffs BYMA's web app against this library
  and opens an issue listing endpoints to add, remove or fix.

[docs/MAINTAINING.md](docs/MAINTAINING.md) has the step-by-step for each
case.

## Contributing

Issues and pull requests are welcome. Run `go test ./...` and `gofmt -s`
before opening a PR. If BYMA changed something, start with
[docs/MAINTAINING.md](docs/MAINTAINING.md).

## Acknowledgments

- [PyOBD](https://github.com/franco-lamas/PyOBD), the Python library this
  started from
- [BYMA](https://www.byma.com.ar/) for the free API

## License

MIT, see [LICENSE](LICENSE). Release notes are on
[GitHub Releases](https://github.com/carvalab/openbymadata/releases).
