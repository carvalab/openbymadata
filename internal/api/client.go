package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// LogField represents a structured log field
type LogField struct {
	Key   string
	Value interface{}
}

// Logger interface for logging within the internal package
type Logger interface {
	Debug(msg string, fields ...LogField)
	Info(msg string, fields ...LogField)
	Warn(msg string, fields ...LogField)
	Error(msg string, fields ...LogField)
}

// ClientOptions represents configuration options for the client
type ClientOptions struct {
	BaseURL       string
	Timeout       time.Duration
	RetryAttempts int
	Logger        Logger
}

// Client implements the openbymadata.Client interface
type Client struct {
	httpClient    *http.Client
	baseURL       string
	headers       map[string]string
	timeout       time.Duration
	retryAttempts int
	logger        Logger
	mu            sync.RWMutex
}

// New creates a new BYMA data client with the provided options.
func New(opts *ClientOptions) *Client {
	httpClient := &http.Client{
		Timeout: opts.Timeout,
	}

	return &Client{
		httpClient:    httpClient,
		baseURL:       opts.BaseURL,
		timeout:       opts.Timeout,
		retryAttempts: opts.RetryAttempts,
		logger:        opts.Logger,
		headers: map[string]string{
			"Connection":         "keep-alive",
			"sec-ch-ua":          `" Not A;Brand";v="99", "Chromium";v="96", "Google Chrome";v="96"`,
			"Accept":             "application/json, text/plain, */*",
			"Content-Type":       "application/json",
			"sec-ch-ua-mobile":   "?0",
			"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36",
			"sec-ch-ua-platform": `"Windows"`,
			"Origin":             opts.BaseURL,
			"Sec-Fetch-Site":     "same-origin",
			"Sec-Fetch-Mode":     "cors",
			"Sec-Fetch-Dest":     "empty",
			"Referer":            opts.BaseURL + "/",
			"Accept-Language":    "es-US,es-419;q=0.9,es;q=0.8,en;q=0.7",
		},
	}
}

// Close releases the underlying HTTP transport.
func (c *Client) Close() {
	if t, ok := c.httpClient.Transport.(io.Closer); ok {
		_ = t.Close()
	}
}

// get performs a GET request with retries
func (c *Client) get(url string) ([]byte, error) {
	return c.doRequest("GET", url, nil)
}

// post performs a POST request with retries
func (c *Client) post(url string, data []byte) ([]byte, error) {
	return c.doRequest("POST", url, data)
}

// doRequest performs an HTTP request with retries and proper error handling
func (c *Client) doRequest(method, url string, data []byte) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		if attempt > 0 {
			waitTime := time.Duration(attempt) * time.Second
			c.logger.Debug("Retrying request",
				LogField{Key: "attempt", Value: attempt},
				LogField{Key: "wait_time", Value: waitTime},
				LogField{Key: "url", Value: url})
			time.Sleep(waitTime)
		}

		resp, err := c.makeRequest(method, url, data)
		if err != nil {
			lastErr = err
			if !isRetryable(err) {
				break
			}
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryAttempts+1, lastErr)
}

func (c *Client) makeRequest(method, url string, data []byte) ([]byte, error) {
	var body io.Reader
	if data != nil {
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	c.mu.RUnlock()

	c.logger.Debug("Making request",
		LogField{Key: "method", Value: method},
		LogField{Key: "url", Value: url})

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.Debug("Request completed",
		LogField{Key: "status_code", Value: resp.StatusCode},
		LogField{Key: "response_size", Value: len(responseBody)})

	return responseBody, nil
}

// buildURL constructs a full URL from the base URL and endpoint
func (c *Client) buildURL(endpoint string) string {
	return c.baseURL + "/vanoms-be-core/rest/api/bymadata/free/" + endpoint
}

// parseAPIResponse parses a standard API response
func (c *Client) parseAPIResponse(data []byte, target interface{}) error {
	var apiResp struct {
		Data    interface{} `json:"data"`
		Success bool        `json:"success"`
		Message string      `json:"message"`
	}

	if err := json.Unmarshal(data, &apiResp); err != nil {
		if err := json.Unmarshal(data, target); err != nil {
			return fmt.Errorf("invalid response: %w", err)
		}
		return nil
	}

	if apiResp.Data == nil {
		return fmt.Errorf("no data in response")
	}

	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal target: %w", err)
	}

	return nil
}

// isRetryable determines if an error is retryable
func isRetryable(err error) bool {
	return true
}
