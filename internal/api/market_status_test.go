package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReposLendingTrusts(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	repos, err := c.GetRepos(ctx)
	require.NoError(t, err)
	require.Len(t, repos, 3)
	assert.Equal(t, "CABLE-1611-U-CT-EXT", repos[0].Symbol)
	assert.Equal(t, "CABLE", repos[0].Underlying)
	assert.Equal(t, "EXT", repos[0].Currency)
	assert.Equal(t, 0.02, repos[0].Bid)
	assert.Equal(t, int64(31), repos[0].DaysToMaturity)
	assert.Equal(t, time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC), repos[0].Expiration)

	lending, err := c.GetSecuritiesLending(ctx)
	require.NoError(t, err)
	assert.Equal(t, "AAL-1910-L-CT-ARS", lending[0].Symbol)
	assert.Equal(t, "AAL", lending[0].Underlying)
	assert.Equal(t, 3.0, lending[0].Last)

	trusts, err := c.GetFinancialTrusts(ctx)
	require.NoError(t, err)
	assert.Equal(t, "AD01C.SB", trusts[0].Symbol)
	assert.Equal(t, "SENEBI", trusts[0].Market)
	assert.Empty(t, trusts[0].Underlying)
}

func TestCorporateGovernanceAndCatalog(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	gov, err := c.GetCorporateGovernance(ctx)
	require.NoError(t, err)
	assert.Equal(t, "BMA", gov[0].Symbol)
	assert.Equal(t, 10280.0, gov[0].Last)

	plus, err := c.GetCorporateGovernancePlus(ctx)
	require.NoError(t, err)
	assert.Equal(t, "BYMA", plus[0].Symbol)
	assert.Equal(t, 252.5, plus[0].Last)

	catalog, err := c.GetInstrumentCatalog(ctx)
	require.NoError(t, err)
	require.Len(t, catalog, 3)
	assert.Equal(t, CatalogInstrument{ID: 1732275, Code: "OM1BL.SB", Kind: "BONDINSTRUMENT"}, catalog[0])
}

func TestMarketOpenAndServerTime(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	open, err := c.IsMarketOpen(ctx)
	require.NoError(t, err)
	assert.True(t, open)

	closed, err := newTestClient(fixtureServerWith(t, "false").URL).IsMarketOpen(ctx)
	require.NoError(t, err)
	assert.False(t, closed)

	st, err := c.GetServerTime(ctx)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 10, 16, 15, 51, 44, 2430591, buenosAires), st)
	_, offset := st.Zone()
	assert.Equal(t, -3*3600, offset)
}

func TestMarketStatusErrors(t *testing.T) {
	ctx := context.Background()

	bad := newTestClient(fixtureServerWith(t, `"maybe"`).URL)
	_, err := bad.IsMarketOpen(ctx)
	assert.ErrorContains(t, err, "invalid market-open")
	_, err = bad.GetServerTime(ctx)
	assert.ErrorContains(t, err, "invalid server-time")

	_, err = newTestClient(fixtureServerWith(t, `{"datetime":"yesterday"}`).URL).GetServerTime(ctx)
	assert.ErrorContains(t, err, `invalid server-time "yesterday"`)

	down := newTestClient(fixtureServerStatus(t, 404).URL)
	_, err = down.IsMarketOpen(ctx)
	assert.ErrorContains(t, err, "HTTP error 404")
	_, err = down.GetServerTime(ctx)
	assert.ErrorContains(t, err, "HTTP error 404")
	_, err = down.GetInstrumentCatalog(ctx)
	assert.ErrorContains(t, err, "HTTP error 404")

	_, err = newTestClient(fixtureServerWith(t, `not json`).URL).GetInstrumentCatalog(ctx)
	assert.ErrorContains(t, err, "invalid catalog")
}

func TestHistoryUpstreamError(t *testing.T) {
	c := newTestClient(fixtureServerWith(t, `{"s":"error","errmsg":"Error interno"}`).URL)
	_, err := c.GetHistory(context.Background(), "GGAL 24HS", ResolutionDaily, time.Now(), time.Now())
	assert.EqualError(t, err, "history for GGAL 24HS: Error interno")
}
