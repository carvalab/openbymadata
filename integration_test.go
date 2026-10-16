//go:build integration
// +build integration

package openbymadata

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
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test - set INTEGRATION_TEST=true to run")
	}

	client := NewClient()
	defer client.Close()
	ctx := context.Background()

	t.Run("GetBluechips", func(t *testing.T) {
		bluechips, err := client.GetBluechips(ctx)
		require.NoError(t, err, "Failed to get bluechips from real API")
		assert.NotEmpty(t, bluechips, "Bluechips list should not be empty")
		if len(bluechips) > 0 {
			first := bluechips[0]
			assert.NotEmpty(t, first.Symbol, "Security symbol should not be empty")
			t.Logf("Retrieved %d bluechips, first symbol: %s", len(bluechips), first.Symbol)
		}
	})

	t.Run("GetCedears", func(t *testing.T) {
		cedears, err := client.GetCedears(ctx)
		require.NoError(t, err, "Failed to get cedears from real API")
		assert.NotEmpty(t, cedears, "Cedears list should not be empty")
		if len(cedears) > 0 {
			first := cedears[0]
			assert.NotEmpty(t, first.Symbol, "Cedear symbol should not be empty")
			t.Logf("Retrieved %d cedears, first symbol: %s", len(cedears), first.Symbol)
		}
	})

	t.Run("GetBonds", func(t *testing.T) {
		bonds, err := client.GetBonds(ctx)
		require.NoError(t, err, "Failed to get bonds from real API")
		t.Logf("Retrieved %d bonds", len(bonds))
	})

	t.Run("GetMarketTime", func(t *testing.T) {
		mt, err := client.GetMarketTime(ctx)
		require.NoError(t, err, "Failed to get market time from real API")
		assert.NotEmpty(t, mt.OpeningText, "market opening time missing")
		assert.NotEmpty(t, mt.ClosingText, "market closing time missing")
		_, offset := mt.Opens.Zone()
		assert.Equal(t, -3*60*60, offset, "market opens should be in Buenos Aires time")
		assert.Equal(t, time.Now().In(mt.Opens.Location()).YearDay(), mt.Opens.YearDay(), "market opens should be dated today")
		t.Logf("Market working day: %v, opens %s, closes %s (%s)", mt.IsWorkingDay, mt.OpeningText, mt.ClosingText, mt.Timezone)
	})

	t.Run("IsWorkingDay", func(t *testing.T) {
		isWorking, err := client.IsWorkingDay(ctx)
		require.NoError(t, err, "Failed to get working day status from real API")
		t.Logf("Market working day status: %v", isWorking)
	})

	t.Run("GetIndices", func(t *testing.T) {
		indices, err := client.GetIndices(ctx)
		require.NoError(t, err, "Failed to get indices from real API")
		assert.NotEmpty(t, indices, "Indices list should not be empty")
		if len(indices) > 0 {
			first := indices[0]
			assert.NotEmpty(t, first.Symbol, "Index symbol should not be empty")
			t.Logf("Retrieved %d indices, first symbol: %s", len(indices), first.Symbol)
		}
	})

	t.Run("GetHistory", func(t *testing.T) {
		to := time.Now()
		from := to.AddDate(0, 0, -30)
		// History symbols need a settlement suffix; bare "GGAL" returns no_data.
		history, err := client.GetHistory(ctx, "GGAL 24HS", ResolutionDaily, from, to)
		require.NoError(t, err, "Failed to get history from real API")
		require.NotEmpty(t, history.Time, "30 days of daily history should not be empty")
		assert.Equal(t, len(history.Time), len(history.Close), "Time and Close arrays should have same length")
		t.Logf("Retrieved %d historical data points for GGAL 24HS", len(history.Time))
	})

	t.Run("GetIntradayHistory", func(t *testing.T) {
		to := time.Now()
		from := to.AddDate(0, 0, -5)
		// The free endpoint serves delayed intraday data, so only assert the call works.
		bars, err := client.GetIntradayHistory(ctx, "GGAL 24HS", Resolution1Min, from, to)
		require.NoError(t, err, "Failed to get intraday history from real API")
		t.Logf("Retrieved %d 1-minute bars for GGAL 24HS", len(bars.Time))
	})

	t.Run("GetCurrentQuote", func(t *testing.T) {
		// BYMA moved to T+1 in 2024, so 48HS no longer quotes.
		quote, err := client.GetCurrentQuote(ctx, "GGAL", Settlement24HS)
		if err != nil {
			t.Logf("Current quote may be empty outside market hours: %v", err)
			return
		}
		assert.Equal(t, "GGAL", quote.Symbol, "quote symbol mismatch")
		t.Logf("GGAL 24HS quote: last $%.2f, bid $%.2f, ask $%.2f", quote.Last, quote.Bid, quote.Ask)
	})

	t.Run("GetEquityProfile", func(t *testing.T) {
		profile, err := client.GetEquityProfile(ctx, "ALUA")
		require.NoError(t, err, "Failed to get equity profile from real API")
		assert.NotEmpty(t, profile.Fields, "Equity profile fields should not be empty")
		t.Logf("ALUA equity profile fields: %d", len(profile.Fields))
	})

	t.Run("GetCompanyInfo", func(t *testing.T) {
		info, err := client.GetCompanyInfo(ctx, "ALUA")
		require.NoError(t, err, "Failed to get company info from real API")
		assert.NotEmpty(t, info.Fields, "Company info fields should not be empty")
		t.Logf("ALUA company info fields: %d", len(info.Fields))
	})

	t.Run("GetCompanyManagement", func(t *testing.T) {
		directors, err := client.GetCompanyManagement(ctx, "ALUA")
		require.NoError(t, err, "Failed to get company management from real API")
		require.NotEmpty(t, directors)
		assert.NotEmpty(t, directors[0].Name, "director name should be parsed")
		assert.NotEmpty(t, directors[0].Position, "director position should be parsed")
		t.Logf("ALUA directors: %d", len(directors))
	})

	t.Run("MarketResume", func(t *testing.T) {
		summary, err := client.MarketResume(ctx)
		require.NoError(t, err, "Failed to get market summary from real API")
		require.NotEmpty(t, summary)
		assert.Positive(t, summary[0].TotalNegotiated, "total negotiated should be parsed")
	})

	t.Run("GetNews", func(t *testing.T) {
		news, err := client.GetNews(ctx)
		require.NoError(t, err, "Failed to get news from real API")
		require.NotEmpty(t, news)
		assert.Regexp(t, `/sba/download/\d+$`, news[0].Descarga, "news download link should carry the file id")
	})

	t.Run("GetIncomeStatement", func(t *testing.T) {
		statements, err := client.GetIncomeStatement(ctx, "ALUA")
		require.NoError(t, err, "Failed to get income statements from real API")
		require.NotEmpty(t, statements)
		assert.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, statements[0].FechaCierre)
		assert.Regexp(t, `/sba/download/\d+$`, statements[0].BalancesArchivo)
	})

	t.Run("GetETFs", func(t *testing.T) {
		etfs, err := client.GetETFs(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, etfs)
		assert.NotEmpty(t, etfs[0].Symbol)
		assert.NotEmpty(t, etfs[0].Currency)
	})

	t.Run("GetMEPIndices", func(t *testing.T) {
		mep, err := client.GetMEPIndices(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, mep)
		assert.Positive(t, mep[0].Last, "MEP rate should be parsed")
		t.Logf("%s: %.2f", mep[0].Symbol, mep[0].Last)
	})

	t.Run("GetTopMovers", func(t *testing.T) {
		movers, err := client.GetTopMovers(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, movers.Stocks)
		assert.NotEmpty(t, movers.Bonds)
	})

	t.Run("GetIndexHistory", func(t *testing.T) {
		to := time.Now()
		h, err := client.GetIndexHistory(ctx, "M", ResolutionDaily, to.AddDate(0, 0, -30), to)
		require.NoError(t, err)
		require.NotEmpty(t, h.Time, "30 days of Merval history should not be empty")
		assert.Positive(t, h.Close[len(h.Close)-1])
	})

	t.Run("BondLists", func(t *testing.T) {
		for name, get := range map[string]func(context.Context) ([]Bond, error){
			"green":          client.GetGreenBonds,
			"social":         client.GetSocialBonds,
			"sustainability": client.GetSustainabilityBonds,
			"linked":         client.GetLinkedBonds,
			"senebi":         client.GetSenebiBonds,
			"senebi letters": client.GetSenebiLetters,
			"senebi corp":    client.GetSenebiCorporateBonds,
		} {
			bonds, err := get(ctx)
			require.NoError(t, err, name)
			t.Logf("%s: %d", name, len(bonds))
		}
		senebi, err := client.GetSenebiBonds(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, senebi)
		assert.Equal(t, "SENEBI", senebi[0].Market)
	})

	t.Run("GetBondIndices", func(t *testing.T) {
		idx, err := client.GetBondIndices(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, idx)
		assert.Positive(t, idx[0].Value)
	})

	t.Run("GetBondCoupons", func(t *testing.T) {
		coupons, err := client.GetBondCoupons(ctx, "GD30")
		require.NoError(t, err)
		require.NotEmpty(t, coupons)
		assert.False(t, coupons[0].InterestPaymentDate.IsZero(), "payment date should be parsed")
	})

	t.Run("GetMarketResponsibles", func(t *testing.T) {
		resp, err := client.GetMarketResponsibles(ctx, "ALUA")
		require.NoError(t, err)
		require.NotEmpty(t, resp)
		assert.NotEmpty(t, resp[0].Name)
	})

	t.Run("SearchInstruments", func(t *testing.T) {
		found, err := client.SearchInstruments(ctx, "GGAL")
		require.NoError(t, err)
		require.NotEmpty(t, found)
		assert.Equal(t, "GGAL", found[0].Symbol)
	})

	t.Run("DeprecatedAliases", func(t *testing.T) {
		bluechips, err := client.GetLeadingEquity(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, bluechips)
		galpones, err := client.GetGeneralEquity(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, galpones)
		letters, err := client.GetLetters(ctx)
		require.NoError(t, err)
		_ = letters
	})
}

// TestIntegrationAPIConnectivity tests basic API connectivity
func TestIntegrationAPIConnectivity(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test - set INTEGRATION_TEST=true to run")
	}

	client := NewClient()
	defer client.Close()
	ctx := context.Background()

	_, err := client.IsWorkingDay(ctx)
	require.NoError(t, err, "Basic API connectivity test failed")
	t.Log("API connectivity test passed")
}

// TestIntegrationClientConfiguration tests different client configurations
func TestIntegrationClientConfiguration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test - set INTEGRATION_TEST=true to run")
	}

	t.Run("WithDifferentTimeouts", func(t *testing.T) {
		shortOpts := &ClientOptions{
			Timeout: 1 * time.Millisecond,
		}
		shortClient := NewClient(shortOpts)
		defer shortClient.Close()
		ctx := context.Background()

		_, err := shortClient.IsWorkingDay(ctx)
		t.Logf("Short timeout result (expected to fail): %v", err)

		normalClient := NewClient()
		defer normalClient.Close()
		_, err = normalClient.IsWorkingDay(ctx)
		require.NoError(t, err, "Default client should work")
		t.Log("Client configuration test passed")
	})
}
