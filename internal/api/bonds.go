package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/carvalab/openbymadata/internal/utils"
)

// GetBonds retrieves government bonds.
func (c *Client) GetBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathPublicBonds)
}

// GetShortTermBonds retrieves short-term government bills (Letras). The
// upstream endpoint is still named "lebacs".
func (c *Client) GetShortTermBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathLetters)
}

// GetCorporateBonds retrieves corporate bonds (obligaciones negociables).
func (c *Client) GetCorporateBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathCorpBonds)
}

// GetGreenBonds retrieves bonds labelled green.
func (c *Client) GetGreenBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathGreenBonds)
}

// GetSocialBonds retrieves bonds labelled social.
func (c *Client) GetSocialBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathSocialBonds)
}

// GetSustainabilityBonds retrieves bonds labelled sustainable.
func (c *Client) GetSustainabilityBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathSustainBonds)
}

// GetLinkedBonds retrieves sustainability-linked bonds.
func (c *Client) GetLinkedBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathLinkedBonds)
}

// GetSenebiBonds retrieves government bonds traded in SENEBI, the
// bilateral segment without central counterparty. Symbols end in ".SB".
func (c *Client) GetSenebiBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathSenebiBonds)
}

// GetSenebiLetters retrieves Letras traded in SENEBI.
func (c *Client) GetSenebiLetters(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathSenebiLetters)
}

// GetSenebiCorporateBonds retrieves corporate bonds traded in SENEBI.
func (c *Client) GetSenebiCorporateBonds(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncome(ctx, pathSenebiCorpBonds)
}

// GetRepos retrieves cauciones (repo loans). Symbol encodes the term and
// currency, e.g. "PESOS-1910-U-CT-ARS"; Underlying is the collateral
// family. Prices are rates as BYMA publishes them: the scale differs
// between fields, so compare like with like.
func (c *Client) GetRepos(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncomeWith(ctx, pathRepos, plainBody)
}

// GetSecuritiesLending retrieves securities lending (préstamos de valores)
// contracts; Underlying is the lent security. Prices are rates, as in
// GetRepos.
func (c *Client) GetSecuritiesLending(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncomeWith(ctx, pathLending, plainBody)
}

// GetFinancialTrusts retrieves financial trusts (fideicomisos financieros).
func (c *Client) GetFinancialTrusts(ctx context.Context) ([]Bond, error) {
	return c.getFixedIncomeWith(ctx, pathTrusts, plainBody)
}

func (c *Client) getFixedIncome(ctx context.Context, endpoint string) ([]Bond, error) {
	return c.getFixedIncomeWith(ctx, endpoint, securitiesBody)
}

func (c *Client) getFixedIncomeWith(ctx context.Context, endpoint string, body []byte) ([]Bond, error) {
	respData, err := c.post(ctx, c.buildURL(endpoint), body)
	if err != nil {
		return nil, err
	}

	var rawBonds []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawBonds); err != nil {
		return nil, err
	}

	bonds := make([]Bond, 0, len(rawBonds))
	for _, raw := range rawBonds {
		bonds = append(bonds, parseBond(raw))
	}
	return bonds, nil
}

func parseBond(raw map[string]interface{}) Bond {
	return Bond{
		Symbol:         utils.GetString(raw, "symbol"),
		Settlement:     utils.GetString(raw, "settlementType"),
		BidSize:        utils.GetInt64(raw, "quantityBid"),
		Bid:            utils.GetFloat64(raw, "bidPrice"),
		Ask:            utils.GetFloat64(raw, "offerPrice"),
		AskSize:        utils.GetInt64(raw, "quantityOffer"),
		Last:           utils.GetFloat64(raw, "settlementPrice"),
		Close:          utils.GetFloat64(raw, "closingPrice"),
		Change:         utils.GetFloat64(raw, "imbalance"),
		Open:           utils.GetFloat64(raw, "openingPrice"),
		High:           utils.GetFloat64(raw, "tradingHighPrice"),
		Low:            utils.GetFloat64(raw, "tradingLowPrice"),
		PreviousClose:  utils.GetFloat64(raw, "previousClosingPrice"),
		Turnover:       utils.GetFloat64(raw, "volumeAmount"),
		Volume:         utils.GetInt64(raw, "volume"),
		Operations:     utils.GetInt64(raw, "numberOfOrders"),
		DateTime:       utils.ParseTradeTime(utils.GetString(raw, "tradeHour")),
		Group:          utils.GetString(raw, "securityType"),
		Expiration:     utils.GetTime(raw, "maturityDate"),
		DaysToMaturity: utils.GetInt64(raw, "daysToMaturity"),
		Underlying:     utils.GetString(raw, "underlyingSymbol"),
		Currency:       utils.GetString(raw, "denominationCcy"),
		Market:         utils.GetString(raw, "market"),
	}
}

// GetBondIndices retrieves BYMA's fixed-income indices.
func (c *Client) GetBondIndices(ctx context.Context) ([]BondIndex, error) {
	respData, err := c.post(ctx, c.buildURL(pathBondIndices), plainBody)
	if err != nil {
		return nil, err
	}
	var raws []map[string]interface{}
	if err := c.parseAPIResponse(respData, &raws); err != nil {
		return nil, err
	}
	out := make([]BondIndex, 0, len(raws))
	for _, raw := range raws {
		out = append(out, BondIndex{
			Symbol:      utils.GetString(raw, "symbol"),
			Description: utils.GetString(raw, "description"),
			Currency:    utils.GetString(raw, "currency"),
			Value:       utils.GetFloat64(raw, "closePrice"),
			Varianza:    utils.GetFloat64(raw, "varianza"),
			Date:        utils.GetTime(raw, "date_"),
		})
	}
	return out, nil
}

// GetBondCoupons retrieves the payment schedule (interest and amortization)
// published for a bond, e.g. "GD30".
func (c *Client) GetBondCoupons(ctx context.Context, symbol string) ([]BondCoupon, error) {
	respData, err := c.post(ctx, c.buildURL(pathCoupons), symbolBody(symbol))
	if err != nil {
		return nil, err
	}
	var raws []map[string]interface{}
	if err := c.parseAPIResponse(respData, &raws); err != nil {
		return nil, err
	}
	out := make([]BondCoupon, 0, len(raws))
	for _, raw := range raws {
		out = append(out, BondCoupon{
			Symbol:                    strings.ToUpper(symbol),
			NoticeDate:                utils.GetTime(raw, "couponsFechaAvisoPago"),
			InterestPaymentDate:       utils.GetTime(raw, "couponsFechaPagoRenta"),
			AmortizationPaymentDate:   utils.GetTime(raw, "couponsFechaPagoAmort"),
			InterestServiceNumber:     utils.GetInt64(raw, "couponsNumServRenta"),
			AmortizationServiceNumber: utils.GetInt64(raw, "couponsNumServAmort"),
			AmortizationPercent:       utils.GetFloat64(raw, "couponsPorcAmort"),
			ResidualValuePercent:      utils.GetFloat64(raw, "couponsPorcValResVig"),
			NoticeURL:                 downloadURL(raw, "archivo"),
		})
	}
	return out, nil
}

// symbolBody is the request body for per-symbol endpoints.
func symbolBody(symbol string) []byte {
	return []byte(fmt.Sprintf(`{"symbol":%q,"Content-Type":"application/json"}`, symbol))
}
