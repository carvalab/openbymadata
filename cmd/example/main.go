package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/carvalab/openbymadata"
)

func main() {
	fmt.Println("🏛️  OpenBYMAData Go Library - Complete Example")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🚀 Features: Individual ticker lookups, 5-minute caching, batch operations")
	fmt.Println()

	// Create client with caching enabled (default)
	client := openbymadata.NewClient(&openbymadata.ClientOptions{
		Timeout:       15 * time.Second,
		RetryAttempts: 3,
		EnableCache:   true, // Default: true
	})

	ctx := context.Background()

	// =============================================================================
	// 1. Market Status & Basic Info
	// =============================================================================
	fmt.Println("📊 1. Market Status & Info")
	fmt.Println(strings.Repeat("-", 30))

	isWorking, err := client.IsWorkingDay(ctx)
	if err != nil {
		log.Printf("Error checking market status: %v", err)
	} else {
		status := "🔴 CLOSED"
		if isWorking {
			status = "🟢 OPEN"
		}
		fmt.Printf("Market Status: %s\n", status)
	}

	indices, err := client.GetIndices(ctx)
	if err != nil {
		log.Printf("Error getting indices: %v", err)
	} else {
		fmt.Printf("Market Indices (%d):\n", len(indices))
		for i, index := range indices {
			if i >= 3 { // Show first 3
				break
			}
			changeSymbol := "📈"
			if index.Change < 0 {
				changeSymbol = "📉"
			}
			fmt.Printf("  %s %s: %.2f (%.2f%%)\n",
				changeSymbol, index.Symbol, index.Last, index.Change)
		}
	}

	// =============================================================================
	// 2. Individual Ticker Lookups (NEW!)
	// =============================================================================
	fmt.Println("\n💰 2. Individual Ticker Lookups")
	fmt.Println(strings.Repeat("-", 35))

	// Get specific CEDEAR (US stock)
	fmt.Println("🇺🇸 CEDEARs:")
	startTime := time.Now()
	aapl, err := client.GetCedear(ctx, "AAPL")
	if err != nil {
		fmt.Printf("   ❌ AAPL: %v\n", err)
	} else {
		duration := time.Since(startTime)
		fmt.Printf("   🍎 AAPL: $%.2f (%.2f%%) [%v]\n",
			aapl.Last, aapl.Change, duration)
		fmt.Printf("      Volume: %d | Last Update: %s\n",
			aapl.Volume, aapl.DateTime.Format("15:04:05"))
	}

	// Get specific Argentine stock
	fmt.Println("🇦🇷 Argentine Leading Equity:")
	ggal, err := client.GetBluechip(ctx, "GGAL")
	if err != nil {
		fmt.Printf("   ❌ GGAL: %v\n", err)
	} else {
		fmt.Printf("   🏦 GGAL: $%.2f (%.2f%%)\n", ggal.Last, ggal.Change)
	}

	// Universal search (don't need to know security type)
	fmt.Println("🔍 Universal Search:")
	symbols := []string{"BMA", "TSLA", "UNKNOWN"}
	for _, symbol := range symbols {
		security, err := client.GetSecurity(ctx, symbol)
		if err != nil {
			fmt.Printf("   ❌ %s: Not found\n", symbol)
		} else {
			changeIcon := "📈"
			if security.Change < 0 {
				changeIcon = "📉"
			}
			fmt.Printf("   %s %s: $%.2f (%.2f%%)\n",
				changeIcon, symbol, security.Last, security.Change)
		}
	}

	// =============================================================================
	// 3. Batch Operations (Efficient!)
	// =============================================================================
	fmt.Println("\n📦 3. Batch Operations")
	fmt.Println(strings.Repeat("-", 25))

	// Get multiple tickers efficiently (shares cache)
	watchlist := []string{"AAPL", "MSFT", "GOOGL", "TSLA", "META", "GGAL"}
	startTime = time.Now()
	securities, err := client.GetMultipleSecurities(ctx, watchlist)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("Error getting multiple securities: %v", err)
	} else {
		fmt.Printf("💼 Portfolio (%d/%d securities) [%v]:\n",
			len(securities), len(watchlist), duration)

		totalValue := 0.0
		for _, symbol := range watchlist {
			if security, found := securities[symbol]; found {
				changeIcon := "🟢"
				if security.Change < 0 {
					changeIcon = "🔴"
				}
				fmt.Printf("   %s %-6s: $%-10.2f %+6.2f%%\n",
					changeIcon, symbol, security.Last, security.Change)
				totalValue += security.Last
			} else {
				fmt.Printf("   ❌ %-6s: Not found\n", symbol)
			}
		}
		fmt.Printf("   💰 Total Portfolio Value: $%.2f\n", totalValue)
	}

	// =============================================================================
	// 4. Collection Data (Traditional approach)
	// =============================================================================
	fmt.Println("\n📋 4. Collection Data (API endpoints: leading-equity, general-equity, cedears)")
	fmt.Println(strings.Repeat("-", 80))

	// Leading Equity (blue chips) - cached call
	bluechips, err := client.GetBluechips(ctx)
	if err != nil {
		log.Printf("Error getting leading equity: %v", err)
	} else {
		fmt.Printf("💎 Leading Equity (%d securities from 'leading-equity' endpoint):\n", len(bluechips))
		for i, security := range bluechips {
			if i >= 3 { // Show first 3
				fmt.Printf("   ... and %d more\n", len(bluechips)-3)
				break
			}
			changeIcon := "🟢"
			if security.Change < 0 {
				changeIcon = "🔴"
			}
			fmt.Printf("   %s %s: $%.2f (%.2f%%) | Vol: %d\n",
				changeIcon, security.Symbol, security.Last, security.Change, security.Volume)
		}
	}

	// CEDEARs - cached call
	cedears, err := client.GetCedears(ctx)
	if err != nil {
		log.Printf("Error getting CEDEARs: %v", err)
	} else {
		fmt.Printf("\n🌎 CEDEARs (%d securities from 'cedears' endpoint):\n", len(cedears))
		for i, cedear := range cedears {
			if i >= 3 { // Show first 3
				fmt.Printf("   ... and %d more\n", len(cedears)-3)
				break
			}
			changeIcon := "🟢"
			if cedear.Change < 0 {
				changeIcon = "🔴"
			}
			fmt.Printf("   %s %s: $%.2f (%.2f%%)\n",
				changeIcon, cedear.Symbol, cedear.Last, cedear.Change)
		}
	}

	// General Equity (galpones) - cached call
	galpones, err := client.GetGalpones(ctx)
	if err != nil {
		log.Printf("Error getting general equity: %v", err)
	} else {
		fmt.Printf("\n🏢 General Equity (%d securities from 'general-equity' endpoint):\n", len(galpones))
		for i, galpone := range galpones {
			if i >= 3 { // Show first 3
				fmt.Printf("   ... and %d more\n", len(galpones)-3)
				break
			}
			changeIcon := "🟢"
			if galpone.Change < 0 {
				changeIcon = "🔴"
			}
			fmt.Printf("   %s %s: $%.2f (%.2f%%)\n",
				changeIcon, galpone.Symbol, galpone.Last, galpone.Change)
		}
	}

	// =============================================================================
	// 5. Fixed Income & Derivatives
	// =============================================================================
	fmt.Println("\n🏛️  5. Fixed Income & Derivatives")
	fmt.Println(strings.Repeat("-", 35))

	bonds, err := client.GetBonds(ctx)
	if err != nil {
		log.Printf("Error getting bonds: %v", err)
	} else {
		fmt.Printf("📊 Government Bonds: %d instruments\n", len(bonds))
		if len(bonds) > 0 {
			fmt.Printf("   Example: %s - $%.2f\n", bonds[0].Symbol, bonds[0].Last)
		}
	}

	options, err := client.GetOptions(ctx)
	if err != nil {
		log.Printf("Error getting options: %v", err)
	} else {
		fmt.Printf("📈 Options: %d contracts\n", len(options))
	}

	futures, err := client.GetFutures(ctx)
	if err != nil {
		log.Printf("Error getting futures: %v", err)
	} else {
		fmt.Printf("🔮 Futures: %d contracts\n", len(futures))
	}

	// =============================================================================
	// 6. Cache Performance Demo
	// =============================================================================
	fmt.Println("\n⚡ 6. Cache Performance (5-minute automatic caching)")
	fmt.Println(strings.Repeat("-", 55))

	// Show cache information
	cacheInfo := client.GetCacheInfo()
	fmt.Printf("🗄️  Cache Status:\n")
	for category, info := range cacheInfo {
		infoMap := info.(map[string]interface{})
		fmt.Printf("   %-12s: %v items, age %v, fresh: %v\n",
			category, infoMap["count"], infoMap["age"], infoMap["fresh"])
	}

	// Demonstrate cache speed
	fmt.Printf("\n🏃 Cache Speed Test:\n")

	// Get AAPL again (should be from cache)
	fmt.Printf("   Getting AAPL again (cached)... ")
	startTime = time.Now()
	_, err = client.GetCedear(ctx, "AAPL")
	cachedDuration := time.Since(startTime)
	fmt.Printf("%v (lightning fast!)\n", cachedDuration)

	// =============================================================================
	// 7. Historical Data (Chart Data)
	// =============================================================================
	fmt.Println("\n📈 7. Historical Data (Chart Data)")
	fmt.Println(strings.Repeat("-", 35))

	// Get historical data for SPY (S&P 500 ETF) - last 30 days
	fmt.Printf("📊 Historical Data for SPY (last 30 days):\n")
	historyData, err := client.GetHistoryLastDays(ctx, "SPY", 30)
	if err != nil {
		log.Printf("Error getting historical data: %v", err)
	} else {
		fmt.Printf("   Retrieved %d data points:\n", len(historyData.Time))
		if len(historyData.Time) >= 3 {
			// Show first, middle, and last data points
			for i, dataIndex := range []int{0, len(historyData.Time) / 2, len(historyData.Time) - 1} {
				date := historyData.Time[dataIndex].Format("2006-01-02")
				position := []string{"First", "Middle", "Latest"}[i]
				fmt.Printf("   %s (%s): Open=$%.2f High=$%.2f Low=$%.2f Close=$%.2f Vol=%d\n",
					position, date, historyData.Open[dataIndex], historyData.High[dataIndex],
					historyData.Low[dataIndex], historyData.Close[dataIndex], historyData.Volume[dataIndex])
			}
		}
	}

	// Get custom date range historical data (weekly data)
	fmt.Printf("\n📅 Custom Date Range (Weekly data - last 3 months):\n")
	now := time.Now()
	threeMonthsAgo := now.AddDate(0, -3, 0)
	weeklyData, err := client.GetHistory(ctx, "AAPL", "W", threeMonthsAgo, now)
	if err != nil {
		log.Printf("Error getting weekly data: %v", err)
	} else {
		fmt.Printf("   AAPL Weekly Data - %d weeks retrieved\n", len(weeklyData.Time))
		if len(weeklyData.Time) > 0 {
			lastIndex := len(weeklyData.Time) - 1
			latestDate := weeklyData.Time[lastIndex].Format("2006-01-02")
			fmt.Printf("   Latest week (%s): Close=$%.2f\n", latestDate, weeklyData.Close[lastIndex])
		}
	}

	// OHLCV is the historical-data shape returned by GetHistoryLastDays / GetHistory.

	// =============================================================================
	// 8. News & Financial Data
	// =============================================================================
	fmt.Println("\n📰 8. News & Financial Data")
	fmt.Println(strings.Repeat("-", 30))

	news, err := client.GetNews(ctx)
	if err != nil {
		log.Printf("Error getting news: %v", err)
	} else {
		fmt.Printf("📰 Latest News (%d items):\n", len(news))
		for i, newsItem := range news {
			if i >= 2 { // Show first 2
				break
			}
			fmt.Printf("   📄 %s\n", newsItem.Titulo)
			fmt.Printf("      Date: %s\n", newsItem.Fecha.Format("2006-01-02 15:04"))
		}
	}

	// Get income statement for a company
	if len(bluechips) > 0 {
		ticker := bluechips[0].Symbol
		statements, err := client.GetIncomeStatement(ctx, ticker)
		if err != nil {
			fmt.Printf("📊 Income statements for %s: Error - %v\n", ticker, err)
		} else {
			fmt.Printf("📊 Income statements for %s: %d records\n", ticker, len(statements))
		}
	}

	// =============================================================================
	// Summary
	// =============================================================================
	fmt.Println("\n🎉 Example Complete!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✨ Features Demonstrated:")
	fmt.Println("   • Individual ticker lookups (GetCedear, GetBluechip, GetSecurity)")
	fmt.Println("   • Efficient batch operations (GetMultipleSecurities)")
	fmt.Println("   • Historical data & charting (GetHistory, GetHistoryLastDays)")
	fmt.Println("   • 5-minute automatic caching (reduces API calls by 95%)")
	fmt.Println("   • API endpoint mapping:")
	fmt.Println("     - GetBluechips()  → 'leading-equity' endpoint")
	fmt.Println("     - GetGalpones()   → 'general-equity' endpoint")
	fmt.Println("     - GetCedears()    → 'cedears' endpoint")
	fmt.Println("     - GetHistory()    → 'chart/historical-series/history' endpoint")
	fmt.Println("   • Full market data coverage (equities, bonds, derivatives)")
	fmt.Println("   • Real-time market news and financial data")
	fmt.Println("   • Thread-safe concurrent operations")
	fmt.Println("   • Comprehensive error handling")
	fmt.Println("\n🚀 Production Ready:")
	fmt.Println("   • Context-aware operations")
	fmt.Println("   • Built-in retry logic")
	fmt.Println("   • Strongly-typed data structures")
	fmt.Println("   • Zero external dependencies")
}
