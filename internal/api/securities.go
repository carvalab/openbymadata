package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/carvalab/openbymadata/internal/utils"
)

// GetBluechips retrieves leading equity securities (blue chip stocks)
func (c *Client) GetBluechips(ctx context.Context) ([]Security, error) {
	return c.getSecurities(ctx, "leading-equity")
}

// GetGalpones retrieves general equity securities
func (c *Client) GetGalpones(ctx context.Context) ([]Security, error) {
	return c.getSecurities(ctx, "general-equity")
}

// GetCedears retrieves CEDEAR securities
func (c *Client) GetCedears(ctx context.Context) ([]Security, error) {
	return c.getSecurities(ctx, "cedears")
}

// getSecurities is a helper function to retrieve securities from different endpoints
func (c *Client) getSecurities(ctx context.Context, endpoint string) ([]Security, error) {
	// Use payload: excludeZeroPxAndQty=false, T1=true, others false
	data := []byte(`{"excludeZeroPxAndQty":false,"T2":false,"T1":true,"T0":false,"Content-Type":"application/json"}`)
	url := c.buildURL(endpoint)

	respData, err := c.post(url, data)
	if err != nil {
		return nil, err
	}

	// Debug: log raw response
	c.debugLogResponse(endpoint, respData)

	var rawSecurities []map[string]interface{}
	// CEDEARs return data directly, others return wrapped in 'data'
	if endpoint == "cedears" {
		if err := json.Unmarshal(respData, &rawSecurities); err != nil {
			return nil, fmt.Errorf("invalid response: %w", err)
		}
	} else {
		if err := c.parseAPIResponse(respData, &rawSecurities); err != nil {
			return nil, err
		}
	}

	securities := make([]Security, 0, len(rawSecurities))
	for _, raw := range rawSecurities {
		security := Security{
			Symbol:        utils.GetString(raw, "symbol"),
			Settlement:    utils.GetString(raw, "settlementType"),
			BidSize:       utils.GetInt64(raw, "quantityBid"),
			Bid:           utils.GetFloat64(raw, "bidPrice"),
			Ask:           utils.GetFloat64(raw, "offerPrice"),
			AskSize:       utils.GetInt64(raw, "quantityOffer"),
			Last:          utils.GetFloat64(raw, "settlementPrice"),
			Close:         utils.GetFloat64(raw, "closingPrice"),
			Change:        utils.GetFloat64(raw, "imbalance"),
			Open:          utils.GetFloat64(raw, "openingPrice"),
			High:          utils.GetFloat64(raw, "tradingHighPrice"),
			Low:           utils.GetFloat64(raw, "tradingLowPrice"),
			PreviousClose: utils.GetFloat64(raw, "previousClosingPrice"),
			Turnover:      utils.GetFloat64(raw, "volumeAmount"),
			Volume:        utils.GetInt64(raw, "volume"),
			Operations:    utils.GetInt64(raw, "numberOfOrders"),
			DateTime:      utils.ParseTradeTime(utils.GetString(raw, "tradeHour")),
			Group:         utils.GetString(raw, "securityType"),
		}
		securities = append(securities, security)
	}

	return securities, nil
}
