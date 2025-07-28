package openbymadata_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/carvalab/openbymadata"
)

// TestClientCreation tests basic client creation
func TestClientCreation(t *testing.T) {
	// Test default client creation
	client := openbymadata.NewClient()
	if client == nil {
		t.Error("Expected client to be created, got nil")
	}

	// Test client creation with custom options
	opts := &openbymadata.ClientOptions{
		Timeout:       15 * time.Second,
		RetryAttempts: 5,
	}
	clientWithOpts := openbymadata.NewClient(opts)
	if clientWithOpts == nil {
		t.Error("Expected client with options to be created, got nil")
	}
}

// TestErrorTypes tests error type creation and methods
func TestErrorTypes(t *testing.T) {
	// Test creating a custom error
	err := openbymadata.NewBYMAError("TEST_ERROR", "Test error message")
	if err.Error() != "TEST_ERROR: Test error message" {
		t.Errorf("Expected 'TEST_ERROR: Test error message', got '%s'", err.Error())
	}

	// Test error with underlying error
	underlying := fmt.Errorf("underlying error")
	errWithUnderlying := err.WithUnderlying(underlying)
	expected := "TEST_ERROR: Test error message (underlying: underlying error)"
	if errWithUnderlying.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, errWithUnderlying.Error())
	}

	// Test error with status code
	errWithStatus := err.WithStatusCode(404)
	if errWithStatus.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", errWithStatus.StatusCode)
	}
}

// TestRetryLogic tests the retry logic for errors
func TestRetryLogic(t *testing.T) {
	tests := []struct {
		name     string
		err      *openbymadata.BYMAError
		expected bool
	}{
		{
			name:     "Timeout error should be retryable",
			err:      openbymadata.ErrTimeout,
			expected: true,
		},
		{
			name:     "API unavailable should be retryable",
			err:      openbymadata.ErrAPIUnavailable,
			expected: true,
		},
		{
			name:     "Rate limited should be retryable",
			err:      openbymadata.ErrRateLimited,
			expected: true,
		},
		{
			name:     "Invalid response should not be retryable",
			err:      openbymadata.ErrInvalidResponse,
			expected: false,
		},
		{
			name:     "HTTP 500 error should be retryable",
			err:      openbymadata.NewBYMAError("HTTP_ERROR", "Server error").WithStatusCode(500),
			expected: true,
		},
		{
			name:     "HTTP 400 error should not be retryable",
			err:      openbymadata.NewBYMAError("HTTP_ERROR", "Bad request").WithStatusCode(400),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := openbymadata.IsRetryable(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for error: %s", tt.expected, result, tt.err.Error())
			}
		})
	}
}

// BenchmarkClientCreation benchmarks client creation
func BenchmarkClientCreation(b *testing.B) {
	opts := &openbymadata.ClientOptions{
		Timeout:       10 * time.Second,
		RetryAttempts: 3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = openbymadata.NewClient(opts)
	}
}
