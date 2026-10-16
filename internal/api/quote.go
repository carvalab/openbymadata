package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetCurrentQuote returns the latest quote for a symbol at the requested
// settlement window. The symbol is sent as provided; clients that need
// a specific suffix ("24HS", "48HS") should include it in the symbol string.
func (c *Client) GetCurrentQuote(ctx context.Context, symbol string, settlement Settlement) (CurrentQuote, error) {
	if _, ok := ParseSettlement(string(settlement)); !ok {
		return CurrentQuote{}, fmt.Errorf("invalid settlement %q", settlement)
	}

	payload := fmt.Sprintf(`{"symbol":%q,"settlementType":%q,"Content-Type":"application/json"}`, symbol, settlement)
	respData, err := c.post(ctx, c.buildURL(pathCurrentQuote), []byte(payload))
	if err != nil {
		return CurrentQuote{}, err
	}

	var envelope struct {
		Data []CurrentQuote `json:"data"`
	}
	if err := json.Unmarshal(respData, &envelope); err != nil {
		return CurrentQuote{}, fmt.Errorf("invalid quote response: %w", err)
	}
	if len(envelope.Data) == 0 {
		return CurrentQuote{}, fmt.Errorf("no quote available for %s (%s)", symbol, settlement)
	}
	return envelope.Data[0], nil
}
