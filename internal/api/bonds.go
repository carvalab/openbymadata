package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/carvalab/openbymadata/internal/utils"
)

// GetBonds retrieves government bonds (public bonds)
func (c *Client) GetBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathPublicBonds)
}

// GetShortTermBonds retrieves short-term government bills (Letras). The
// upstream endpoint is still named "lebacs".
func (c *Client) GetShortTermBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathLetters)
}

// GetCorporateBonds retrieves corporate bonds (negotiable obligations)
func (c *Client) GetCorporateBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathCorpBonds)
}

func (c *Client) getFixedIncome(ctx context.Context, endpoint string) ([]Bond, error) {
	data := []byte(`{"excludeZeroPxAndQty":false,"T2":false,"T1":true,"T0":false,"Content-Type":"application/json"}`)
	respData, err := c.post(c.buildURL(endpoint), data)
	if err != nil {
		return nil, err
	}

	var rawBonds []map[string]interface{}
	if endpoint == pathPublicBonds || endpoint == pathLetters {
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
		bonds = append(bonds, Bond{
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
		})
	}

	return bonds, nil
}
