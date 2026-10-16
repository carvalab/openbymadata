package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	return c.doRequest(ctx, "GET", url, nil)
}

// post performs a POST request with retries
func (c *Client) post(ctx context.Context, url string, data []byte) ([]byte, error) {
	return c.doRequest(ctx, "POST", url, data)
}

// doRequest performs an HTTP request with retries and proper error handling
// Cancelling ctx aborts the request in flight and any pending retry.
func (c *Client) doRequest(ctx context.Context, method, url string, data []byte) ([]byte, error) {
	var lastErr error
	attempts := 0

	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		attempts++
		if attempt > 0 {
			waitTime := time.Duration(attempt) * time.Second
			c.logger.Debug("Retrying request",
				LogField{Key: "attempt", Value: attempt},
				LogField{Key: "wait_time", Value: waitTime},
				LogField{Key: "url", Value: url})
			timer := time.NewTimer(waitTime)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, fmt.Errorf("request cancelled after %d attempts: %w", attempts-1, ctx.Err())
			case <-timer.C:
			}
		}

		resp, err := c.makeRequest(ctx, method, url, data)
		if err != nil {
			lastErr = err
			if ctx.Err() != nil || !isRetryable(err) {
				break
			}
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", attempts, lastErr)
}

func (c *Client) makeRequest(ctx context.Context, method, url string, data []byte) ([]byte, error) {
	var body io.Reader
	if data != nil {
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	c.mu.RUnlock()
	if section, ok := sectionHeader[strings.TrimPrefix(req.URL.Path, c.basePath())]; ok {
		req.Header.Set("Options", section)
	}

	c.logger.Debug("Making request",
		LogField{Key: "method", Value: method},
		LogField{Key: "url", Value: url})

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &statusError{code: resp.StatusCode}
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

// apiPrefix is the path shared by every free BYMA Open Data endpoint.
const apiPrefix = "/vanoms-be-core/rest/api/bymadata/free/"

// buildURL constructs a full URL from the base URL and endpoint
func (c *Client) buildURL(endpoint string) string {
	return c.baseURL + apiPrefix + endpoint
}

// basePath is the URL path that precedes the endpoint name, including any
// path component of the configured base URL.
func (c *Client) basePath() string {
	if u, err := url.Parse(c.baseURL); err == nil {
		return strings.TrimSuffix(u.Path, "/") + apiPrefix
	}
	return apiPrefix
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

// statusError reports a non-2xx HTTP response.
type statusError struct {
	code int
}

func (e *statusError) Error() string {
	return fmt.Sprintf("HTTP error %d", e.code)
}

// isRetryable reports whether a failed request is worth retrying. Network
// errors, timeouts, 429 and 5xx are. So is 400: BYMA returns it at random
// under load and the same request then succeeds. Other 4xx won't change.
func isRetryable(err error) bool {
	var se *statusError
	if errors.As(err, &se) {
		switch se.code {
		case http.StatusBadRequest, http.StatusRequestTimeout, http.StatusTooManyRequests:
			return true
		}
		return se.code >= 500
	}
	return true
}
