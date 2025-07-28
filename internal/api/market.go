package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/carvalab/openbymadata/internal/utils"
)

// IsWorkingDay checks if the current day is a working day for the BYMA market
func (c *Client) IsWorkingDay(ctx context.Context) (bool, error) {
	url := c.buildURL("market-time")
	respData, err := c.post(url, []byte(`{"Content-Type":"application/json"}`))
	if err != nil {
		return false, err
	}

	// Debug: log raw response
	c.debugLogResponse("market-time", respData)

	var response MarketTimeResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		// If we can't parse the specific response, try to check if we got data back
		// The presence of data usually indicates it's a working day
		var rawData interface{}
		if json.Unmarshal(respData, &rawData) == nil {
			return true, nil
		}
		return false, fmt.Errorf("invalid response: %w", err)
	}

	return response.IsWorkingDay, nil
}

// GetIndices retrieves market indices information
func (c *Client) GetIndices(ctx context.Context) ([]Index, error) {
	url := c.buildURL("index-price")
	respData, err := c.post(url, []byte(`{"Content-Type":"application/json"}`))
	if err != nil {
		return nil, err
	}

	// Debug: log raw response
	c.debugLogResponse("index-price", respData)

	var rawIndices []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawIndices); err != nil {
		return nil, err
	}

	indices := make([]Index, 0, len(rawIndices))
	for _, raw := range rawIndices {
		index := Index{
			Description:   c.applyDictionary(utils.GetString(raw, "description")),
			Symbol:        utils.GetString(raw, "symbol"),
			Last:          utils.GetFloat64(raw, "price"),
			Change:        utils.GetFloat64(raw, "variation"),
			High:          utils.GetFloat64(raw, "highValue"),
			Low:           utils.GetFloat64(raw, "minValue"),
			PreviousClose: utils.GetFloat64(raw, "previousClosingPrice"),
		}
		indices = append(indices, index)
	}

	return indices, nil
}

// MarketResume retrieves market summary data
func (c *Client) MarketResume(ctx context.Context) ([]MarketSummary, error) {
	url := c.buildURL("total-negotiated")
	respData, err := c.post(url, []byte(`{"Content-Type":"application/json"}`))
	if err != nil {
		return nil, err
	}

	// Debug: log raw response
	c.debugLogResponse("total-negotiated", respData)

	var rawSummaries []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawSummaries); err != nil {
		return nil, err
	}

	summaries := make([]MarketSummary, 0, len(rawSummaries))
	for _, raw := range rawSummaries {
		summary := MarketSummary{
			Symbol:          utils.GetString(raw, "symbol"),
			AssetType:       utils.GetString(raw, "assetType"),
			ParentKey:       utils.GetString(raw, "parentKey"),
			TotalNegotiated: utils.GetFloat64(raw, "totalNegotiated"),
			Volume:          utils.GetInt64(raw, "volume"),
			Operations:      utils.GetInt64(raw, "operations"),
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}
