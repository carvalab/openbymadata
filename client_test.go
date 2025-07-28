package openbymadata

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		opts    *ClientOptions
		wantErr bool
	}{
		{
			name: "default options",
			opts: nil,
		},
		{
			name: "custom options",
			opts: &ClientOptions{
				BaseURL:       "https://custom.example.com",
				Timeout:       15 * time.Second,
				RetryAttempts: 2,
				Logger:        &NoOpLogger{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client Client
			if tt.opts != nil {
				client = NewClient(tt.opts)
			} else {
				client = NewClient()
			}

			assert.NotNil(t, client)
		})
	}
}

func TestClient_IsWorkingDay(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		wantWorking    bool
		wantErr        bool
	}{
		{
			name:           "working day",
			responseBody:   `{"isWorkingDay": true}`,
			responseStatus: http.StatusOK,
			wantWorking:    true,
			wantErr:        false,
		},
		{
			name:           "not working day",
			responseBody:   `{"isWorkingDay": false}`,
			responseStatus: http.StatusOK,
			wantWorking:    false,
			wantErr:        false,
		},
		{
			name:           "server error",
			responseBody:   ``,
			responseStatus: http.StatusInternalServerError,
			wantWorking:    false,
			wantErr:        true,
		},
		{
			name:           "invalid response",
			responseBody:   `{"invalid": "json"}`,
			responseStatus: http.StatusOK,
			wantWorking:    false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := createTestClient(server.URL)
			ctx := context.Background()

			working, err := client.IsWorkingDay(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantWorking, working)
			}
		})
	}
}

func TestClient_GetIndices(t *testing.T) {
	mockResponse := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"description":          "Merval",
				"symbol":               "MERV",
				"price":                1234.56,
				"variation":            12.34,
				"highValue":            1245.00,
				"minValue":             1220.00,
				"previousClosingPrice": 1222.22,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := createTestClient(server.URL)
	ctx := context.Background()

	indices, err := client.GetIndices(ctx)

	require.NoError(t, err)
	require.Len(t, indices, 1)

	index := indices[0]
	assert.Equal(t, "Merval", index.Description)
	assert.Equal(t, "MERV", index.Symbol)
	assert.Equal(t, 1234.56, index.Last)
	assert.Equal(t, 12.34, index.Change)
	assert.Equal(t, 1245.00, index.High)
	assert.Equal(t, 1220.00, index.Low)
	assert.Equal(t, 1222.22, index.PreviousClose)
}

func TestClient_GetBluechips(t *testing.T) {
	mockResponse := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"symbol":               "GGAL",
				"settlementType":       "48hs",
				"quantityBid":          1000,
				"bidPrice":             150.50,
				"offerPrice":           151.00,
				"quantityOffer":        500,
				"settlementPrice":      150.75,
				"closingPrice":         150.00,
				"imbalance":            0.75,
				"openingPrice":         149.50,
				"tradingHighPrice":     151.50,
				"tradingLowPrice":      149.00,
				"previousClosingPrice": 150.00,
				"volumeAmount":         1500000.00,
				"volume":               10000,
				"numberOfOrders":       50,
				"tradeHour":            "16:00:00",
				"securityType":         "EQUITY",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := createTestClient(server.URL)
	ctx := context.Background()

	securities, err := client.GetBluechips(ctx)

	require.NoError(t, err)
	require.Len(t, securities, 1)

	security := securities[0]
	assert.Equal(t, "GGAL", security.Symbol)
	assert.Equal(t, "48hs", security.Settlement)
	assert.Equal(t, int64(1000), security.BidSize)
	assert.Equal(t, 150.50, security.Bid)
	assert.Equal(t, 151.00, security.Ask)
	assert.Equal(t, int64(500), security.AskSize)
	assert.Equal(t, 150.75, security.Last)
	assert.Equal(t, 150.00, security.Close)
	assert.Equal(t, 0.75, security.Change)
	assert.Equal(t, 149.50, security.Open)
	assert.Equal(t, 151.50, security.High)
	assert.Equal(t, 149.00, security.Low)
	assert.Equal(t, 150.00, security.PreviousClose)
	assert.Equal(t, 1500000.00, security.Turnover)
	assert.Equal(t, int64(10000), security.Volume)
	assert.Equal(t, int64(50), security.Operations)
	assert.Equal(t, "EQUITY", security.Group)
}

func TestClient_GetNews(t *testing.T) {
	mockResponse := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"fecha":       "2023-12-01T10:00:00Z",
				"titulo":      "Test News Title",
				"descripcion": "Test news description",
				"descarga":    "test-file.pdf",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := createTestClient(server.URL)
	ctx := context.Background()

	news, err := client.GetNews(ctx)

	require.NoError(t, err)
	require.Len(t, news, 1)

	newsItem := news[0]
	assert.Equal(t, "Test News Title", newsItem.Titulo)
	assert.Equal(t, "Test news description", newsItem.Descripcion)
	assert.Contains(t, newsItem.Descarga, "test-file.pdf")
	assert.False(t, newsItem.Fecha.IsZero())
}

func TestBYMAError(t *testing.T) {
	tests := []struct {
		name      string
		err       *BYMAError
		wantStr   string
		retryable bool
	}{
		{
			name:      "simple error",
			err:       ErrInvalidResponse,
			wantStr:   "INVALID_RESPONSE: Invalid API response",
			retryable: false,
		},
		{
			name:      "error with underlying",
			err:       ErrAPIUnavailable.WithUnderlying(assert.AnError),
			retryable: true,
		},
		{
			name:      "error with status code",
			err:       ErrTimeout.WithStatusCode(408),
			retryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantStr != "" {
				assert.Equal(t, tt.wantStr, tt.err.Error())
			}
			assert.Equal(t, tt.retryable, IsRetryable(tt.err))
		})
	}
}

func TestMapHTTPError(t *testing.T) {
	tests := []struct {
		statusCode int
		wantCode   string
	}{
		{http.StatusUnauthorized, "UNAUTHORIZED"},
		{http.StatusTooManyRequests, "RATE_LIMITED"},
		{http.StatusInternalServerError, "API_UNAVAILABLE"},
		{http.StatusRequestTimeout, "TIMEOUT"},
		{http.StatusNotFound, "HTTP_ERROR"},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.statusCode), func(t *testing.T) {
			err := MapHTTPError(tt.statusCode)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.statusCode, err.StatusCode)
		})
	}
}

// createTestClient creates a client configured for testing with a test server
func createTestClient(baseURL string) Client {
	opts := &ClientOptions{
		BaseURL:       baseURL,
		Timeout:       5 * time.Second,
		RetryAttempts: 1,
		Logger:        &NoOpLogger{},
	}
	return NewClient(opts)
}
