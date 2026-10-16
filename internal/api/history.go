package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// GetHistory retrieves historical OHLCV data for a given symbol and resolution.
// symbol is sent verbatim; callers that need a specific settlement window
// should include it (e.g. "GGAL 24HS"). When the API responds with status
// "no_data", an empty *OHLCV is returned without raising an error so callers
// can distinguish "no data" from transport failures.
func (c *Client) GetHistory(ctx context.Context, symbol string, resolution Resolution, from, to time.Time) (*OHLCV, error) {
	return c.fetchHistory(symbol, resolution, from, to)
}

// GetHistoryLastDays is a convenience wrapper that requests daily bars for
// the trailing N days, ending now.
func (c *Client) GetHistoryLastDays(ctx context.Context, symbol string, days int) (*OHLCV, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -days)
	return c.GetHistory(ctx, symbol, ResolutionDaily, from, to)
}

// GetIntradayHistory requests 1-minute bars by default. Use a different
// Resolution (5 or 60) for coarser aggregation.
func (c *Client) GetIntradayHistory(ctx context.Context, symbol string, resolution Resolution, from, to time.Time) (*OHLCV, error) {
	if resolution == ResolutionDaily || resolution == ResolutionWeekly || resolution == ResolutionMonthly {
		return nil, fmt.Errorf("intraday resolution must be %q, %q or %q", Resolution1Min, Resolution5Min, Resolution60Min)
	}
	return c.fetchHistory(symbol, resolution, from, to)
}

// GetIndexHistory retrieves OHLCV bars for a market index, e.g. "M" for the
// S&P Merval or "G" for the general index (see GetIndices for symbols).
// Index volume is always 0.
func (c *Client) GetIndexHistory(ctx context.Context, symbol string, resolution Resolution, from, to time.Time) (*OHLCV, error) {
	return c.fetchSeries(pathIndexHistory, symbol, resolution, from, to)
}

func (c *Client) fetchHistory(symbol string, resolution Resolution, from, to time.Time) (*OHLCV, error) {
	return c.fetchSeries(pathHistorical, symbol, resolution, from, to)
}

func (c *Client) fetchSeries(endpoint, symbol string, resolution Resolution, from, to time.Time) (*OHLCV, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("resolution", string(resolution))
	params.Set("from", strconv.FormatInt(from.Unix(), 10))
	params.Set("to", strconv.FormatInt(to.Unix(), 10))

	respData, err := c.get(c.buildURL(endpoint) + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to get history data: %w", err)
	}

	var resp HistoryResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse history response: %w", err)
	}

	if resp.Status != "ok" || len(resp.Time) == 0 {
		return &OHLCV{}, nil
	}

	return &OHLCV{
		Time:   parseDates(resp.Time),
		Open:   resp.Open,
		High:   resp.High,
		Low:    resp.Low,
		Close:  resp.Close,
		Volume: resp.Volume,
	}, nil
}

// OHLCV represents historical data as separate slices.
type OHLCV struct {
	Time   []time.Time `json:"time"`
	Open   []float64   `json:"open"`
	High   []float64   `json:"high"`
	Low    []float64   `json:"low"`
	Close  []float64   `json:"close"`
	Volume []int64     `json:"volume"`
}

func parseDates(unixTimes []int64) []time.Time {
	res := make([]time.Time, len(unixTimes))
	for i, t := range unixTimes {
		res[i] = time.Unix(t, 0)
	}
	return res
}
