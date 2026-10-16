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
	respData, err := c.post(c.buildURL(pathMarketTime), []byte(`{"Content-Type":"application/json"}`))
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
	respData, err := c.post(c.buildURL(pathIndexPrice), []byte(`{"Content-Type":"application/json"}`))
	if err != nil {
		return nil, err
	}

	var rawIndices []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawIndices); err != nil {
		return nil, err
	}

	indices := make([]Index, 0, len(rawIndices))
	for _, raw := range rawIndices {
		indices = append(indices, Index{
			Description:   utils.GetString(raw, "description"),
			Symbol:        utils.GetString(raw, "symbol"),
			Last:          utils.GetFloat64(raw, "price"),
			Change:        utils.GetFloat64(raw, "variation"),
			High:          utils.GetFloat64(raw, "highValue"),
			Low:           utils.GetFloat64(raw, "minValue"),
			PreviousClose: utils.GetFloat64(raw, "previousClosingPrice"),
		})
	}

	return indices, nil
}

// MarketResume retrieves market summary data
func (c *Client) MarketResume(ctx context.Context) ([]MarketSummary, error) {
	respData, err := c.post(c.buildURL(pathTotalNegotiated), []byte(`{"Content-Type":"application/json"}`))
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
