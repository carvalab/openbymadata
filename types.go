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

// Client defines the interface for BYMA data operations.
type Client interface {
	Close()

	// Market status and general info
	IsWorkingDay(ctx context.Context) (bool, error)
	GetMarketTime(ctx context.Context) (MarketTime, error)
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
	GetCurrentQuote(ctx context.Context, symbol string, settlement Settlement) (CurrentQuote, error)

	// Batch operations
	GetMultipleSecurities(ctx context.Context, symbols []string) (map[string]*Security, error)
	SearchSecurities(ctx context.Context, searchText string) ([]Security, error)

	// Historical Data
	GetHistory(ctx context.Context, symbol string, resolution Resolution, from, to time.Time) (*OHLCV, error)
	GetHistoryLastDays(ctx context.Context, symbol string, days int) (*OHLCV, error)
	GetIntradayHistory(ctx context.Context, symbol string, resolution Resolution, from, to time.Time) (*OHLCV, error)

	// Company metadata
	GetCompanyInfo(ctx context.Context, symbol string) (*CompanyInfo, error)
	GetEquityProfile(ctx context.Context, symbol string) (*EquityProfile, error)
	GetCompanyManagement(ctx context.Context, symbol string) ([]CompanyDirector, error)
	GetCompanyBalance(ctx context.Context, symbol string) ([]CompanyBalance, error)

	// Deprecated aliases
	GetLeadingEquity(ctx context.Context) ([]Security, error)
	GetGeneralEquity(ctx context.Context) ([]Security, error)
	GetPublicBonds(ctx context.Context) ([]Bond, error)
	GetLetters(ctx context.Context) ([]Bond, error)

	// Cache management
	GetCacheInfo() map[string]interface{}
	ClearCache()
}

// =============================================================================
// Supporting Interfaces
// =============================================================================

// Logger defines the interface for logging operations
type Logger = api.Logger

// LogField represents a structured log field
type LogField = api.LogField

// Settlement identifies the settlement window for quote endpoints.
type Settlement = api.Settlement

// Resolution identifies the timeframe for historical data points.
type Resolution = api.Resolution

// Settlement window constants accepted by GetCurrentQuote.
const (
	SettlementCI   = api.SettlementCI
	Settlement24HS = api.Settlement24HS
	Settlement48HS = api.Settlement48HS
)

// Resolution constants accepted by GetHistory and GetIntradayHistory.
const (
	Resolution1Min    = api.Resolution1Min
	Resolution5Min    = api.Resolution5Min
	Resolution60Min   = api.Resolution60Min
	ResolutionDaily   = api.ResolutionDaily
	ResolutionWeekly  = api.ResolutionWeekly
	ResolutionMonthly = api.ResolutionMonthly
)

// ParseSettlement accepts either numeric ("1"/"2"/"3") or textual
// ("CI"/"24HS"/"48HS") settlement codes.
func ParseSettlement(value string) (Settlement, bool) {
	return api.ParseSettlement(value)
}

// =============================================================================
// Data Models (Type Aliases)
// =============================================================================

type (
	Security        = api.Security
	Bond            = api.Bond
	Option          = api.Option
	Future          = api.Future
	Index           = api.Index
	MarketSummary   = api.MarketSummary
	News            = api.News
	IncomeStatement = api.IncomeStatement
	OHLCV           = api.OHLCV
	MarketTime      = api.MarketTime
	CurrentQuote    = api.CurrentQuote
	CompanyInfo     = api.CompanyInfo
	EquityProfile   = api.EquityProfile
	CompanyDirector = api.CompanyDirector
	CompanyBalance  = api.CompanyBalance
)

// =============================================================================
// Configuration Types
// =============================================================================

// ClientOptions represents configuration options for the client
type ClientOptions struct {
	BaseURL       string
	Timeout       time.Duration
	RetryAttempts int
	Logger        Logger
	EnableCache   bool // Enable 5-minute caching. On with NewClient() or DefaultClientOptions(); set it explicitly when passing your own options.
}

// DefaultClientOptions returns default client options
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		BaseURL:       "https://open.bymadata.com.ar",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		Logger:        &NoOpLogger{},
		EnableCache:   true,
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

var (
	ErrInvalidResponse = &BYMAError{Code: "INVALID_RESPONSE", Message: "Invalid API response"}
	ErrAPIUnavailable  = &BYMAError{Code: "API_UNAVAILABLE", Message: "BYMA API is unavailable"}
	ErrInvalidTicker   = &BYMAError{Code: "INVALID_TICKER", Message: "Invalid ticker symbol"}
	ErrTimeout         = &BYMAError{Code: "TIMEOUT", Message: "Request timeout"}
	ErrUnauthorized    = &BYMAError{Code: "UNAUTHORIZED", Message: "Unauthorized access"}
	ErrRateLimited     = &BYMAError{Code: "RATE_LIMITED", Message: "Rate limit exceeded"}
	ErrInternalError   = &BYMAError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
)

type BYMAError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
	Underlying error  `json:"-"`
}

func (e *BYMAError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s: %s (underlying: %v)", e.Code, e.Message, e.Underlying)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *BYMAError) Unwrap() error {
	return e.Underlying
}

func (e *BYMAError) WithUnderlying(err error) *BYMAError {
	return &BYMAError{Code: e.Code, Message: e.Message, StatusCode: e.StatusCode, Underlying: err}
}

func (e *BYMAError) WithStatusCode(code int) *BYMAError {
	return &BYMAError{Code: e.Code, Message: e.Message, StatusCode: code, Underlying: e.Underlying}
}

func NewBYMAError(code, message string) *BYMAError {
	return &BYMAError{Code: code, Message: message}
}

func MapHTTPError(statusCode int) *BYMAError {
	switch statusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized.WithStatusCode(statusCode)
	case http.StatusTooManyRequests:
		return ErrRateLimited.WithStatusCode(statusCode)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return ErrAPIUnavailable.WithStatusCode(statusCode)
	case http.StatusRequestTimeout:
		return ErrTimeout.WithStatusCode(statusCode)
	default:
		return NewBYMAError("HTTP_ERROR", http.StatusText(statusCode)).WithStatusCode(statusCode)
	}
}

func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if be, ok := err.(*BYMAError); ok {
		switch be.Code {
		case "INVALID_RESPONSE", "INVALID_TICKER", "UNAUTHORIZED", "HTTP_ERROR":
			return false
		}
	}
	return true
}
