package helpers

import (
	"testing"

	"github.com/carvalab/openbymadata/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	bluechips = []api.Security{{Symbol: "GGAL"}, {Symbol: "YPFD"}}
	cedears   = []api.Security{{Symbol: "AAPL"}, {Symbol: "GOOGL"}}
	galpones  = []api.Security{{Symbol: "A3"}, {Symbol: "GGAL", Group: "dup"}}
)

func TestFindSecurityBySymbol(t *testing.T) {
	for _, sym := range []string{"YPFD", "GOOGL", "A3"} {
		s, err := FindSecurityBySymbol(sym, bluechips, cedears, galpones)
		require.NoError(t, err, sym)
		assert.Equal(t, sym, s.Symbol)
	}

	// Bluechips win over galpones for duplicated symbols.
	s, err := FindSecurityBySymbol("GGAL", bluechips, cedears, galpones)
	require.NoError(t, err)
	assert.Empty(t, s.Group)

	_, err = FindSecurityBySymbol("NOPE", bluechips, cedears, galpones)
	assert.EqualError(t, err, "security NOPE not found")
}

func TestFindSecurityInCollection(t *testing.T) {
	s, err := FindSecurityInCollection("AAPL", cedears)
	require.NoError(t, err)
	assert.Equal(t, "AAPL", s.Symbol)

	_, err = FindSecurityInCollection("AAPL", bluechips)
	assert.Error(t, err)
}

func TestFindBondBySymbol(t *testing.T) {
	bonds := []api.Bond{{Symbol: "AL30"}}
	short := []api.Bond{{Symbol: "S31O6"}}
	corp := []api.Bond{{Symbol: "YCA6O"}}
	for _, sym := range []string{"AL30", "S31O6", "YCA6O"} {
		b, err := FindBondBySymbol(sym, bonds, short, corp)
		require.NoError(t, err, sym)
		assert.Equal(t, sym, b.Symbol)
	}
	_, err := FindBondBySymbol("NOPE", bonds, short, corp)
	assert.EqualError(t, err, "bond NOPE not found")
}

func TestFindOptionAndFuture(t *testing.T) {
	o, err := FindOptionBySymbol("GFGC100", []api.Option{{Symbol: "GFGC100"}})
	require.NoError(t, err)
	assert.Equal(t, "GFGC100", o.Symbol)
	_, err = FindOptionBySymbol("X", nil)
	assert.EqualError(t, err, "option X not found")

	f, err := FindFutureBySymbol("DLR/NOV26", []api.Future{{Symbol: "DLR/NOV26"}})
	require.NoError(t, err)
	assert.Equal(t, "DLR/NOV26", f.Symbol)
	_, err = FindFutureBySymbol("X", nil)
	assert.EqualError(t, err, "future X not found")
}

func TestGetMultipleSecurities(t *testing.T) {
	got := GetMultipleSecurities([]string{"GGAL", "AAPL", "A3", "NOPE"}, bluechips, cedears, galpones)
	assert.Len(t, got, 3)
	assert.Empty(t, got["GGAL"].Group, "bluechip wins")
	assert.NotContains(t, got, "NOPE")
}

func TestSearchSecurities(t *testing.T) {
	got := SearchSecurities("g", bluechips, cedears, galpones)
	var symbols []string
	for _, s := range got {
		symbols = append(symbols, s.Symbol)
	}
	assert.Equal(t, []string{"GGAL", "GOOGL", "GGAL"}, symbols)
	assert.Empty(t, SearchSecurities("zzz", bluechips, cedears, galpones))
}
