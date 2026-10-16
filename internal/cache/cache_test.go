package cache

import (
	"testing"
	"time"

	"github.com/carvalab/openbymadata/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSet(t *testing.T) {
	c := New()

	_, ok := Get[[]api.Security](c, "bluechips")
	assert.False(t, ok, "empty cache")

	Set(c, "bluechips", []api.Security{{Symbol: "GGAL"}})
	got, ok := Get[[]api.Security](c, "bluechips")
	require.True(t, ok)
	assert.Equal(t, "GGAL", got[0].Symbol)

	_, ok = Get[[]api.Bond](c, "bluechips")
	assert.False(t, ok, "wrong type")
}

func TestExpiry(t *testing.T) {
	c := New()
	c.duration = time.Millisecond
	Set(c, "k", 1)
	time.Sleep(5 * time.Millisecond)

	_, ok := Get[int](c, "k")
	assert.False(t, ok)
	assert.False(t, c.GetInfo()["k"].(map[string]any)["fresh"].(bool))
}

func TestGetInfoCounts(t *testing.T) {
	c := New()
	Set(c, "securities", []api.Security{{}, {}})
	Set(c, "bonds", []api.Bond{{}})
	Set(c, "options", []api.Option{{}})
	Set(c, "futures", []api.Future{{}})
	Set(c, "indices", []api.Index{{}})
	Set(c, "summary", []api.MarketSummary{{}})
	Set(c, "news", []api.News{{}})
	Set(c, IncomeKey("ALUA"), []api.IncomeStatement{{}, {}, {}})
	Set(c, "other", "x")

	info := c.GetInfo()
	count := func(key string) int { return info[key].(map[string]any)["count"].(int) }
	assert.Equal(t, 2, count("securities"))
	assert.Equal(t, 3, count("income:ALUA"))
	for _, k := range []string{"bonds", "options", "futures", "indices", "summary", "news"} {
		assert.Equal(t, 1, count(k), k)
	}
	assert.Equal(t, 0, count("other"))
	assert.True(t, info["bonds"].(map[string]any)["fresh"].(bool))

	c.Clear()
	assert.Empty(t, c.GetInfo())
}
