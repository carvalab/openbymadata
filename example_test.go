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
