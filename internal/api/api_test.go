package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const freePrefix = "/vanoms-be-core/rest/api/bymadata/free/"

type nopLogger struct{}

func (nopLogger) Debug(string, ...LogField) {}
func (nopLogger) Info(string, ...LogField)  {}
func (nopLogger) Warn(string, ...LogField)  {}
func (nopLogger) Error(string, ...LogField) {}

// fixtureServer replays testdata/<endpoint>.json for each request, where
// <endpoint> is the path after the free API prefix with "/" replaced by "_".
func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.ReplaceAll(strings.TrimPrefix(r.URL.Path, freePrefix), "/", "_") + ".json"
		body, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// fixtureServerWith answers every request with body.
func fixtureServerWith(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// fixtureServerStatus answers every request with an empty response and code.
func fixtureServerStatus(t *testing.T, code int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newTestClient(baseURL string) *Client {
	return New(&ClientOptions{BaseURL: baseURL, Timeout: 5 * time.Second, Logger: nopLogger{}})
}

func TestSecurities(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	bluechips, err := c.GetBluechips(ctx)
	require.NoError(t, err)
	require.Len(t, bluechips, 3)
	assert.Equal(t, "ALUA", bluechips[0].Symbol)
	assert.Equal(t, 846.0, bluechips[0].Last)
	assert.Equal(t, 846.5, bluechips[0].Ask)
	assert.Equal(t, int64(328459), bluechips[0].Volume)
	assert.Equal(t, "CS", bluechips[0].Group)

	galpones, err := c.GetGalpones(ctx)
	require.NoError(t, err)
	assert.Equal(t, "A3", galpones[0].Symbol)

	// CEDEARs come back as a bare array instead of the {data: [...]} envelope.
	cedears, err := c.GetCedears(ctx)
	require.NoError(t, err)
	assert.Equal(t, "AAL", cedears[0].Symbol)
	assert.Equal(t, int64(3871), cedears[0].Volume)
}

func TestFixedIncome(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	bonds, err := c.GetBonds(ctx)
	require.NoError(t, err)
	assert.Equal(t, "AE38", bonds[0].Symbol)
	assert.Equal(t, time.Date(2038, 1, 9, 0, 0, 0, 0, time.UTC), bonds[0].Expiration)

	letters, err := c.GetShortTermBonds(ctx)
	require.NoError(t, err)
	assert.Equal(t, "B2D6C", letters[0].Symbol)

	corp, err := c.GetCorporateBonds(ctx)
	require.NoError(t, err)
	assert.Equal(t, "A11LD", corp[0].Symbol)
	assert.Equal(t, "CORP", corp[0].Group)
}

func TestDerivatives(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	options, err := c.GetOptions(ctx)
	require.NoError(t, err)
	assert.Equal(t, "ALUC1000DI", options[0].Symbol)
	assert.Equal(t, "ALUA", options[0].UnderlyingAsset)
	assert.Equal(t, time.Date(2026, 12, 18, 0, 0, 0, 0, time.UTC), options[0].Expiration)

	// Futures prices are quoted per 1000 and scaled up by the client.
	futures, err := c.GetFutures(ctx)
	require.NoError(t, err)
	require.Len(t, futures, 1)
	assert.Equal(t, "DLR/NOV26", futures[0].Symbol)
	assert.InDelta(t, 1455.0, futures[0].Last, 1e-9)
	assert.Equal(t, int64(12000), futures[0].Volume)
	assert.Equal(t, int64(340), futures[0].OpenInterest)
}

func TestMarket(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	indices, err := c.GetIndices(ctx)
	require.NoError(t, err)
	assert.Equal(t, "G", indices[0].Symbol)
	assert.Equal(t, 0.0016, indices[0].Change)

	summary, err := c.MarketResume(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Total Negociado BYMA", summary[0].Symbol)
	assert.Equal(t, 11325569533965.3, summary[0].TotalNegotiated)

	mt, err := c.GetMarketTime(ctx)
	require.NoError(t, err)
	assert.True(t, mt.IsWorkingDay)
	_, offset := mt.Opens.Zone()
	assert.Equal(t, -3*60*60, offset)
	assert.Equal(t, 10, mt.Opens.Hour())
	assert.Equal(t, 30, mt.Opens.Minute())
	assert.Equal(t, 17, mt.Closes.Hour())
	assert.Equal(t, time.Now().In(mt.Opens.Location()).YearDay(), mt.Opens.YearDay())

	working, err := c.IsWorkingDay(ctx)
	require.NoError(t, err)
	assert.True(t, working)
}

func TestQuote(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	q, err := c.GetCurrentQuote(ctx, "GGAL", Settlement24HS)
	require.NoError(t, err)
	assert.Equal(t, "GGAL", q.Symbol)
	assert.Equal(t, 5830.0, q.Bid)
	assert.Equal(t, 5840.0, q.Ask)
	assert.Equal(t, int64(2966), q.Operations)

	_, err = c.GetCurrentQuote(ctx, "GGAL", Settlement("bogus"))
	assert.ErrorContains(t, err, "invalid settlement")
}

func TestNewsAndStatements(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	news, err := c.GetNews(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Gobierno Nacional", news[0].Titulo)
	assert.Equal(t, downloadBase+"502125", news[0].Descarga)
	assert.Equal(t, time.Date(2026, 10, 16, 9, 57, 0, 0, time.UTC), news[0].Fecha)

	statements, err := c.GetIncomeStatement(ctx, "ALUA")
	require.NoError(t, err)
	require.Len(t, statements, 1)
	assert.Equal(t, "ALUA", statements[0].Symbol)
	assert.Equal(t, "2026-06-30", statements[0].FechaCierre)
	assert.Equal(t, "2026-09-08", statements[0].FechaPresentacion)
	assert.Equal(t, downloadBase+"500007", statements[0].BalancesArchivo)
}

func TestCompany(t *testing.T) {
	c := newTestClient(fixtureServer(t).URL)
	ctx := context.Background()

	info, err := c.GetCompanyInfo(ctx, "ALUA")
	require.NoError(t, err)
	assert.Equal(t, "www.aluar.com.ar", info.Fields["website"])

	profile, err := c.GetEquityProfile(ctx, "ALUA")
	require.NoError(t, err)
	assert.Equal(t, "ARALUA010258", profile.Fields["codigoIsin"])
	assert.Equal(t, "1", profile.Fields["valorNominal"])

	directors, err := c.GetCompanyManagement(ctx, "ALUA")
	require.NoError(t, err)
	require.Len(t, directors, 3)
	assert.Equal(t, "LEVINAS MARTIN JOSE", directors[0].Name)
	assert.Equal(t, "Presidente", directors[0].Position)

	balance, err := c.GetCompanyBalance(ctx, "ALUA")
	require.NoError(t, err)
	require.Len(t, balance, 2)
	assert.Equal(t, "Activo Total", balance[0].Account)
	assert.Equal(t, "1.2345675e+06", balance[0].Values["Valor"])
	assert.Equal(t, "890000", balance[1].Values["Valor"])
}

func TestCompanyEmptyData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c := newTestClient(srv.URL)
	ctx := context.Background()

	info, err := c.GetCompanyInfo(ctx, "XXXX")
	require.NoError(t, err)
	assert.Equal(t, "XXXX", info.Symbol)
	assert.Nil(t, info.Fields)

	profile, err := c.GetEquityProfile(ctx, "XXXX")
	require.NoError(t, err)
	assert.Nil(t, profile.Fields)

	balance, err := c.GetCompanyBalance(ctx, "XXXX")
	require.NoError(t, err)
	assert.Empty(t, balance)

	_, err = c.GetCurrentQuote(ctx, "XXXX", SettlementCI)
	assert.ErrorContains(t, err, "no quote available")
}

func TestHistory(t *testing.T) {
	var gotQuery atomic.Value
	fixtures := fixtureServer(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery.Store(r.URL.RawQuery)
		fixtures.Config.Handler.ServeHTTP(w, r)
	}))
	defer srv.Close()
	c := newTestClient(srv.URL)
	ctx := context.Background()

	from := time.Unix(1789000000, 0)
	to := time.Unix(1790100000, 0)
	h, err := c.GetHistory(ctx, "GGAL 24HS", ResolutionDaily, from, to)
	require.NoError(t, err)
	require.Len(t, h.Time, 5)
	assert.Equal(t, int64(1789527600), h.Time[0].Unix())
	assert.Equal(t, 6760.0, h.Close[0])
	assert.Equal(t, int64(3240323), h.Volume[4])
	assert.Equal(t, "from=1789000000&resolution=D&symbol=GGAL+24HS&to=1790100000", gotQuery.Load())

	last, err := c.GetHistoryLastDays(ctx, "GGAL 24HS", 30)
	require.NoError(t, err)
	assert.Len(t, last.Time, 5)

	bars, err := c.GetIntradayHistory(ctx, "GGAL 24HS", Resolution5Min, from, to)
	require.NoError(t, err)
	assert.Len(t, bars.Time, 5)

	_, err = c.GetIntradayHistory(ctx, "GGAL 24HS", ResolutionDaily, from, to)
	assert.ErrorContains(t, err, "intraday resolution")
}

func TestHistoryNoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"s":"no_data"}`))
	}))
	defer srv.Close()

	h, err := newTestClient(srv.URL).GetHistory(context.Background(), "GGAL", ResolutionDaily, time.Now().AddDate(0, 0, -1), time.Now())
	require.NoError(t, err)
	assert.Empty(t, h.Time)
}

func TestInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()
	c := newTestClient(srv.URL)
	ctx := context.Background()

	calls := map[string]func() error{
		"securities": func() error { _, err := c.GetBluechips(ctx); return err },
		"cedears":    func() error { _, err := c.GetCedears(ctx); return err },
		"bonds":      func() error { _, err := c.GetBonds(ctx); return err },
		"corporate":  func() error { _, err := c.GetCorporateBonds(ctx); return err },
		"options":    func() error { _, err := c.GetOptions(ctx); return err },
		"futures":    func() error { _, err := c.GetFutures(ctx); return err },
		"indices":    func() error { _, err := c.GetIndices(ctx); return err },
		"summary":    func() error { _, err := c.MarketResume(ctx); return err },
		"news":       func() error { _, err := c.GetNews(ctx); return err },
		"income":     func() error { _, err := c.GetIncomeStatement(ctx, "X"); return err },
		"marketTime": func() error { _, err := c.GetMarketTime(ctx); return err },
		"workingDay": func() error { _, err := c.IsWorkingDay(ctx); return err },
		"quote":      func() error { _, err := c.GetCurrentQuote(ctx, "X", SettlementCI); return err },
		"info":       func() error { _, err := c.GetCompanyInfo(ctx, "X"); return err },
		"profile":    func() error { _, err := c.GetEquityProfile(ctx, "X"); return err },
		"management": func() error { _, err := c.GetCompanyManagement(ctx, "X"); return err },
		"balance":    func() error { _, err := c.GetCompanyBalance(ctx, "X"); return err },
		"history": func() error {
			_, err := c.GetHistory(ctx, "X", ResolutionDaily, time.Now(), time.Now())
			return err
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, call())
		})
	}
}

func TestEnvelopeWithoutData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"message":"nope"}`))
	}))
	defer srv.Close()

	_, err := newTestClient(srv.URL).GetBluechips(context.Background())
	assert.ErrorContains(t, err, "no data in response")
}

func TestRetryPolicy(t *testing.T) {
	tests := []struct {
		status    int
		wantCalls int32
	}{
		{http.StatusBadRequest, 2},
		{http.StatusNotFound, 1},
		{http.StatusUnauthorized, 1},
		{http.StatusServiceUnavailable, 2},
		{http.StatusTooManyRequests, 2},
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			c := New(&ClientOptions{BaseURL: srv.URL, Timeout: time.Second, RetryAttempts: 1, Logger: nopLogger{}})
			_, err := c.GetIndices(context.Background())
			require.Error(t, err)
			assert.Contains(t, err.Error(), "HTTP error")
			assert.Equal(t, tt.wantCalls, calls.Load())
		})
	}
}

func TestParseSettlement(t *testing.T) {
	tests := []struct {
		in   string
		want Settlement
		ok   bool
	}{
		{"1", SettlementCI, true},
		{"ci", SettlementCI, true},
		{"2", Settlement24HS, true},
		{"24hs", Settlement24HS, true},
		{"3", Settlement48HS, true},
		{"48HS", Settlement48HS, true},
		{"T+5", "", false},
	}
	for _, tt := range tests {
		got, ok := ParseSettlement(tt.in)
		assert.Equal(t, tt.want, got, tt.in)
		assert.Equal(t, tt.ok, ok, tt.in)
	}
}

func TestResolveLocation(t *testing.T) {
	tests := []struct {
		in         string
		wantOffset int
	}{
		{"GMT-03:00", -3 * 3600},
		{"UTC+05:30", 5*3600 + 30*60},
		{"-0300", -3 * 3600},
		{"GMT", 0},
		{"America/Argentina/Buenos_Aires", -3 * 3600},
		{"", -3 * 3600},
		{"Not/AZone", -3 * 3600},
		{"GMT+bogus", -3 * 3600},
	}
	ref := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	for _, tt := range tests {
		_, offset := ref.In(resolveLocation(tt.in)).Zone()
		assert.Equal(t, tt.wantOffset, offset, tt.in)
	}
}

func TestParseClockTime(t *testing.T) {
	day := time.Date(2026, 10, 16, 0, 0, 0, 0, buenosAires)
	assert.Equal(t, time.Date(2026, 10, 16, 11, 0, 0, 0, buenosAires), parseClockTime("11:00", day))
	assert.Equal(t, time.Date(2026, 10, 16, 17, 0, 5, 0, buenosAires), parseClockTime("17:00:05", day))
	assert.True(t, parseClockTime("", day).IsZero())
	assert.True(t, parseClockTime("noon", day).IsZero())
}

func TestStringifyValue(t *testing.T) {
	assert.Equal(t, "", stringifyValue(nil))
	assert.Equal(t, "x", stringifyValue("x"))
	assert.Equal(t, "42", stringifyValue(42.0))
	assert.Equal(t, "1.5", stringifyValue(1.5))
	assert.Equal(t, "true", stringifyValue(true))
	assert.Equal(t, "false", stringifyValue(false))
	assert.Equal(t, "[a b]", stringifyValue([]interface{}{"a", "b"}))
}

func TestDownloadURL(t *testing.T) {
	assert.Equal(t, downloadBase+"7", downloadURL(map[string]interface{}{"id": 7.0}, "id"))
	assert.Equal(t, downloadBase+"abc", downloadURL(map[string]interface{}{"id": "abc"}, "id"))
	assert.Equal(t, "", downloadURL(map[string]interface{}{}, "id"))
	assert.Equal(t, "2026", dateOnly("2026"))
}

func TestClose(t *testing.T) {
	c := newTestClient("http://example.invalid")
	c.Close()
	c.httpClient.Transport = &http.Transport{}
	c.Close()
}

func TestSectionHeader(t *testing.T) {
	got := map[string]string{}
	var mu sync.Mutex
	fixtures := fixtureServer(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		got[strings.TrimPrefix(r.URL.Path, "/base"+freePrefix)] = r.Header.Get("Options")
		mu.Unlock()
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/base")
		fixtures.Config.Handler.ServeHTTP(w, r)
	}))
	defer srv.Close()

	// A base URL with its own path must still resolve the endpoint name.
	c := newTestClient(srv.URL + "/base")
	ctx := context.Background()
	_, _ = c.GetNews(ctx)
	_, _ = c.GetMarketTime(ctx)
	_, _ = c.GetBonds(ctx)
	_, _ = c.GetCurrentQuote(ctx, "GGAL", Settlement24HS)
	_, _ = c.GetHistory(ctx, "GGAL 24HS", ResolutionDaily, time.Now(), time.Now())

	assert.Equal(t, map[string]string{
		pathNews:         "notices-and-events",
		pathMarketTime:   "dashboard",
		pathPublicBonds:  "renta-fija",
		pathCurrentQuote: "technical-details",
		pathHistorical:   "renta-variable",
	}, got)
	for path := range sectionHeader {
		assert.NotEmpty(t, sectionHeader[path], path)
	}
}

func TestAttemptCount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := New(&ClientOptions{BaseURL: srv.URL, Timeout: time.Second, RetryAttempts: 3, Logger: nopLogger{}})
	_, err := c.GetNews(context.Background())
	assert.EqualError(t, err, "request failed after 1 attempts: HTTP error 404")
}

func TestContextCancelsRetries(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := New(&ClientOptions{BaseURL: srv.URL, Timeout: 5 * time.Second, RetryAttempts: 5, Logger: nopLogger{}})
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := c.GetIndices(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 900*time.Millisecond, "must not wait out the 1s retry backoff")
	assert.Equal(t, int32(1), calls.Load())
}

func TestContextAbortsInFlightRequest(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)

	c := New(&ClientOptions{BaseURL: srv.URL, Timeout: 10 * time.Second, RetryAttempts: 3, Logger: nopLogger{}})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := c.IsMarketOpen(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 2*time.Second)
}
