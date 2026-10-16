package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/carvalab/openbymadata/internal/utils"
)

// SearchInstruments looks up instruments of any type by symbol or name,
// the same search as the open.bymadata.com.ar header. BYMA returns at most
// 10 matches.
func (c *Client) SearchInstruments(ctx context.Context, query string) ([]InstrumentMatch, error) {
	body, err := json.Marshal(map[string]string{"value": query, "Content-Type": "application/json"})
	if err != nil {
		return nil, err
	}
	respData, err := c.post(ctx, c.buildURL(pathSearch), body)
	if err != nil {
		return nil, err
	}
	var raws []map[string]interface{}
	if err := c.parseAPIResponse(respData, &raws); err != nil {
		return nil, fmt.Errorf("invalid search response: %w", err)
	}
	out := make([]InstrumentMatch, 0, len(raws))
	for _, raw := range raws {
		out = append(out, InstrumentMatch{
			Symbol:       utils.GetString(raw, "symbol"),
			Description:  utils.GetString(raw, "description"),
			Type:         utils.GetString(raw, "type"),
			Market:       utils.GetString(raw, "market"),
			Settlement:   utils.GetString(raw, "vencimiento"),
			Presentation: utils.GetString(raw, "presentation"),
		})
	}
	return out, nil
}

// GetInstrumentCatalog returns every instrument BYMA lists (about 11,000),
// including SENEBI (".SB") variants. Use it to validate or enumerate
// tickers; SearchInstruments is better for lookups.
func (c *Client) GetInstrumentCatalog(ctx context.Context) ([]CatalogInstrument, error) {
	respData, err := c.get(ctx, c.buildURL(pathCatalog))
	if err != nil {
		return nil, err
	}
	var raws []map[string]interface{}
	if err := c.parseAPIResponse(respData, &raws); err != nil {
		return nil, fmt.Errorf("invalid catalog response: %w", err)
	}
	out := make([]CatalogInstrument, 0, len(raws))
	for _, raw := range raws {
		deleted, _ := raw["deleted"].(bool)
		out = append(out, CatalogInstrument{
			ID:      utils.GetInt64(raw, "oid"),
			Code:    utils.GetString(raw, "code"),
			Kind:    utils.GetString(raw, "description"),
			Deleted: deleted,
		})
	}
	return out, nil
}
