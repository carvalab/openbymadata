package openbymadata

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/carvalab/openbymadata/internal/api"
)

// =============================================================================
// Client Interface
// =============================================================================

// Client defines the interface for BYMA data operations
type Client interface {
	// Market status and general info
	IsWorkingDay(ctx context.Context) (bool, error)
	GetIndices(ctx context.Context) ([]Index, error)
	MarketResume(ctx context.Context) ([]MarketSummary, error)

	// Securities
	GetBluechips(ctx context.Context) ([]Security, error)
	GetGalpones(ctx context.Context) ([]Security, error)
	GetCedears(ctx context.Context) ([]Security, error)

	// Fixed Income
	GetBonds(ctx context.Context) ([]Bond, error)
	GetShortTermBonds(ctx context.Context) ([]Bond, error)
	GetCorporateBonds(ctx context.Context) ([]Bond, error)

	// Derivatives
	GetOptions(ctx context.Context) ([]Option, error)
	GetFutures(ctx context.Context) ([]Future, error)

	// News and Financial Data
	GetNews(ctx context.Context) ([]News, error)
	GetIncomeStatement(ctx context.Context, ticker string) ([]IncomeStatement, error)

	// Individual security lookups
	GetSecurity(ctx context.Context, symbol string) (*Security, error)
	GetBluechip(ctx context.Context, symbol string) (*Security, error)
	GetCedear(ctx context.Context, symbol string) (*Security, error)
	GetGalpone(ctx context.Context, symbol string) (*Security, error)
	GetBond(ctx context.Context, symbol string) (*Bond, error)
	GetOption(ctx context.Context, symbol string) (*Option, error)
	GetFuture(ctx context.Context, symbol string) (*Future, error)

	// Batch operations
	GetMultipleSecurities(ctx context.Context, symbols []string) (map[string]*Security, error)
	SearchSecurities(ctx context.Context, searchText string) ([]Security, error)

	// Historical Data
	GetHistory(ctx context.Context, symbol, resolution string, from, to time.Time) (*OHLCV, error)
	GetHistoryLastDays(ctx context.Context, symbol string, days int) (*OHLCV, error)
	ConvertToHistoricalData(slices *OHLCV) ([]HistoricalData, error)

	// Cache management
	GetCacheInfo() map[string]interface{}
	ClearCache()
}

// =============================================================================
// Supporting Interfaces
// =============================================================================

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Get(url string) (*HTTPResponse, error)
	Post(url string, data []byte) (*HTTPResponse, error)
}

// Logger defines the interface for logging operations
type Logger interface {
	Debug(msg string, fields ...LogField)
	Info(msg string, fields ...LogField)
	Warn(msg string, fields ...LogField)
	Error(msg string, fields ...LogField)
}

// ConfigProvider defines the interface for configuration
type ConfigProvider interface {
	GetTimeout() time.Duration
	GetRetryAttempts() int
	GetBaseURL() string
}

// =============================================================================
// Data Models (Type Aliases)
// =============================================================================

// Type aliases to internal types
type (
	Security        = api.Security
	Bond            = api.Bond
	Option          = api.Option
	Future          = api.Future
	Index           = api.Index
	MarketSummary   = api.MarketSummary
	News            = api.News
	IncomeStatement = api.IncomeStatement
	HistoricalData  = api.HistoricalData
	OHLCV           = api.OHLCV
)

// =============================================================================
// Configuration Types
// =============================================================================

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	Body       []byte
	StatusCode int
	Headers    map[string]string
}

// LogField represents a structured log field
type LogField struct {
	Key   string
	Value interface{}
}

// APIResponse represents a generic API response wrapper
type APIResponse struct {
	Data    interface{} `json:"data"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
}

// MarketTimeResponse represents the market time API response
type MarketTimeResponse struct {
	IsWorkingDay bool `json:"isWorkingDay"`
}

// ClientOptions represents configuration options for the client
type ClientOptions struct {
	BaseURL       string
	Timeout       time.Duration
	RetryAttempts int
	Logger        Logger
	HTTPClient    HTTPClient
	EnableCache   bool // Enable 5-minute caching (default: true)
}

// DefaultClientOptions returns default client options
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		BaseURL:       "https://open.bymadata.com.ar",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		Logger:        &NoOpLogger{},
		HTTPClient:    nil,  // Will be created by client
		EnableCache:   true, // Cache enabled by default
	}
}

// NoOpLogger is a no-operation logger implementation
type NoOpLogger struct{}

func (l *NoOpLogger) Debug(msg string, fields ...LogField) {}
func (l *NoOpLogger) Info(msg string, fields ...LogField)  {}
func (l *NoOpLogger) Warn(msg string, fields ...LogField)  {}
func (l *NoOpLogger) Error(msg string, fields ...LogField) {}

// =============================================================================
// Error Types
// =============================================================================

// Error types for the BYMA library
var (
	ErrInvalidResponse = &BYMAError{Code: "INVALID_RESPONSE", Message: "Invalid API response"}
	ErrAPIUnavailable  = &BYMAError{Code: "API_UNAVAILABLE", Message: "BYMA API is unavailable"}
	ErrInvalidTicker   = &BYMAError{Code: "INVALID_TICKER", Message: "Invalid ticker symbol"}
	ErrTimeout         = &BYMAError{Code: "TIMEOUT", Message: "Request timeout"}
	ErrUnauthorized    = &BYMAError{Code: "UNAUTHORIZED", Message: "Unauthorized access"}
	ErrRateLimited     = &BYMAError{Code: "RATE_LIMITED", Message: "Rate limit exceeded"}
	ErrInternalError   = &BYMAError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
)

// BYMAError represents a custom error from the BYMA library
type BYMAError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
	Underlying error  `json:"-"`
}

// Error implements the error interface
func (e *BYMAError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s: %s (underlying: %v)", e.Code, e.Message, e.Underlying)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *BYMAError) Unwrap() error {
	return e.Underlying
}

// WithUnderlying adds an underlying error
func (e *BYMAError) WithUnderlying(err error) *BYMAError {
	return &BYMAError{
		Code:       e.Code,
		Message:    e.Message,
		StatusCode: e.StatusCode,
		Underlying: err,
	}
}

// WithStatusCode adds an HTTP status code
func (e *BYMAError) WithStatusCode(code int) *BYMAError {
	return &BYMAError{
		Code:       e.Code,
		Message:    e.Message,
		StatusCode: code,
		Underlying: e.Underlying,
	}
}

// NewBYMAError creates a new BYMA error
func NewBYMAError(code, message string) *BYMAError {
	return &BYMAError{
		Code:    code,
		Message: message,
	}
}

// MapHTTPError maps HTTP status codes to BYMA errors
func MapHTTPError(statusCode int) *BYMAError {
	switch statusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized.WithStatusCode(statusCode)
	case http.StatusTooManyRequests:
		return ErrRateLimited.WithStatusCode(statusCode)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return ErrAPIUnavailable.WithStatusCode(statusCode)
	case http.StatusRequestTimeout:
		return ErrTimeout.WithStatusCode(statusCode)
	default:
		return NewBYMAError("HTTP_ERROR", fmt.Sprintf("HTTP error %d", statusCode)).WithStatusCode(statusCode)
	}
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	if bymaErr, ok := err.(*BYMAError); ok {
		switch bymaErr.Code {
		case "TIMEOUT", "API_UNAVAILABLE", "RATE_LIMITED":
			return true
		case "HTTP_ERROR":
			return bymaErr.StatusCode >= 500
		}
	}
	return false
}
