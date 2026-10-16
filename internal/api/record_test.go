//go:build record

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRecordFixtures refreshes testdata/ from the live BYMA API using the
// client's own request bodies and headers. Run it with:
//
//	go test -tags record -run TestRecordFixtures -v ./internal/api/
//
// By default it only compares each live response with the committed fixture
// and reports fields that disappeared (test failure) or appeared (log).
// Set RECORD_WRITE=1 to also overwrite the fixtures; prices will differ, so
// update the expected values in the tests afterwards. Hand-written fixtures
// listed in testdata/README.md are skipped.
func TestRecordFixtures(t *testing.T) {
	handWritten := map[string]bool{"index-future.json": true, "bnown_bolsar_balances.json": true}

	c := New(&ClientOptions{BaseURL: "https://open.bymadata.com.ar", Timeout: 30 * time.Second, RetryAttempts: 3, Logger: nopLogger{}})
	c.httpClient.Transport = recorder{t: t, skip: handWritten, base: c.basePath()}
	ctx := context.Background()
	to := time.Now()

	calls := map[string]func() error{
		"market-time":    func() error { _, err := c.GetMarketTime(ctx); return err },
		"indices":        func() error { _, err := c.GetIndices(ctx); return err },
		"summary":        func() error { _, err := c.MarketResume(ctx); return err },
		"bluechips":      func() error { _, err := c.GetBluechips(ctx); return err },
		"galpones":       func() error { _, err := c.GetGalpones(ctx); return err },
		"cedears":        func() error { _, err := c.GetCedears(ctx); return err },
		"bonds":          func() error { _, err := c.GetBonds(ctx); return err },
		"letters":        func() error { _, err := c.GetShortTermBonds(ctx); return err },
		"corporate":      func() error { _, err := c.GetCorporateBonds(ctx); return err },
		"options":        func() error { _, err := c.GetOptions(ctx); return err },
		"futures":        func() error { _, err := c.GetFutures(ctx); return err },
		"news":           func() error { _, err := c.GetNews(ctx); return err },
		"income":         func() error { _, err := c.GetIncomeStatement(ctx, "ALUA"); return err },
		"quote":          func() error { _, err := c.GetCurrentQuote(ctx, "GGAL", Settlement24HS); return err },
		"company-info":   func() error { _, err := c.GetCompanyInfo(ctx, "ALUA"); return err },
		"equity-profile": func() error { _, err := c.GetEquityProfile(ctx, "ALUA"); return err },
		"management":     func() error { _, err := c.GetCompanyManagement(ctx, "ALUA"); return err },
		"etfs":           func() error { _, err := c.GetETFs(ctx); return err },
		"mep":            func() error { _, err := c.GetMEPIndices(ctx); return err },
		"top-movers":     func() error { _, err := c.GetTopMovers(ctx); return err },
		"bond-indices":   func() error { _, err := c.GetBondIndices(ctx); return err },
		"green":          func() error { _, err := c.GetGreenBonds(ctx); return err },
		"social":         func() error { _, err := c.GetSocialBonds(ctx); return err },
		"sustainability": func() error { _, err := c.GetSustainabilityBonds(ctx); return err },
		"linked":         func() error { _, err := c.GetLinkedBonds(ctx); return err },
		"senebi-bonds":   func() error { _, err := c.GetSenebiBonds(ctx); return err },
		"senebi-letters": func() error { _, err := c.GetSenebiLetters(ctx); return err },
		"senebi-corp":    func() error { _, err := c.GetSenebiCorporateBonds(ctx); return err },
		"coupons":        func() error { _, err := c.GetBondCoupons(ctx, "GD30"); return err },
		"responsibles":   func() error { _, err := c.GetMarketResponsibles(ctx, "ALUA"); return err },
		"search":         func() error { _, err := c.SearchInstruments(ctx, "GGAL"); return err },
		"index-history": func() error {
			_, err := c.GetIndexHistory(ctx, "M", ResolutionDaily, to.AddDate(0, 0, -30), to)
			return err
		},
		"history": func() error {
			_, err := c.GetHistory(ctx, "GGAL 24HS", ResolutionDaily, to.AddDate(0, 0, -30), to)
			return err
		},
	}
	for name, call := range calls {
		// Call failures are the integration test's job; here only shape matters.
		if err := call(); err != nil {
			t.Logf("%s: skipped: %v", name, err)
		}
	}
}

// recorder saves each successful response body, trimmed, to testdata.
type recorder struct {
	t    *testing.T
	skip map[string]bool
	base string
}

func (r recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return resp, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	name := strings.ReplaceAll(strings.TrimPrefix(req.URL.Path, r.base), "/", "_") + ".json"
	if r.skip[name] {
		return resp, nil
	}
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		r.t.Errorf("%s: not JSON: %v", name, err)
		return resp, nil
	}
	path := filepath.Join("testdata", name)
	if old, err := os.ReadFile(path); err == nil {
		var prev any
		if json.Unmarshal(old, &prev) == nil {
			reportDrift(r.t, name, keyPaths(prev, ""), keyPaths(v, ""))
		}
	}
	if os.Getenv("RECORD_WRITE") == "" {
		return resp, nil
	}
	out, _ := json.MarshalIndent(trim(v, ""), "", " ")
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		r.t.Errorf("%s: %v", name, err)
	}
	r.t.Logf("recorded %s", name)
	return resp, nil
}

// keyPaths lists every JSON key path in v, treating all list items as one
// ("data[].symbol"), so two payloads can be compared by shape only.
func keyPaths(v any, prefix string) map[string]bool {
	out := map[string]bool{}
	switch x := v.(type) {
	case []any:
		for _, item := range x {
			for k := range keyPaths(item, prefix+"[]") {
				out[k] = true
			}
		}
	case map[string]any:
		for k, val := range x {
			p := prefix + "." + k
			out[p] = true
			for kk := range keyPaths(val, p) {
				out[kk] = true
			}
		}
	}
	return out
}

// reportDrift fails on fields BYMA stopped sending (the client may rely on
// them) and logs fields it started sending (possible new features).
func reportDrift(t *testing.T, name string, before, after map[string]bool) {
	if len(after) == 0 {
		return // empty list today, nothing to compare
	}
	for k := range before {
		if !after[k] {
			t.Errorf("DRIFT %s: field gone: %s", name, k)
		}
	}
	for k := range after {
		if !before[k] {
			t.Logf("DRIFT %s: new field: %s", name, k)
		}
	}
}

// trim keeps fixtures small: three records per list, five points per
// OHLCV series (keys t/o/h/l/c/v).
func trim(v any, key string) any {
	switch x := v.(type) {
	case []any:
		n := 3
		if len(key) == 1 && strings.Contains("tohlcv", key) {
			n = 5
		}
		if len(x) > n {
			x = x[:n]
		}
		for i := range x {
			x[i] = trim(x[i], "")
		}
		return x
	case map[string]any:
		for k, val := range x {
			x[k] = trim(val, k)
		}
		return x
	}
	return v
}
