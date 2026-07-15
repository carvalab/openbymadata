package api

import (
	"context"
	"fmt"

	"github.com/carvalab/openbymadata/internal/utils"
)

// GetNews retrieves market news
func (c *Client) GetNews(ctx context.Context) ([]News, error) {
	respData, err := c.post(c.buildURL(pathNews), []byte(`{"Content-Type":"application/json"}`))
	if err != nil {
		return nil, err
	}

	var rawNews []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawNews); err != nil {
		return nil, err
	}

	news := make([]News, 0, len(rawNews))
	for _, raw := range rawNews {
		news = append(news, News{
			Fecha:       utils.GetTime(raw, "fecha"),
			Titulo:      utils.GetString(raw, "emisor"),
			Descripcion: utils.GetString(raw, "referencia"),
			Descarga:    "https://open.bymadata.com.ar/vanoms-be-core/rest/api/bymadata/free/sba/download/" + utils.GetString(raw, "descarga"),
		})
	}

	return news, nil
}

// GetIncomeStatement retrieves income statement data for a specific ticker
func (c *Client) GetIncomeStatement(ctx context.Context, ticker string) ([]IncomeStatement, error) {
	respData, err := c.post(c.buildURL(pathBalances), []byte(fmt.Sprintf(`{"symbol": "%s", "Content-Type": "application/json"}`, ticker)))
	if err != nil {
		return nil, err
	}

	var rawStatements []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawStatements); err != nil {
		return nil, err
	}

	statements := make([]IncomeStatement, 0, len(rawStatements))
	for _, raw := range rawStatements {
		statements = append(statements, IncomeStatement{
			Symbol:          utils.GetString(raw, "symbol"),
			Periodo:         utils.GetString(raw, "periodo"),
			TipoPeriodo:     utils.GetString(raw, "tipoPeriodo"),
			FechaCierre:     utils.GetString(raw, "fechaCierre"),
			BalancesArchivo: "https://open.bymadata.com.ar/vanoms-be-core/rest/api/bymadata/free/sba/download/" + utils.GetString(raw, "balancesArchivo"),
		})
	}

	return statements, nil
}
