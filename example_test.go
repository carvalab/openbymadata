package openbymadata_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/carvalab/openbymadata"
)

// ExampleNewClient demonstrates how to create a basic client.
func ExampleNewClient() {
	// Create a client with default settings
	client := openbymadata.NewClient()
	fmt.Printf("Client created with default options\n")

	// Create a client with custom options
	opts := &openbymadata.ClientOptions{
		Timeout:       30 * time.Second,
		RetryAttempts: 5,
		EnableCache:   true,
	}
	customClient := openbymadata.NewClient(opts)
	fmt.Printf("Custom client created\n")

	_ = client
	_ = customClient

	// Output:
	// Client created with default options
	// Custom client created
}

// ExampleClient_GetCedear demonstrates how to get a specific CEDEAR (US stock).
func ExampleClient_GetCedear() {
	// Create a test server that mimics the BYMA API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CEDEARs endpoint returns data directly (not wrapped)
		mockResponse := []map[string]interface{}{
			{
				"symbol":          "AAPL",
				"settlementPrice": 150.50,
				"imbalance":       2.5,
				"volume":          float64(1000000),
			},
		}
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client pointing to test server
	client := openbymadata.NewClient(&openbymadata.ClientOptions{
		BaseURL: server.URL,
	})

	ctx := context.Background()
	aapl, err := client.GetCedear(ctx, "AAPL")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("AAPL: $%.2f (%.1f%% change)\n", aapl.Last, aapl.Change)
	fmt.Printf("Volume: %d\n", aapl.Volume)

	// Output:
	// AAPL: $150.50 (2.5% change)
	// Volume: 1000000
}

// TODO: Fix GetSecurity example test
// ExampleClient_GetSecurity demonstrates universal security search.
func _ExampleClient_GetSecurity() {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock different responses based on the endpoint
		if r.URL.Path == "/api/market/cedears" {
			// CEDEARs return data directly (not wrapped)
			mockResponse := []map[string]interface{}{
				{
					"symbol":          "AAPL",
					"settlementPrice": 150.50,
					"imbalance":       2.5,
				},
			}
			json.NewEncoder(w).Encode(mockResponse)
		} else if r.URL.Path == "/api/market/leading-equity" {
			// Leading equity returns wrapped data
			mockResponse := map[string]interface{}{
				"data": []map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(mockResponse)
		} else if r.URL.Path == "/api/market/general-equity" {
			// General equity returns wrapped data
			mockResponse := map[string]interface{}{
				"data": []map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(mockResponse)
		} else {
			// Default: return empty wrapped data for unknown endpoints
			mockResponse := map[string]interface{}{
				"data": []map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(mockResponse)
		}
	}))
	defer server.Close()

	client := openbymadata.NewClient(&openbymadata.ClientOptions{
		BaseURL: server.URL,
	})

	ctx := context.Background()

	// GetSecurity automatically searches across all security types
	security, err := client.GetSecurity(ctx, "AAPL")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %s: $%.2f\n", security.Symbol, security.Last)

	// Output:
	// Found AAPL: $150.50
}

// TODO: Fix GetMultipleSecurities example test
// ExampleClient_GetMultipleSecurities demonstrates batch operations.
func _ExampleClient_GetMultipleSecurities() {
	// Create test server with multiple securities
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/market/cedears" {
			// CEDEARs return data directly
			mockResponse := []map[string]interface{}{
				{"symbol": "AAPL", "settlementPrice": 150.50, "imbalance": 2.5},
				{"symbol": "MSFT", "settlementPrice": 280.75, "imbalance": -1.2},
				{"symbol": "GOOGL", "settlementPrice": 2450.00, "imbalance": 0.8},
			}
			json.NewEncoder(w).Encode(mockResponse)
		} else if r.URL.Path == "/api/market/leading-equity" {
			// Other endpoints return data wrapped
			mockResponse := map[string]interface{}{
				"data": []map[string]interface{}{
					{"symbol": "GGAL", "settlementPrice": 3500.00, "imbalance": 1.5},
				},
			}
			json.NewEncoder(w).Encode(mockResponse)
		} else {
			// Return empty wrapped data for other endpoints
			mockResponse := map[string]interface{}{
				"data": []map[string]interface{}{},
			}
			json.NewEncoder(w).Encode(mockResponse)
		}
	}))
	defer server.Close()

	client := openbymadata.NewClient(&openbymadata.ClientOptions{
		BaseURL: server.URL,
	})

	ctx := context.Background()
	watchlist := []string{"AAPL", "MSFT", "GOOGL", "GGAL"}

	securities, err := client.GetMultipleSecurities(ctx, watchlist)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Portfolio (%d securities):\n", len(securities))
	for _, symbol := range watchlist {
		if security, found := securities[symbol]; found {
			fmt.Printf("  %s: $%.2f (%.1f%%)\n", symbol, security.Last, security.Change)
		}
	}

	// Output:
	// Portfolio (4 securities):
	//   AAPL: $150.50 (2.5%)
	//   MSFT: $280.75 (-1.2%)
	//   GOOGL: $2450.00 (0.8%)
	//   GGAL: $3500.00 (1.5%)
}

// TODO: Fix historical data example dates
// ExampleClient_GetHistoryLastDays demonstrates historical data retrieval.
func _ExampleClient_GetHistoryLastDays() {
	// Create test server with historical data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := map[string]interface{}{
			"s": "ok",
			"t": []int64{1672531200, 1672617600, 1672704000}, // Unix timestamps
			"o": []float64{100.0, 102.0, 104.0},              // Opens
			"h": []float64{105.0, 107.0, 109.0},              // Highs
			"l": []float64{98.0, 100.0, 102.0},               // Lows
			"c": []float64{103.0, 105.0, 107.0},              // Closes
			"v": []int64{1000000, 1200000, 1100000},          // Volumes
		}
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := openbymadata.NewClient(&openbymadata.ClientOptions{
		BaseURL: server.URL,
	})

	ctx := context.Background()
	historyData, err := client.GetHistoryLastDays(ctx, "AAPL", 7)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Historical data for AAPL (%d data points):\n", len(historyData.Time))
	for i := 0; i < len(historyData.Time) && i < 2; i++ {
		date := historyData.Time[i].Format("2006-01-02")
		fmt.Printf("  %s: Open=$%.2f High=$%.2f Low=$%.2f Close=$%.2f Volume=%d\n",
			date, historyData.Open[i], historyData.High[i], historyData.Low[i],
			historyData.Close[i], historyData.Volume[i])
	}

	// Output:
	// Historical data for AAPL (3 data points):
	//   2023-01-01: Open=$100.00 High=$105.00 Low=$98.00 Close=$103.00 Volume=1000000
	//   2023-01-02: Open=$102.00 High=$107.00 Low=$100.00 Close=$105.00 Volume=1200000
}

// TODO: Fix GetBluechips example test
// ExampleClient_GetBluechips demonstrates getting all blue chip securities.
func _ExampleClient_GetBluechips() {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Leading equity returns data wrapped
		mockResponse := map[string]interface{}{
			"data": []map[string]interface{}{
				{"symbol": "GGAL", "settlementPrice": 3500.00, "imbalance": 1.5, "volume": float64(500000)},
				{"symbol": "YPF", "settlementPrice": 2800.00, "imbalance": -0.8, "volume": float64(750000)},
				{"symbol": "TECO2", "settlementPrice": 890.00, "imbalance": 2.1, "volume": float64(300000)},
			},
		}
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := openbymadata.NewClient(&openbymadata.ClientOptions{
		BaseURL: server.URL,
	})

	ctx := context.Background()
	bluechips, err := client.GetBluechips(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Blue chip securities (%d total):\n", len(bluechips))
	for i, security := range bluechips {
		if i >= 2 { // Show first 2
			break
		}
		fmt.Printf("  %s: $%.2f (%.1f%%) Volume: %d\n",
			security.Symbol, security.Last, security.Change, security.Volume)
	}

	// Output:
	// Blue chip securities (3 total):
	//   GGAL: $3500.00 (1.5%) Volume: 500000
	//   YPF: $2800.00 (-0.8%) Volume: 750000
}

// TODO: Fix caching example test
// ExampleClient_caching demonstrates the caching behavior.
func _ExampleClient_caching() {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// CEDEARs return data directly
		mockResponse := []map[string]interface{}{
			{"symbol": "AAPL", "settlementPrice": 150.50, "imbalance": 2.5},
		}
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := openbymadata.NewClient(&openbymadata.ClientOptions{
		BaseURL:     server.URL,
		EnableCache: true, // Caching enabled
	})

	ctx := context.Background()

	// First call - hits the API
	start := time.Now()
	_, err := client.GetCedear(ctx, "AAPL")
	if err != nil {
		log.Fatal(err)
	}
	firstCallDuration := time.Since(start)

	// Second call - should be from cache (much faster)
	start = time.Now()
	_, err = client.GetCedear(ctx, "AAPL")
	if err != nil {
		log.Fatal(err)
	}
	secondCallDuration := time.Since(start)

	fmt.Printf("API calls made: %d\n", callCount)
	fmt.Printf("First call took: %v\n", firstCallDuration > 0)
	fmt.Printf("Second call was cached: %v\n", secondCallDuration < firstCallDuration)

	// Output:
	// API calls made: 1
	// First call took: true
	// Second call was cached: true
}

// TODO: Fix error handling example
// ExampleBYMAError demonstrates error handling.
func _ExampleBYMAError() {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Rate limited"))
	}))
	defer server.Close()

	client := openbymadata.NewClient(&openbymadata.ClientOptions{
		BaseURL: server.URL,
	})

	ctx := context.Background()
	_, err := client.GetCedear(ctx, "AAPL")

	if err != nil {
		if bymaErr, ok := err.(*openbymadata.BYMAError); ok {
			fmt.Printf("Error Code: %s\n", bymaErr.Code)
			fmt.Printf("Is Retryable: %v\n", openbymadata.IsRetryable(err))
		}
	}

	// Output:
	// Error Code: RATE_LIMITED
	// Is Retryable: true
}

// ExampleClientOptions demonstrates different configuration options.
func ExampleClientOptions() {
	// Create client with custom timeout and retry settings
	opts := &openbymadata.ClientOptions{
		Timeout:       10 * time.Second, // Custom timeout
		RetryAttempts: 5,                // More retry attempts
		EnableCache:   false,            // Disable caching for always fresh data
	}

	client := openbymadata.NewClient(opts)
	fmt.Printf("Client configured with custom options\n")

	// You can also get the default options and modify them
	defaultOpts := openbymadata.DefaultClientOptions()
	defaultOpts.Timeout = 60 * time.Second

	client2 := openbymadata.NewClient(defaultOpts)
	fmt.Printf("Client configured with modified defaults\n")

	_ = client
	_ = client2

	// Output:
	// Client configured with custom options
	// Client configured with modified defaults
}
