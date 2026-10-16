package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/carvalab/openbymadata/internal/utils"
)

// GetCompanyInfo returns the general information (society/company data) for
// the given symbol, flattened into a key/value record set.
func (c *Client) GetCompanyInfo(ctx context.Context, symbol string) (*CompanyInfo, error) {
	record, err := c.fetchCompanyRecord(ctx, pathCompanyGeneral, symbol)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return &CompanyInfo{Symbol: symbol}, nil
	}
	return &CompanyInfo{Symbol: symbol, Fields: flattenRecord(record)}, nil
}

// GetEquityProfile returns the technical profile for the given symbol.
func (c *Client) GetEquityProfile(ctx context.Context, symbol string) (*EquityProfile, error) {
	record, err := c.fetchCompanyRecord(ctx, pathCompanyProfile, symbol)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return &EquityProfile{Symbol: symbol}, nil
	}
	return &EquityProfile{Symbol: symbol, Fields: flattenRecord(record)}, nil
}

// GetCompanyManagement returns the directors/management roster for a society.
func (c *Client) GetCompanyManagement(ctx context.Context, symbol string) ([]CompanyDirector, error) {
	respData, err := c.post(c.buildURL(pathCompanyMgmt), []byte(fmt.Sprintf(`{"symbol":%q,"Content-Type":"application/json"}`, symbol)))
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(respData, &envelope); err != nil {
		return nil, fmt.Errorf("invalid management response: %w", err)
	}
	directors := make([]CompanyDirector, 0, len(envelope.Data))
	for _, raw := range envelope.Data {
		// BYMA sends the name as "persona" and pads "cargo" with spaces.
		directors = append(directors, CompanyDirector{
			Name:     strings.TrimSpace(firstNonEmpty(utils.GetString(raw, "persona"), utils.GetString(raw, "nombre"))),
			Position: strings.TrimSpace(utils.GetString(raw, "cargo")),
			Since:    strings.TrimSpace(utils.GetString(raw, "desde")),
			Until:    strings.TrimSpace(utils.GetString(raw, "hasta")),
			Type:     strings.TrimSpace(utils.GetString(raw, "tipo")),
		})
	}
	return directors, nil
}

// GetCompanyBalance returns the latest balance sheet for the symbol.
// Since 2026 BYMA redirects this endpoint to an OAuth login, so it fails
// on the free API. Use GetIncomeStatement for the filed statement PDFs.
func (c *Client) GetCompanyBalance(ctx context.Context, symbol string) ([]CompanyBalance, error) {
	respData, err := c.post(c.buildURL(pathCompanyBalance), []byte(fmt.Sprintf(`{"symbol":%q,"Content-Type":"application/json"}`, symbol)))
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(respData, &envelope); err != nil {
		return nil, fmt.Errorf("invalid balance response: %w", err)
	}
	if len(envelope.Data) == 0 {
		return []CompanyBalance{}, nil
	}

	first := envelope.Data[0]
	rawAccounts, _ := first["Cuentas"].([]interface{})
	balances := make([]CompanyBalance, 0, len(rawAccounts))
	for _, item := range rawAccounts {
		record, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		balances = append(balances, CompanyBalance{
			Account: utils.GetString(record, "Cuenta"),
			Values:  flattenRecord(record),
		})
	}
	return balances, nil
}

func (c *Client) fetchCompanyRecord(ctx context.Context, endpoint, symbol string) (map[string]interface{}, error) {
	respData, err := c.post(c.buildURL(endpoint), []byte(fmt.Sprintf(`{"symbol":%q,"Content-Type":"application/json"}`, symbol)))
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(respData, &envelope); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	if len(envelope.Data) == 0 {
		return nil, nil
	}
	return envelope.Data[0], nil
}

func flattenRecord(record map[string]interface{}) map[string]string {
	out := make(map[string]string, len(record))
	for k, v := range record {
		out[k] = stringifyValue(v)
	}
	return out
}

func stringifyValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(fmt.Sprintf("%v", val), "}"), "{"))
	}
}

// GetMarketResponsibles returns the company's designated market
// representatives (responsables de relaciones con el mercado).
func (c *Client) GetMarketResponsibles(ctx context.Context, symbol string) ([]MarketResponsible, error) {
	respData, err := c.post(c.buildURL(pathResponsibles), symbolBody(symbol))
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(respData, &envelope); err != nil {
		return nil, fmt.Errorf("invalid responsibles response: %w", err)
	}
	out := make([]MarketResponsible, 0, len(envelope.Data))
	for _, raw := range envelope.Data {
		out = append(out, MarketResponsible{
			Name:            strings.TrimSpace(utils.GetString(raw, "responsableMercado")),
			Position:        strings.TrimSpace(utils.GetString(raw, "cargo")),
			CompanyPosition: strings.TrimSpace(utils.GetString(raw, "cargoEmpresa")),
			Since:           utils.GetTime(raw, "fechaDesignacion"),
		})
	}
	return out, nil
}
