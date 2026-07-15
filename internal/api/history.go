package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// GetHistory retrieves historical price data for a given symbol as OHLCV arrays
// symbol: the ticker symbol (will always have "24HS" suffix added)
// resolution: "D" for daily, "W" for weekly, "M" for monthly
// from: Start date as time.Time
// to: End date as time.Time
func (c *Client) GetHistory(ctx context.Context, symbol, resolution string, from, to time.Time) (*OHLCV, error) {
	symbol = symbol + " 24HS"

	fromUnix := from.Unix()
	toUnix := to.Unix()

	endpoint := "chart/historical-series/history"
	baseURL := c.buildURL(endpoint)

	params := url.Values{}
	params.Add("symbol", symbol)
	params.Add("resolution", resolution)
	params.Add("from", strconv.FormatInt(fromUnix, 10))
	params.Add("to", strconv.FormatInt(toUnix, 10))

	fullURL := baseURL + "?" + params.Encode()

	respData, err := c.get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get history data: %w", err)
	}

	var historyResp HistoryResponse
	if err := json.Unmarshal(respData, &historyResp); err != nil {
		return nil, fmt.Errorf("failed to parse history response: %w", err)
	}

	if historyResp.Status != "ok" {
		return nil, fmt.Errorf("no historical data available for symbol %s (status: %s)", symbol, historyResp.Status)
	}

	return &OHLCV{
		Time:   parseDates(historyResp.Time),
		Open:   historyResp.Open,
		High:   historyResp.High,
		Low:    historyResp.Low,
		Close:  historyResp.Close,
		Volume: historyResp.Volume,
	}, nil
}

// GetHistoryLastDays is a convenience method to get history for the last N days
func (c *Client) GetHistoryLastDays(ctx context.Context, symbol string, days int) (*OHLCV, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -days)

	return c.GetHistory(ctx, symbol, "D", from, to)
}

// OHLCV represents historical data as separate slices
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
