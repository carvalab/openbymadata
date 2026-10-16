package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/carvalab/openbymadata/internal/utils"
)

// GetMarketTime returns the full market-time payload, with the opening
// and closing instants parsed in the timezone advertised by the response.
func (c *Client) GetMarketTime(ctx context.Context) (MarketTime, error) {
	respData, err := c.post(ctx, c.buildURL(pathMarketTime), []byte(`{"Content-Type":"application/json"}`))
	if err != nil {
		return MarketTime{}, err
	}

	var raw struct {
		IsWorkingDay      bool   `json:"isWorkingDay"`
		MarketOpeningTime string `json:"marketOpeningTime"`
		MarketClosingTime string `json:"marketClosingTime"`
		Timezone          string `json:"timezone"`
	}
	if err := json.Unmarshal(respData, &raw); err != nil {
		return MarketTime{}, fmt.Errorf("invalid market-time response: %w", err)
	}

	mt := MarketTime{
		IsWorkingDay: raw.IsWorkingDay,
		Timezone:     raw.Timezone,
		OpeningText:  raw.MarketOpeningTime,
		ClosingText:  raw.MarketClosingTime,
	}
	loc := resolveLocation(raw.Timezone)
	today := time.Now().In(loc)
	mt.Opens = parseClockTime(raw.MarketOpeningTime, today)
	mt.Closes = parseClockTime(raw.MarketClosingTime, today)
	return mt, nil
}

// IsWorkingDay reports whether the BYMA market is open today.
func (c *Client) IsWorkingDay(ctx context.Context) (bool, error) {
	mt, err := c.GetMarketTime(ctx)
	if err != nil {
		var rawData interface{}
		if json.Unmarshal([]byte("{}"), &rawData) == nil {
			return false, err
		}
		return false, err
	}
	return mt.IsWorkingDay, nil
}

// GetIndices retrieves market indices information
func (c *Client) GetIndices(ctx context.Context) ([]Index, error) {
	respData, err := c.post(ctx, c.buildURL(pathIndexPrice), []byte(`{"Content-Type":"application/json"}`))
	if err != nil {
		return nil, err
	}

	var rawIndices []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawIndices); err != nil {
		return nil, err
	}

	indices := make([]Index, 0, len(rawIndices))
	for _, raw := range rawIndices {
		indices = append(indices, parseIndex(raw))
	}

	return indices, nil
}

// IsMarketOpen reports whether BYMA is trading right now. It is false on
// weekends, holidays and outside trading hours; IsWorkingDay only covers
// the day.
func (c *Client) IsMarketOpen(ctx context.Context) (bool, error) {
	respData, err := c.get(ctx, c.buildURL(pathMarketOpen))
	if err != nil {
		return false, err
	}
	var open bool
	if err := json.Unmarshal(respData, &open); err != nil {
		return false, fmt.Errorf("invalid market-open response: %w", err)
	}
	return open, nil
}

// GetServerTime returns BYMA's server clock. The API sends local Buenos
// Aires time without a zone, so the result is in UTC-3.
func (c *Client) GetServerTime(ctx context.Context) (time.Time, error) {
	respData, err := c.get(ctx, c.buildURL(pathServerTime))
	if err != nil {
		return time.Time{}, err
	}
	var raw struct {
		DateTime string `json:"datetime"`
	}
	if err := json.Unmarshal(respData, &raw); err != nil {
		return time.Time{}, fmt.Errorf("invalid server-time response: %w", err)
	}
	t, err := time.ParseInLocation("2006-01-02T15:04:05.999999999", raw.DateTime, buenosAires)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid server-time %q: %w", raw.DateTime, err)
	}
	return t, nil
}

// GetMEPIndices retrieves BYMA's MEP dollar indices (IDXDOLARB and
// related), the implied ARS/USD rate from bond trading.
func (c *Client) GetMEPIndices(ctx context.Context) ([]Index, error) {
	respData, err := c.post(ctx, c.buildURL(pathMEP), []byte(`{"Content-Type":"application/json"}`))
	if err != nil {
		return nil, err
	}
	var raws []map[string]interface{}
	if err := c.parseAPIResponse(respData, &raws); err != nil {
		return nil, err
	}
	out := make([]Index, 0, len(raws))
	for _, raw := range raws {
		out = append(out, parseIndex(raw))
	}
	return out, nil
}

func parseIndex(raw map[string]interface{}) Index {
	return Index{
		Description:   utils.GetString(raw, "description"),
		Symbol:        utils.GetString(raw, "symbol"),
		Last:          utils.GetFloat64(raw, "price"),
		Change:        utils.GetFloat64(raw, "variation"),
		High:          utils.GetFloat64(raw, "highValue"),
		Low:           utils.GetFloat64(raw, "minValue"),
		PreviousClose: utils.GetFloat64(raw, "previousClosingPrice"),
	}
}

// GetTopMovers retrieves the stocks, government bonds and options from
// BYMA's "main ups and downs" panel.
func (c *Client) GetTopMovers(ctx context.Context) (*TopMovers, error) {
	respData, err := c.post(ctx, c.buildURL(pathTopMovers), []byte(`{"Content-Type":"application/json"}`))
	if err != nil {
		return nil, err
	}
	var raw struct {
		Stocks  []map[string]interface{} `json:"actions"`
		Bonds   []map[string]interface{} `json:"public-bonds"`
		Options []map[string]interface{} `json:"options"`
	}
	if err := json.Unmarshal(respData, &raw); err != nil {
		return nil, fmt.Errorf("invalid top movers response: %w", err)
	}
	out := &TopMovers{
		Stocks:  make([]Security, 0, len(raw.Stocks)),
		Bonds:   make([]Bond, 0, len(raw.Bonds)),
		Options: make([]Option, 0, len(raw.Options)),
	}
	for _, r := range raw.Stocks {
		out.Stocks = append(out.Stocks, parseSecurity(r))
	}
	for _, r := range raw.Bonds {
		out.Bonds = append(out.Bonds, parseBond(r))
	}
	for _, r := range raw.Options {
		out.Options = append(out.Options, parseOption(r))
	}
	return out, nil
}

// MarketResume retrieves market summary data
func (c *Client) MarketResume(ctx context.Context) ([]MarketSummary, error) {
	respData, err := c.post(ctx, c.buildURL(pathTotalNegotiated), []byte(`{"Content-Type":"application/json"}`))
	if err != nil {
		return nil, err
	}

	var rawSummaries []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawSummaries); err != nil {
		return nil, err
	}

	summaries := make([]MarketSummary, 0, len(rawSummaries))
	for _, raw := range rawSummaries {
		summaries = append(summaries, MarketSummary{
			Symbol:          utils.GetString(raw, "symbol"),
			AssetType:       utils.GetString(raw, "assetType"),
			ParentKey:       utils.GetString(raw, "parentKey"),
			TotalNegotiated: totalNegotiated(raw),
			Volume:          utils.GetInt64(raw, "volume"),
			Operations:      utils.GetInt64(raw, "operations"),
		})
	}

	return summaries, nil
}

// totalNegotiated reads the traded amount, which BYMA now sends as
// volumeAmount instead of totalNegotiated.
func totalNegotiated(raw map[string]interface{}) float64 {
	if v := utils.GetFloat64(raw, "totalNegotiated"); v != 0 {
		return v
	}
	return utils.GetFloat64(raw, "volumeAmount")
}

// buenosAires is used when the API timezone can't be resolved. BYMA has
// traded at UTC-3 without DST since 2009.
var buenosAires = time.FixedZone("ART", -3*60*60)

// resolveLocation maps BYMA's timezone field to a *time.Location. The API
// sends offsets such as "GMT-03:00", which time.LoadLocation rejects, so
// those are turned into a fixed zone. IANA names are also accepted.
func resolveLocation(tz string) *time.Location {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return buenosAires
	}
	offset := strings.TrimPrefix(strings.TrimPrefix(tz, "GMT"), "UTC")
	if offset != tz || strings.HasPrefix(offset, "+") || strings.HasPrefix(offset, "-") {
		if offset == "" {
			return time.UTC
		}
		if t, err := time.Parse("-07:00", offset); err == nil {
			_, secs := t.Zone()
			return time.FixedZone(tz, secs)
		}
		if t, err := time.Parse("-0700", offset); err == nil {
			_, secs := t.Zone()
			return time.FixedZone(tz, secs)
		}
		return buenosAires
	}
	if loc, err := time.LoadLocation(tz); err == nil {
		return loc
	}
	return buenosAires
}

// parseClockTime parses BYMA's opening/closing time strings (e.g. "11:00:00")
// as a wall-clock instant on day's date, in day's location.
func parseClockTime(value string, day time.Time) time.Time {
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{"15:04:05", "15:04"} {
		if t, err := time.Parse(layout, value); err == nil {
			return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), t.Second(), 0, day.Location())
		}
	}
	return time.Time{}
}
