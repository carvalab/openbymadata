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
	// Always ensure "24HS" suffix is needed for the api
	symbol = symbol + " 24HS"

	// Convert time.Time to Unix timestamps
	fromUnix := from.Unix()
	toUnix := to.Unix()

	// Build URL with query parameters
	endpoint := "chart/historical-series/history"
	baseURL := c.buildURL(endpoint)

	// Add query parameters
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

	c.debugLogResponse("chart/historical-series/history", respData)

	var historyResp HistoryResponse
	if err := json.Unmarshal(respData, &historyResp); err != nil {
		return nil, fmt.Errorf("failed to parse history response: %w", err)
	}

	if historyResp.Status != "ok" {
		return nil, fmt.Errorf("no historical data available for symbol %s (status: %s)", symbol, historyResp.Status)
	}

	return &OHLCV{
		Time:   historyResp.Time,
		Open:   historyResp.Open,
		High:   historyResp.High,
		Low:    historyResp.Low,
		Close:  historyResp.Close,
		Volume: historyResp.Volume,
	}, nil
}

// GetHistoryLastDays is a convenience method to get history for the last N days
func (c *Client) GetHistoryLastDays(ctx context.Context, symbol string, days int) (*OHLCV, error) {
	// Calculate dates
	to := time.Now()
	from := to.AddDate(0, 0, -days) // Go back N days

	return c.GetHistory(ctx, symbol, "D", from, to)
}

// ConvertToHistoricalData converts OHLCV to HistoricalData array (utility function)
func (c *Client) ConvertToHistoricalData(slices *OHLCV) ([]HistoricalData, error) {
	// Validate that all arrays have the same length
	length := len(slices.Time)
	if len(slices.Close) != length || len(slices.Open) != length ||
		len(slices.High) != length || len(slices.Low) != length || len(slices.Volume) != length {
		return nil, fmt.Errorf("inconsistent array lengths in OHLCV slices")
	}

	// Convert to structured format
	data := make([]HistoricalData, length)
	for i := range length {
		data[i] = HistoricalData{
			Time:   slices.Time[i],
			Open:   slices.Open[i],
			High:   slices.High[i],
			Low:    slices.Low[i],
			Close:  slices.Close[i],
			Volume: slices.Volume[i],
		}
	}

	return data, nil
}

// OHLCV represents historical data as separate slices
type OHLCV struct {
	Time   []int64   `json:"time"`
	Open   []float64 `json:"open"`
	High   []float64 `json:"high"`
	Low    []float64 `json:"low"`
	Close  []float64 `json:"close"`
	Volume []int64   `json:"volume"`
}
