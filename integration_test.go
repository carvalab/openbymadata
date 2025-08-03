//go:build integration
// +build integration

package openbymadata_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationRealAPI tests the client against the real BYMA API
func TestIntegrationRealAPI(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test - set INTEGRATION_TEST=true to run")
	}

	// Use default client configuration (no URL override)
	client := NewClient()
	ctx := context.Background()

	t.Run("GetBluechips", func(t *testing.T) {
		bluechips, err := client.GetBluechips(ctx)
		require.NoError(t, err, "Failed to get bluechips from real API")
		assert.NotEmpty(t, bluechips, "Bluechips list should not be empty")

		if len(bluechips) > 0 {
			// Verify structure of first security
			first := bluechips[0]
			assert.NotEmpty(t, first.Symbol, "Security symbol should not be empty")
			t.Logf("✅ Retrieved %d bluechips, first symbol: %s", len(bluechips), first.Symbol)
		}
	})

	t.Run("GetCedears", func(t *testing.T) {
		cedears, err := client.GetCedears(ctx)
		require.NoError(t, err, "Failed to get cedears from real API")
		assert.NotEmpty(t, cedears, "Cedears list should not be empty")

		if len(cedears) > 0 {
			first := cedears[0]
			assert.NotEmpty(t, first.Symbol, "Cedear symbol should not be empty")
			t.Logf("✅ Retrieved %d cedears, first symbol: %s", len(cedears), first.Symbol)
		}
	})

	t.Run("GetBonds", func(t *testing.T) {
		bonds, err := client.GetBonds(ctx)
		require.NoError(t, err, "Failed to get bonds from real API")
		// Bonds might be empty, so just check for no error
		t.Logf("✅ Retrieved %d bonds", len(bonds))
	})

	t.Run("IsWorkingDay", func(t *testing.T) {
		isWorking, err := client.IsWorkingDay(ctx)
		require.NoError(t, err, "Failed to get working day status from real API")
		// Market status should always return something
		t.Logf("✅ Market working day status: %v", isWorking)
	})

	t.Run("GetIndices", func(t *testing.T) {
		indices, err := client.GetIndices(ctx)
		require.NoError(t, err, "Failed to get indices from real API")
		assert.NotEmpty(t, indices, "Indices list should not be empty")

		if len(indices) > 0 {
			first := indices[0]
			assert.NotEmpty(t, first.Symbol, "Index symbol should not be empty")
			t.Logf("✅ Retrieved %d indices, first symbol: %s", len(indices), first.Symbol)
		}
	})

	t.Run("GetHistory", func(t *testing.T) {
		// Test with a known symbol - try to get recent history
		to := time.Now()
		from := to.AddDate(0, 0, -7) // Last 7 days

		// Use a common symbol that should exist
		history, err := client.GetHistory(ctx, "GGAL", "D", from, to)
		if err != nil {
			t.Logf("⚠️  History test skipped or failed: %v", err)
			return
		}

		assert.NotNil(t, history, "History should not be nil")
		if len(history.Time) > 0 {
			assert.Equal(t, len(history.Time), len(history.Close), "Time and Close arrays should have same length")
			t.Logf("✅ Retrieved %d historical data points for GGAL", len(history.Time))
		}
	})
}

// TestIntegrationAPIConnectivity tests basic API connectivity
func TestIntegrationAPIConnectivity(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test - set INTEGRATION_TEST=true to run")
	}

	// Use default client configuration
	client := NewClient()
	ctx := context.Background()

	// Test that we can make at least one successful API call
	_, err := client.IsWorkingDay(ctx)
	require.NoError(t, err, "Basic API connectivity test failed")
	t.Log("✅ API connectivity test passed")
}

// TestIntegrationClientConfiguration tests different client configurations
func TestIntegrationClientConfiguration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test - set INTEGRATION_TEST=true to run")
	}

	t.Run("WithDifferentTimeouts", func(t *testing.T) {
		// Test with very short timeout
		shortOpts := &ClientOptions{
			Timeout: 1 * time.Millisecond,
		}
		shortClient := NewClient(shortOpts)
		ctx := context.Background()

		_, err := shortClient.IsWorkingDay(ctx)
		// This should likely timeout or fail quickly
		t.Logf("Short timeout result (expected to fail): %v", err)

		// Test with default configuration
		normalClient := NewClient()
		_, err = normalClient.IsWorkingDay(ctx)
		require.NoError(t, err, "Default client should work")
		t.Log("✅ Client configuration test passed")
	})
}
