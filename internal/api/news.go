package api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/carvalab/openbymadata/internal/utils"
)

// GetNews retrieves market news
func (c *Client) GetNews(ctx context.Context) ([]News, error) {
	respData, err := c.post(ctx, c.buildURL(pathNews), []byte(`{"Content-Type":"application/json"}`))
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
			Descarga:    downloadURL(raw, "descarga"),
		})
	}

	return news, nil
}

// GetIncomeStatement retrieves income statement data for a specific ticker
func (c *Client) GetIncomeStatement(ctx context.Context, ticker string) ([]IncomeStatement, error) {
	respData, err := c.post(ctx, c.buildURL(pathBalances), []byte(fmt.Sprintf(`{"symbol": "%s", "Content-Type": "application/json"}`, ticker)))
	if err != nil {
		return nil, err
	}

	var rawStatements []map[string]interface{}
	if err := c.parseAPIResponse(respData, &rawStatements); err != nil {
		return nil, err
	}

	statements := make([]IncomeStatement, 0, len(rawStatements))
	for _, raw := range rawStatements {
		link := downloadURL(raw, "archivo")
		if link == "" {
			link = utils.GetString(raw, "balancesArchivo")
		}
		statements = append(statements, IncomeStatement{
			Symbol:            ticker,
			Periodo:           utils.GetString(raw, "periodo"),
			TipoPeriodo:       utils.GetString(raw, "tipoPeriodo"),
			FechaCierre:       firstNonEmpty(dateOnly(utils.GetString(raw, "balancesFechaEstadoContable")), utils.GetString(raw, "fechaCierre")),
			FechaPresentacion: dateOnly(utils.GetString(raw, "balancesFechaPresentacion")),
			BalancesArchivo:   link,
		})
	}

	return statements, nil
}

// downloadBase serves the PDFs referenced by news and balance records.
const downloadBase = "https://open.bymadata.com.ar/vanoms-be-core/rest/api/bymadata/free/sba/download/"

// downloadURL builds a download link from a file id, which BYMA sends as a
// number (older payloads used a string). It returns "" when the id is missing.
func downloadURL(raw map[string]interface{}, key string) string {
	if id := utils.GetString(raw, key); id != "" {
		return downloadBase + id
	}
	if id := utils.GetInt64(raw, key); id != 0 {
		return downloadBase + strconv.FormatInt(id, 10)
	}
	return ""
}

// dateOnly trims BYMA timestamps like "2026-06-30 00:00:00.0" to "2026-06-30".
func dateOnly(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
