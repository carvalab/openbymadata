package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/carvalab/openbymadata/internal/utils"
)

// GetBonds retrieves government bonds (public bonds)
func (c *Client) GetBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, "public-bonds")
}

// GetShortTermBonds retrieves short-term government bonds (LEBACs)
func (c *Client) GetShortTermBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, "lebacs")
}

// GetCorporateBonds retrieves corporate bonds (negotiable obligations)
func (c *Client) GetCorporateBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, "negociable-obligations")
}

// getFixedIncome is a helper function to retrieve bonds from different endpoints
func (c *Client) getFixedIncome(ctx context.Context, endpoint string) ([]Bond, error) {
	// Use standard payload for fixed income data
	data := []byte(`{"excludeZeroPxAndQty":false,"T2":false,"T1":true,"T0":false,"Content-Type":"application/json"}`)
	url := c.buildURL(endpoint)

	respData, err := c.post(url, data)
	if err != nil {
		return nil, err
	}

	// Debug: log raw response
	c.debugLogResponse(endpoint, respData)

	var rawBonds []map[string]interface{}
	// public-bonds and lebacs return data wrapped, negociable-obligations return direct
	if endpoint == "public-bonds" || endpoint == "lebacs" {
		if err := c.parseAPIResponse(respData, &rawBonds); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(respData, &rawBonds); err != nil {
			return nil, fmt.Errorf("invalid response: %w", err)
		}
	}

	bonds := make([]Bond, 0, len(rawBonds))
	for _, raw := range rawBonds {
		bond := Bond{
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
			Expiration:    utils.GetTime(raw, "maturityDate"),
		}
		bonds = append(bonds, bond)
	}

	return bonds, nil
}
