package api

import (
	"context"

	"github.com/carvalab/openbymadata/internal/utils"
)

// securitiesBody asks for T+1 (24hs) prices, which is what the web app shows.
var securitiesBody = []byte(`{"excludeZeroPxAndQty":false,"T2":false,"T1":true,"T0":false,"Content-Type":"application/json"}`)

// plainBody is what the web app sends to endpoints without filters.
var plainBody = []byte(`{"Content-Type":"application/json"}`)

// GetBluechips retrieves leading equity securities (blue chip stocks)
func (c *Client) GetBluechips(ctx context.Context) ([]Security, error) {
	return c.getSecurities(ctx, pathLeadingEquity)
}

// GetGalpones retrieves general equity securities
func (c *Client) GetGalpones(ctx context.Context) ([]Security, error) {
	return c.getSecurities(ctx, pathGeneralEquity)
}

// GetCedears retrieves CEDEAR securities
func (c *Client) GetCedears(ctx context.Context) ([]Security, error) {
	return c.getSecurities(ctx, pathCedears)
}

// GetETFs retrieves the ETFs listed on BYMA.
func (c *Client) GetETFs(ctx context.Context) ([]Security, error) {
	return c.getSecurities(ctx, pathETF)
}

// GetCorporateGovernance retrieves the stocks in BYMA's Corporate
// Governance panel.
func (c *Client) GetCorporateGovernance(ctx context.Context) ([]Security, error) {
	return c.getSecuritiesWith(ctx, pathGovernance, plainBody)
}

// GetCorporateGovernancePlus retrieves the stocks in BYMA's Corporate
// Governance Plus panel.
func (c *Client) GetCorporateGovernancePlus(ctx context.Context) ([]Security, error) {
	return c.getSecuritiesWith(ctx, pathGovernancePlus, plainBody)
}

func (c *Client) getSecurities(ctx context.Context, endpoint string) ([]Security, error) {
	return c.getSecuritiesWith(ctx, endpoint, securitiesBody)
}

func (c *Client) getSecuritiesWith(ctx context.Context, endpoint string, body []byte) ([]Security, error) {
	respData, err := c.post(ctx, c.buildURL(endpoint), body)
	if err != nil {
		return nil, err
	}

	// Some endpoints wrap the list in {"data": [...]}, others return it bare;
	// parseAPIResponse handles both.
	var rawSecurities []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawSecurities); err != nil {
		return nil, err
	}

	securities := make([]Security, 0, len(rawSecurities))
	for _, raw := range rawSecurities {
		securities = append(securities, parseSecurity(raw))
	}
	return securities, nil
}

func parseSecurity(raw map[string]interface{}) Security {
	return Security{
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
		Currency:      utils.GetString(raw, "denominationCcy"),
		Market:        utils.GetString(raw, "market"),
	}
}
