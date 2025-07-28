package api

import (
	"context"
	"fmt"

	"github.com/carvalab/openbymadata/internal/utils"
)

// GetNews retrieves market news
func (c *Client) GetNews(ctx context.Context) ([]News, error) {
	url := c.buildURL("bnown/byma-ads")
	respData, err := c.post(url, []byte(`{"Content-Type":"application/json"}`))
	if err != nil {
		return nil, err
	}

	// Debug: log raw response
	c.debugLogResponse("bnown/byma-ads", respData)

	var rawNews []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawNews); err != nil {
		return nil, err
	}

	news := make([]News, 0, len(rawNews))
	for _, raw := range rawNews {
		newsItem := News{
			Fecha:       utils.GetTime(raw, "fecha"),
			Titulo:      utils.GetString(raw, "emisor"),     // emisor is the company name (title)
			Descripcion: utils.GetString(raw, "referencia"), // referencia is the description
			Descarga:    "https://open.bymadata.com.ar/vanoms-be-core/rest/api/bymadata/free/sba/download/" + utils.GetString(raw, "descarga"),
		}
		news = append(news, newsItem)
	}

	return news, nil
}

// GetIncomeStatement retrieves income statement data for a specific ticker
func (c *Client) GetIncomeStatement(ctx context.Context, ticker string) ([]IncomeStatement, error) {
	url := c.buildURL("bnown/seriesHistoricas/balances")
	data := fmt.Sprintf(`{"symbol": "%s", "Content-Type": "application/json"}`, ticker)
	respData, err := c.post(url, []byte(data))
	if err != nil {
		return nil, err
	}

	// Debug: log raw response
	c.debugLogResponse("bnown/seriesHistoricas/balances", respData)

	var rawStatements []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawStatements); err != nil {
		return nil, err
	}

	statements := make([]IncomeStatement, 0, len(rawStatements))
	for _, raw := range rawStatements {
		statement := IncomeStatement{
			Symbol:          utils.GetString(raw, "symbol"),
			Periodo:         utils.GetString(raw, "periodo"),
			TipoPeriodo:     utils.GetString(raw, "tipoPeriodo"),
			FechaCierre:     utils.GetString(raw, "fechaCierre"),
			BalancesArchivo: "https://open.bymadata.com.ar/vanoms-be-core/rest/api/bymadata/free/sba/download/" + utils.GetString(raw, "balancesArchivo"),
		}
		statements = append(statements, statement)
	}

	return statements, nil
}
