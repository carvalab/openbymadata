package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestETFsAndCurrency(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)

	etfs, err := c.GetETFs(context.Background())
	require.NoError(t, err)
	require.Len(t, etfs, 3)
	assert.Equal(t, "ARKK", etfs[0].Symbol)
	assert.Equal(t, 13300.0, etfs[0].Last)
	assert.Equal(t, "ARS", etfs[0].Currency)
	assert.Equal(t, "BYMA", etfs[0].Market)
}

func TestThematicAndSenebiBonds(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	tests := []struct {
		name     string
		get      func(context.Context) ([]Bond, error)
		symbol   string
		market   string
		currency string
		days     int64
	}{
		{"green", c.GetGreenBonds, "EAC2O", "BYMA", "ARS", 2553},
		{"sustainability", c.GetSustainabilityBonds, "IP01A", "BYMA", "ARS", 267},
		{"senebi bonds", c.GetSenebiBonds, "AE38.SB", "SENEBI", "ARS", 4103},
		{"senebi letters", c.GetSenebiLetters, "B2D6C.SB", "SENEBI", "EXT", 62},
		{"senebi corporate", c.GetSenebiCorporateBonds, "A11LD.SB", "SENEBI", "EXT", 532},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonds, err := tt.get(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, bonds)
			assert.Equal(t, tt.symbol, bonds[0].Symbol)
			assert.Equal(t, tt.market, bonds[0].Market)
			assert.Equal(t, tt.currency, bonds[0].Currency)
			assert.Equal(t, tt.days, bonds[0].DaysToMaturity)
		})
	}

	// Recorded while BYMA listed none; must still parse as empty, not error.
	social, err := c.GetSocialBonds(ctx)
	require.NoError(t, err)
	assert.Empty(t, social)
	linked, err := c.GetLinkedBonds(ctx)
	require.NoError(t, err)
	assert.Empty(t, linked)
}

func TestBondIndicesAndCoupons(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	idx, err := c.GetBondIndices(ctx)
	require.NoError(t, err)
	require.Len(t, idx, 3)
	assert.Equal(t, "COR_DOL", idx[0].Symbol)
	assert.Equal(t, 686692.991342, idx[0].Value)
	assert.Equal(t, 1.00027433, idx[0].Varianza)
	assert.Equal(t, time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC), idx[0].Date)

	coupons, err := c.GetBondCoupons(ctx, "gd30")
	require.NoError(t, err)
	require.Len(t, coupons, 3)
	first := coupons[0]
	assert.Equal(t, "GD30", first.Symbol)
	assert.Equal(t, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), first.NoticeDate)
	assert.Equal(t, time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC), first.InterestPaymentDate)
	assert.Equal(t, time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC), first.AmortizationPaymentDate)
	assert.Equal(t, int64(11), first.InterestServiceNumber)
	assert.Equal(t, int64(5), first.AmortizationServiceNumber)
	assert.Equal(t, 8.0, first.AmortizationPercent)
	assert.Equal(t, 64.0, first.ResidualValuePercent)
	assert.Equal(t, downloadBase+"495839", first.NoticeURL)
}

func TestMEPAndTopMovers(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	mep, err := c.GetMEPIndices(ctx)
	require.NoError(t, err)
	require.Len(t, mep, 3)
	assert.Equal(t, "IDXDOLARB", mep[0].Symbol)
	assert.InDelta(t, 1536.97, mep[0].Last, 0.01)
	assert.InDelta(t, 1539.88, mep[0].High, 0.01)
	assert.InDelta(t, 1532.46, mep[0].Low, 0.01)

	movers, err := c.GetTopMovers(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"GCLA", "DOME", "BBARC"}, symbols(movers.Stocks, func(s Security) string { return s.Symbol }))
	assert.Equal(t, []string{"SA24D", "GD46C", "BB37D"}, symbols(movers.Bonds, func(b Bond) string { return b.Symbol }))
	require.Len(t, movers.Options, 3)
	assert.Equal(t, "GFGC6200OC", movers.Options[0].Symbol)
	assert.Equal(t, "CALL", movers.Options[0].OptionType)
	assert.Equal(t, "GGAL", movers.Options[0].UnderlyingAsset)
}

func TestIndexHistoryResponsiblesAndSearch(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	h, err := c.GetIndexHistory(ctx, "M", ResolutionDaily, time.Now().AddDate(0, 0, -30), time.Now())
	require.NoError(t, err)
	require.Len(t, h.Time, 5)
	assert.Equal(t, int64(1789527600), h.Time[0].Unix())
	assert.Equal(t, 3028870.87, h.Close[0])

	resp, err := c.GetMarketResponsibles(ctx, "ALUA")
	require.NoError(t, err)
	require.Len(t, resp, 2)
	assert.Equal(t, "Martínez Costa, Alberto Eduardo", resp[0].Name)
	assert.Equal(t, "Titular", resp[0].Position)
	assert.Equal(t, time.Date(2001, 11, 30, 0, 0, 0, 0, time.UTC), resp[0].Since)

	found, err := c.SearchInstruments(ctx, "GGAL")
	require.NoError(t, err)
	require.Len(t, found, 3)
	assert.Equal(t, InstrumentMatch{
		Symbol: "GGAL", Description: "GRUPO FINANCIERO GALICIA - ORD. B 1 V.", Type: "Accion",
		Market: "BYMA", Settlement: "24hs", Presentation: "GGAL",
	}, found[0])
}

func TestNewEndpointsInvalidJSON(t *testing.T) {
	c := newTestClient(fixtureServerWith(t, `not json`).URL)
	ctx := context.Background()
	calls := map[string]func() error{
		"etfs":         func() error { _, err := c.GetETFs(ctx); return err },
		"bond indices": func() error { _, err := c.GetBondIndices(ctx); return err },
		"coupons":      func() error { _, err := c.GetBondCoupons(ctx, "X"); return err },
		"mep":          func() error { _, err := c.GetMEPIndices(ctx); return err },
		"top movers":   func() error { _, err := c.GetTopMovers(ctx); return err },
		"responsibles": func() error { _, err := c.GetMarketResponsibles(ctx, "X"); return err },
		"search":       func() error { _, err := c.SearchInstruments(ctx, "X"); return err },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) { assert.Error(t, call()) })
	}
}

func TestNewEndpointsHTTPError(t *testing.T) {
	c := newTestClient(fixtureServerStatus(t, 404).URL)
	ctx := context.Background()
	calls := map[string]func() error{
		"bond indices": func() error { _, err := c.GetBondIndices(ctx); return err },
		"coupons":      func() error { _, err := c.GetBondCoupons(ctx, "X"); return err },
		"mep":          func() error { _, err := c.GetMEPIndices(ctx); return err },
		"top movers":   func() error { _, err := c.GetTopMovers(ctx); return err },
		"responsibles": func() error { _, err := c.GetMarketResponsibles(ctx, "X"); return err },
		"search":       func() error { _, err := c.SearchInstruments(ctx, "X"); return err },
		"index history": func() error {
			_, err := c.GetIndexHistory(ctx, "M", ResolutionDaily, time.Now(), time.Now())
			return err
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) { assert.ErrorContains(t, call(), "HTTP error 404") })
	}
}

func symbols[T any](items []T, symbol func(T) string) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = symbol(it)
	}
	return out
}
