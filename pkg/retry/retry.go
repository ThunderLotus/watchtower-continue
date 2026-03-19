package retry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	log "github.com/sirupsen/logrus"
)

// Config holds the retry configuration
type Config struct {
	MaxAttempts   uint64        // Maximum number of retry attempts
	InitialDelay  time.Duration // Initial delay before first retry
	MaxDelay      time.Duration // Maximum delay between retries
	Multiplier    float64       // Multiplier for exponential backoff
	Randomization float64       // Randomization factor for jitter
	EnableRetry   bool          // Whether retry is enabled
}

// DefaultConfig returns the default retry configuration
func DefaultConfig() *Config {
	return &Config{
		MaxAttempts:   5,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		Multiplier:    2.0,
		Randomization: 0.1,
		EnableRetry:   true,
	}
}

// Stats holds retry statistics
type Stats struct {
	TotalAttempts int
	Successful    bool
	Duration      time.Duration
	LastError     error
	Operation     string
}

// RetryableError is an error that can be retried
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable error: %v", e.Err)
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NonRetryableError is an error that should not be retried
type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return fmt.Sprintf("non-retryable error: %v", e.Err)
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

// NewRetryableError wraps an error as retryable
func NewRetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &RetryableError{Err: err}
}

// NewNonRetryableError wraps an error as non-retryable
func NewNonRetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &NonRetryableError{Err: err}
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's explicitly non-retryable
	var nonRetryable *NonRetryableError
	if errors.As(err, &nonRetryable) {
		return false
	}

	// Check if it's explicitly retryable
	var retryable *RetryableError
	if errors.As(err, &retryable) {
		return true
	}

	// Check if it's a network error (temporary)
	return isTemporaryError(err)
}

// isTemporaryError checks if an error is temporary and can be retried
func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	// Network timeout errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}

	// Temporary errors
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	// Connection refused errors
	if errors.Is(err, net.ErrClosed) {
		return true
	}

	// Check for specific error messages that indicate temporary failures
	errMsg := strings.ToLower(err.Error())
	temporaryPatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"temporary error",
		"network is unreachable",
		"no route to host",
		"i/o timeout",
		"eof",
		"broken pipe",
	}

	for _, pattern := range temporaryPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// WithRetry executes a function with retry logic
func WithRetry(ctx context.Context, config *Config, operation string, fn func() error) (Stats, error) {
	if !config.EnableRetry {
		err := fn()
		return Stats{
			TotalAttempts: 1,
			Successful:    err == nil,
			Duration:      0,
			LastError:     err,
			Operation:     operation,
		}, err
	}

	log.Debugf("Starting retry operation: %s", operation)
	startTime := time.Now()
	stats := Stats{
		Operation: operation,
	}

	// Create exponential backoff with jitter
	backoffStrategy := backoff.NewExponentialBackOff()
	backoffStrategy.InitialInterval = config.InitialDelay
	backoffStrategy.MaxInterval = config.MaxDelay
	backoffStrategy.Multiplier = config.Multiplier
	backoffStrategy.RandomizationFactor = config.Randomization

	// Wrap with context
	backoffContext := backoff.WithContext(backoffStrategy, ctx)

	// Retry function
	var attempts uint64 = 0
	retryableFunc := func() error {
		attempts++
		stats.TotalAttempts++

		err := fn()
		if err == nil {
			return nil
		}

		stats.LastError = err

		// Check if we've exceeded max attempts
		if attempts >= config.MaxAttempts {
			return backoff.Permanent(fmt.Errorf("max retry attempts (%d) exceeded", config.MaxAttempts))
		}

		// Check if error is retryable
		if !IsRetryableError(err) {
			return backoff.Permanent(err)
		}

		return err
	}

	// Execute with backoff
	err := backoff.RetryNotify(retryableFunc, backoffContext, func(err error, delay time.Duration) {
		log.Infof("Retrying operation %s in %v (attempt %d/%d)", operation, delay, stats.TotalAttempts, config.MaxAttempts)
	})

	stats.Duration = time.Since(startTime)
	stats.Successful = err == nil

	if err != nil {
		log.Errorf("Operation %s failed after %d attempts: %v", operation, stats.TotalAttempts, err)
	} else {
		log.Infof("Operation %s succeeded after %d attempts", operation, stats.TotalAttempts)
	}

	return stats, err
}