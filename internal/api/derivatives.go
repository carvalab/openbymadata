package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/carvalab/openbymadata/internal/utils"
)

// GetOptions retrieves options contracts
func (c *Client) GetOptions(ctx context.Context) ([]Option, error) {
	data := []byte(`{"Content-Type":"application/json"}`)
	url := c.buildURL("options")

	respData, err := c.post(url, data)
	if err != nil {
		return nil, err
	}

	// Debug: log raw response
	c.debugLogResponse("options", respData)

	var rawOptions []map[string]interface{}
	if err := json.Unmarshal(respData, &rawOptions); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	options := make([]Option, 0, len(rawOptions))
	for _, raw := range rawOptions {
		option := Option{
			Symbol:          utils.GetString(raw, "symbol"),
			BidSize:         utils.GetInt64(raw, "quantityBid"),
			Bid:             utils.GetFloat64(raw, "bidPrice"),
			Ask:             utils.GetFloat64(raw, "offerPrice"),
			AskSize:         utils.GetInt64(raw, "quantityOffer"),
			Last:            utils.GetFloat64(raw, "settlementPrice"),
			Close:           utils.GetFloat64(raw, "closingPrice"),
			Change:          utils.GetFloat64(raw, "imbalance"),
			Open:            utils.GetFloat64(raw, "openingPrice"),
			High:            utils.GetFloat64(raw, "tradingHighPrice"),
			Low:             utils.GetFloat64(raw, "tradingLowPrice"),
			PreviousClose:   utils.GetFloat64(raw, "previousClosingPrice"),
			Turnover:        utils.GetFloat64(raw, "volumeAmount"),
			Volume:          utils.GetInt64(raw, "volume"),
			Operations:      utils.GetInt64(raw, "numberOfOrders"),
			DateTime:        utils.ParseTradeTime(utils.GetString(raw, "tradeHour")),
			UnderlyingAsset: utils.GetString(raw, "underlyingSymbol"),
			Expiration:      utils.GetTime(raw, "maturityDate"),
		}
		options = append(options, option)
	}

	return options, nil
}

// GetFutures retrieves futures contracts
func (c *Client) GetFutures(ctx context.Context) ([]Future, error) {
	data := []byte(`{"page_number":1,"excludeZeroPxAndQty":true,"Content-Type":"application/json"}`)
	url := c.buildURL("index-future")

	respData, err := c.post(url, data)
	if err != nil {
		return nil, err
	}

	// Debug: log raw response
	c.debugLogResponse("index-future", respData)

	var rawFutures []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawFutures); err != nil {
		return nil, err
	}

	futures := make([]Future, 0, len(rawFutures))
	for _, raw := range rawFutures {
		future := Future{
			Symbol:        utils.GetString(raw, "symbol"),
			BidSize:       utils.GetInt64(raw, "quantityBid"),
			Bid:           utils.GetFloat64(raw, "bidPrice") * 1000, // Apply price multiplier for futures
			Ask:           utils.GetFloat64(raw, "offerPrice") * 1000,
			AskSize:       utils.GetInt64(raw, "quantityOffer"),
			Last:          utils.GetFloat64(raw, "settlementPrice") * 1000,
			Close:         utils.GetFloat64(raw, "closingPrice") * 1000,
			Change:        utils.GetFloat64(raw, "imbalance"),
			Open:          utils.GetFloat64(raw, "openingPrice") * 1000,
			High:          utils.GetFloat64(raw, "tradingHighPrice") * 1000,
			Low:           utils.GetFloat64(raw, "tradingLowPrice") * 1000,
			PreviousClose: utils.GetFloat64(raw, "previousClosingPrice") * 1000,
			Turnover:      utils.GetFloat64(raw, "volumeAmount") * 1000,
			Volume:        utils.GetInt64(raw, "volume") * 1000,
			Operations:    utils.GetInt64(raw, "numberOfOrders"),
			DateTime:      utils.ParseTradeTime(utils.GetString(raw, "tradeHour")),
			Expiration:    utils.GetTime(raw, "maturityDate"),
			OpenInterest:  utils.GetInt64(raw, "openInterest"),
		}
		futures = append(futures, future)
	}

	return futures, nil
}
