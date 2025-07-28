# BYMA Data Caching Guide

## Overview
The BYMA client provides intelligent 5-minute caching to reduce API calls and improve performance when retrieving market data. Caching is **enabled by default** and works transparently.

## Quick Start

### 1. Create a Client (Caching Enabled by Default)

```go
import "github.com/carvalab/openbymadata"

// Caching is enabled by default
client := openbymadata.NewClient()

// Or with custom options (cache still enabled)
client := openbymadata.NewClient(&openbymadata.ClientOptions{
    Timeout:       15 * time.Second,
    RetryAttempts: 2,
    EnableCache:   true, // This is the default
})

// To disable caching (not recommended)
client := openbymadata.NewClient(&openbymadata.ClientOptions{
    EnableCache: false,
})
```

### 2. Get Single Ticker Value

```go
ctx := context.Background()

// Get a specific CEDEAR
aapl, err := client.GetCedear(ctx, "AAPL")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("AAPL: $%.2f (Change: %.2f%%)\n", aapl.Last, aapl.Change)
fmt.Printf("Last Updated: %s\n", aapl.DateTime.Format("15:04:05"))
```

### 3. Universal Symbol Lookup

```go
// Automatically searches across all security types
security, err := client.GetSecurity(ctx, "YPF")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("%s: $%.2f\n", security.Symbol, security.Last)
```

### 4. Get Multiple Tickers Efficiently

```go
symbols := []string{"AAPL", "GOOGL", "MSFT", "TSLA"}
securities, err := client.GetMultipleSecurities(ctx, symbols)
if err != nil {
    log.Fatal(err)
}

for symbol, security := range securities {
    fmt.Printf("%s: $%.2f (%.2f%%)\n", 
        symbol, security.Last, security.Change)
}
```

## Available Methods

### Security Lookup Methods
- `GetSecurity(ctx, symbol)` - Universal lookup across all types
- `GetBluechip(ctx, symbol)` - Blue chip stocks only
- `GetCedear(ctx, symbol)` - CEDEARs only  
- `GetGalpone(ctx, symbol)` - General equity only
- `GetBond(ctx, symbol)` - All bond types
- `GetOption(ctx, symbol)` - Options
- `GetFuture(ctx, symbol)` - Futures

### Batch Operations
- `GetMultipleSecurities(ctx, []symbols)` - Multiple securities at once
- `SearchSecurities(ctx, searchText)` - Search by partial symbol

### Collection Methods (All Cached)
- `GetBluechips(ctx)` - All blue chip stocks
- `GetCedears(ctx)` - All CEDEARs
- `GetGalpones(ctx)` - All general equity
- `GetBonds(ctx)` - All bonds
- `GetShortTermBonds(ctx)` - Short-term bonds
- `GetCorporateBonds(ctx)` - Corporate bonds
- `GetOptions(ctx)` - All options
- `GetFutures(ctx)` - All futures
- `GetIndices(ctx)` - Market indices
- `MarketResume(ctx)` - Market summary
- `GetNews(ctx)` - Market news

### Cache Management
- `GetCacheInfo()` - View cache status
- `ClearCache()` - Clear all cached data

## Caching Behavior

### 5-Minute Cache Duration
- Data is cached for exactly **5 minutes**
- Fresh data is returned immediately from cache
- Expired data triggers new API call

### Cache Sharing
- Individual symbol lookups use the same cached collections
- Multiple requests for different symbols share the same data
- No duplicate API calls within the cache period

### Thread Safety
- All operations are thread-safe
- Multiple goroutines can safely use the same client instance

## Example: Real-world Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/carvalab/openbymadata"
)

func main() {
    client := openbymadata.NewClient() // Cache enabled by default
    ctx := context.Background()
    
    // Monitor multiple tickers efficiently
    watchlist := []string{"YPF", "GGAL", "AAPL", "GOOGL", "BMA"}
    
    for {
        startTime := time.Now()
        securities, err := client.GetMultipleSecurities(ctx, watchlist)
        if err != nil {
            log.Printf("Error: %v", err)
            time.Sleep(30 * time.Second)
            continue
        }
        
        fmt.Printf("\n📊 Portfolio Update (%v)\n", time.Since(startTime))
        fmt.Println(strings.Repeat("=", 40))
        
        for _, symbol := range watchlist {
            if security, found := securities[symbol]; found {
                changeIcon := "📈"
                if security.Change < 0 {
                    changeIcon = "📉"
                }
                fmt.Printf("%s %s: $%-8.2f %+6.2f%%\n",
                    changeIcon, symbol, security.Last, security.Change)
            } else {
                fmt.Printf("❌ %s: Not found\n", symbol)
            }
        }
        
        // Check cache efficiency
        cacheInfo := client.GetCacheInfo()
        if info, exists := cacheInfo["cedears"]; exists {
            infoMap := info.(map[string]interface{})
            fmt.Printf("\n💾 Cache: %v securities, age %v\n", 
                infoMap["count"], infoMap["age"])
        }
        
        // Wait 1 minute before next update (cache will be used)
        time.Sleep(1 * time.Minute)
    }
}
```

## Performance Benefits

### Speed Improvements
- **First call**: ~500ms-2s (API call)
- **Cached calls**: ~1-5ms (memory access)
- **Speed gain**: 100-2000x faster

### Bandwidth Savings
- Reduces API calls by up to 95%
- Lower bandwidth usage
- Faster response times

### Rate Limit Protection
- Prevents hitting API rate limits
- Allows frequent data queries
- Maintains data freshness

## Best Practices

1. **Use cached client for frequent queries**
2. **Batch multiple symbol lookups** with `GetMultipleSecurities()`
3. **Monitor cache status** with `GetCacheInfo()`
4. **Consider data freshness** for time-sensitive applications
5. **Clear cache** if you need absolutely fresh data

## Data Freshness

### Market Data Timestamp
Each security includes a `DateTime` field showing when the data was last traded:

```go
security, _ := client.GetCedear(ctx, "AAPL")
dataAge := time.Since(security.DateTime)
fmt.Printf("Data is %v old\n", dataAge.Round(time.Minute))
```

### Cache Expiration
Cache automatically expires after 5 minutes. You can check cache age:

```go
cacheInfo := client.GetCacheInfo()
if info, exists := cacheInfo["cedears"]; exists {
    infoMap := info.(map[string]interface{})
    fmt.Printf("Cache age: %v\n", infoMap["age"])
    fmt.Printf("Cache fresh: %v\n", infoMap["fresh"])
}
```

## Migration from Previous Versions

No migration needed! Caching is now integrated and enabled by default:

```go
// This automatically includes 5-minute caching
client := openbymadata.NewClient()

// All existing code works exactly the same
bluechips, err := client.GetBluechips(ctx)
aapl, err := client.GetCedear(ctx, "AAPL")

// New methods are also available
security, err := client.GetSecurity(ctx, "YPF")
securities, err := client.GetMultipleSecurities(ctx, []string{"AAPL", "GOOGL"})
``` 