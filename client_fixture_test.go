package openbymadata

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newFixtureClient serves the recorded API responses in internal/api/testdata
// and counts requests per endpoint.
func newFixtureClient(t *testing.T, enableCache bool) (Client, map[string]*atomic.Int32) {
	t.Helper()
	counts := map[string]*atomic.Int32{}
	for _, f := range mustGlob(t, "internal/api/testdata/*.json") {
		counts[strings.TrimSuffix(filepath.Base(f), ".json")] = &atomic.Int32{}
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.ReplaceAll(strings.TrimPrefix(r.URL.Path, "/vanoms-be-core/rest/api/bymadata/free/"), "/", "_")
		body, err := os.ReadFile(filepath.Join("internal/api/testdata", name+".json"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		counts[name].Add(1)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	c := NewClient(&ClientOptions{BaseURL: srv.URL, EnableCache: enableCache})
	t.Cleanup(c.Close)
	return c, counts
}

func mustGlob(t *testing.T, pattern string) []string {
	t.Helper()
	files, err := filepath.Glob(pattern)
	require.NoError(t, err)
	require.NotEmpty(t, files)
	return files
}

func TestClientLookups(t *testing.T) {
	c, _ := newFixtureClient(t, true)
	ctx := context.Background()

	s, err := c.GetSecurity(ctx, "AAL")
	require.NoError(t, err)
	assert.Equal(t, "AAL", s.Symbol)

	s, err = c.GetBluechip(ctx, "ALUA")
	require.NoError(t, err)
	assert.Equal(t, 846.0, s.Last)

	s, err = c.GetGalpone(ctx, "A3")
	require.NoError(t, err)
	assert.Equal(t, "A3", s.Symbol)

	_, err = c.GetCedear(ctx, "NOPE")
	assert.Error(t, err)

	for _, sym := range []string{"AE38", "B2D6C", "A11LD"} {
		b, err := c.GetBond(ctx, sym)
		require.NoError(t, err, sym)
		assert.Equal(t, sym, b.Symbol)
	}

	o, err := c.GetOption(ctx, "ALUC1000DI")
	require.NoError(t, err)
	assert.Equal(t, "ALUA", o.UnderlyingAsset)

	f, err := c.GetFuture(ctx, "DLR/NOV26")
	require.NoError(t, err)
	assert.InDelta(t, 1455.0, f.Last, 1e-9)

	multi, err := c.GetMultipleSecurities(ctx, []string{"ALUA", "AAL", "A3", "NOPE"})
	require.NoError(t, err)
	assert.Len(t, multi, 3)

	found, err := c.SearchSecurities(ctx, "alu")
	require.NoError(t, err)
	require.NotEmpty(t, found)
	assert.Equal(t, "ALUA", found[0].Symbol)
}

func TestClientEndpoints(t *testing.T) {
	c, _ := newFixtureClient(t, false)
	ctx := context.Background()

	idx, err := c.GetIndices(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, idx)

	sum, err := c.MarketResume(ctx)
	require.NoError(t, err)
	assert.Equal(t, 11325569533965.3, sum[0].TotalNegotiated)

	news, err := c.GetNews(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, news[0].Descarga)

	inc, err := c.GetIncomeStatement(ctx, "ALUA")
	require.NoError(t, err)
	assert.Equal(t, "2026-06-30", inc[0].FechaCierre)

	mt, err := c.GetMarketTime(ctx)
	require.NoError(t, err)
	assert.Equal(t, "10:30:00", mt.OpeningText)
	assert.Equal(t, 10, mt.Opens.Hour())

	q, err := c.GetCurrentQuote(ctx, "GGAL", Settlement24HS)
	require.NoError(t, err)
	assert.Equal(t, 5840.0, q.Ask)

	from, to := time.Now().AddDate(0, 0, -30), time.Now()
	h, err := c.GetHistory(ctx, "GGAL 24HS", ResolutionDaily, from, to)
	require.NoError(t, err)
	assert.Len(t, h.Time, 5)
	h, err = c.GetHistoryLastDays(ctx, "GGAL 24HS", 30)
	require.NoError(t, err)
	assert.Len(t, h.Time, 5)
	h, err = c.GetIntradayHistory(ctx, "GGAL 24HS", "", from, to)
	require.NoError(t, err, "empty resolution defaults to 1 minute")
	assert.Len(t, h.Time, 5)

	info, err := c.GetCompanyInfo(ctx, "ALUA")
	require.NoError(t, err)
	assert.NotEmpty(t, info.Fields)
	prof, err := c.GetEquityProfile(ctx, "ALUA")
	require.NoError(t, err)
	assert.NotEmpty(t, prof.Fields)
	dirs, err := c.GetCompanyManagement(ctx, "ALUA")
	require.NoError(t, err)
	assert.Equal(t, "LEVINAS MARTIN JOSE", dirs[0].Name)
	bal, err := c.GetCompanyBalance(ctx, "ALUA")
	require.NoError(t, err)
	assert.Len(t, bal, 2)

	lead, err := c.GetLeadingEquity(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, lead)
	gen, err := c.GetGeneralEquity(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, gen)
	pub, err := c.GetPublicBonds(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, pub)
	let, err := c.GetLetters(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, let)
}

func TestClientCache(t *testing.T) {
	ctx := context.Background()

	c, counts := newFixtureClient(t, true)
	for range 3 {
		_, err := c.GetBluechips(ctx)
		require.NoError(t, err)
	}
	assert.Equal(t, int32(1), counts["leading-equity"].Load(), "cached after first call")
	assert.Contains(t, c.GetCacheInfo(), "bluechips")

	c.ClearCache()
	assert.Empty(t, c.GetCacheInfo())
	_, err := c.GetBluechips(ctx)
	require.NoError(t, err)
	assert.Equal(t, int32(2), counts["leading-equity"].Load())

	uncached, counts := newFixtureClient(t, false)
	for range 3 {
		_, err := uncached.GetBluechips(ctx)
		require.NoError(t, err)
	}
	assert.Equal(t, int32(3), counts["leading-equity"].Load(), "EnableCache: false must hit the API every time")
	assert.Empty(t, uncached.GetCacheInfo())
	uncached.ClearCache()
}

func TestClientErrorsPropagate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	c := NewClient(&ClientOptions{BaseURL: srv.URL})
	defer c.Close()
	ctx := context.Background()

	calls := map[string]func() error{
		"security":   func() error { _, err := c.GetSecurity(ctx, "X"); return err },
		"bluechip":   func() error { _, err := c.GetBluechip(ctx, "X"); return err },
		"cedear":     func() error { _, err := c.GetCedear(ctx, "X"); return err },
		"galpone":    func() error { _, err := c.GetGalpone(ctx, "X"); return err },
		"bond":       func() error { _, err := c.GetBond(ctx, "X"); return err },
		"option":     func() error { _, err := c.GetOption(ctx, "X"); return err },
		"future":     func() error { _, err := c.GetFuture(ctx, "X"); return err },
		"multiple":   func() error { _, err := c.GetMultipleSecurities(ctx, []string{"X"}); return err },
		"search":     func() error { _, err := c.SearchSecurities(ctx, "X"); return err },
		"marketTime": func() error { _, err := c.GetMarketTime(ctx); return err },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			assert.ErrorContains(t, call(), "HTTP error 404")
		})
	}
}

func TestClientPartialFailures(t *testing.T) {
	// Each lookup fans out to several endpoints; a failure on a later one
	// must still surface.
	for _, failing := range []string{"cedears", "general-equity", "lebacs", "negociable-obligations"} {
		t.Run(failing, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				name := strings.TrimPrefix(r.URL.Path, "/vanoms-be-core/rest/api/bymadata/free/")
				if name == failing {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				body, _ := os.ReadFile(filepath.Join("internal/api/testdata", name+".json"))
				_, _ = w.Write(body)
			}))
			defer srv.Close()
			c := NewClient(&ClientOptions{BaseURL: srv.URL})
			defer c.Close()
			ctx := context.Background()

			if failing == "lebacs" || failing == "negociable-obligations" {
				_, err := c.GetBond(ctx, "X")
				assert.Error(t, err)
				return
			}
			_, err := c.GetSecurity(ctx, "X")
			assert.Error(t, err)
			_, err = c.GetMultipleSecurities(ctx, []string{"X"})
			assert.Error(t, err)
			_, err = c.SearchSecurities(ctx, "X")
			assert.Error(t, err)
		})
	}
}

func TestParseSettlementAlias(t *testing.T) {
	s, ok := ParseSettlement("24HS")
	assert.True(t, ok)
	assert.Equal(t, Settlement24HS, s)
}

func TestNoOpLoggerAndErrors(t *testing.T) {
	var l NoOpLogger
	l.Debug("d")
	l.Info("i")
	l.Warn("w")
	l.Error("e")

	cause := errors.New("boom")
	err := ErrTimeout.WithUnderlying(cause)
	assert.ErrorIs(t, err, cause)
	assert.False(t, IsRetryable(nil))
}
