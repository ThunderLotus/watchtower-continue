package retry

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, uint64(5), config.MaxAttempts)
	assert.Equal(t, 1*time.Second, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
	assert.Equal(t, 0.1, config.Randomization)
	assert.True(t, config.EnableRetry)
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "retryable error",
			err:      NewRetryableError(errors.New("temporary error")),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      NewNonRetryableError(errors.New("permanent error")),
			expected: false,
		},
		{
			name:     "network timeout",
			err:      &net.OpError{Err: errors.New("timeout"), Op: "read"},
			expected: true,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithRetry_Success(t *testing.T) {
	config := DefaultConfig()
	config.MaxAttempts = 3
	config.InitialDelay = 10 * time.Millisecond
	config.MaxDelay = 50 * time.Millisecond

	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	stats, err := WithRetry(context.Background(), config, "test_operation", fn)
	require.NoError(t, err)
	assert.True(t, stats.Successful)
	assert.Equal(t, 2, stats.TotalAttempts)
	assert.Equal(t, "test_operation", stats.Operation)
}

func TestWithRetry_AllAttemptsFailed(t *testing.T) {
	config := DefaultConfig()
	config.MaxAttempts = 3
	config.InitialDelay = 10 * time.Millisecond
	config.MaxDelay = 50 * time.Millisecond

	attempts := 0
	fn := func() error {
		attempts++
		return errors.New("temporary error")
	}

	stats, err := WithRetry(context.Background(), config, "test_operation", fn)
	assert.Error(t, err)
	assert.False(t, stats.Successful)
	assert.GreaterOrEqual(t, stats.TotalAttempts, 1)
	assert.Equal(t, "test_operation", stats.Operation)
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	config := DefaultConfig()
	config.MaxAttempts = 3
	config.InitialDelay = 10 * time.Millisecond
	config.MaxDelay = 50 * time.Millisecond

	attempts := 0
	fn := func() error {
		attempts++
		return NewNonRetryableError(errors.New("permanent error"))
	}

	stats, err := WithRetry(context.Background(), config, "test_operation", fn)
	assert.Error(t, err)
	assert.False(t, stats.Successful)
	assert.Equal(t, 1, stats.TotalAttempts) // Should only attempt once
	assert.Equal(t, "test_operation", stats.Operation)
}

func TestWithRetry_Disabled(t *testing.T) {
	config := DefaultConfig()
	config.EnableRetry = false

	attempts := 0
	fn := func() error {
		attempts++
		return errors.New("error")
	}

	stats, err := WithRetry(context.Background(), config, "test_operation", fn)
	assert.Error(t, err)
	assert.False(t, stats.Successful)
	assert.Equal(t, 1, stats.TotalAttempts)
	assert.Equal(t, 1, attempts)
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	config := DefaultConfig()
	config.MaxAttempts = 10
	config.InitialDelay = 10 * time.Millisecond
	config.MaxDelay = 50 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	fn := func() error {
		attempts++
		if attempts == 2 {
			cancel() // Cancel context on second attempt
		}
		return errors.New("error")
	}

	stats, err := WithRetry(ctx, config, "test_operation", fn)
	assert.Error(t, err)
	assert.False(t, stats.Successful)
	assert.GreaterOrEqual(t, stats.TotalAttempts, 1)
}

// Removed TestWithRetryResult_Success and TestWithRetryResult_Failure
// as WithRetryResult function was removed for simplification

// Removed TestWithRetryCustom
// as WithRetryCustom function was removed for simplification

func TestNewRetryableError(t *testing.T) {
	err := errors.New("test error")
	retryableErr := NewRetryableError(err)
	
	assert.Error(t, retryableErr)
	assert.Contains(t, retryableErr.Error(), "retryable error")
	
	var rErr *RetryableError
	assert.True(t, errors.As(retryableErr, &rErr))
	assert.Equal(t, err, rErr.Err)
}

func TestNewNonRetryableError(t *testing.T) {
	err := errors.New("test error")
	nonRetryableErr := NewNonRetryableError(err)
	
	assert.Error(t, nonRetryableErr)
	assert.Contains(t, nonRetryableErr.Error(), "non-retryable error")
	
	var nrErr *NonRetryableError
	assert.True(t, errors.As(nonRetryableErr, &nrErr))
	assert.Equal(t, err, nrErr.Err)
}

func TestStats(t *testing.T) {
	stats := Stats{
		TotalAttempts: 5,
		Successful:    true,
		Duration:      10 * time.Second,
		LastError:     errors.New("last error"),
		Operation:     "test_op",
	}

	assert.Equal(t, 5, stats.TotalAttempts)
	assert.True(t, stats.Successful)
	assert.Equal(t, 10*time.Second, stats.Duration)
	assert.Error(t, stats.LastError)
	assert.Equal(t, "test_op", stats.Operation)
}

func TestIsTemporaryError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "network error with timeout",
			err:      &net.OpError{Err: errors.New("i/o timeout"), Op: "read"},
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "connection reset",
			err:      errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("some permanent error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTemporaryError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkWithRetry_Success(b *testing.B) {
	config := DefaultConfig()
	config.MaxAttempts = 3
	config.InitialDelay = 10 * time.Millisecond
	config.MaxDelay = 50 * time.Millisecond

	fn := func() error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = WithRetry(context.Background(), config, "bench_op", fn)
	}
}

func BenchmarkWithRetry_WithRetries(b *testing.B) {
	config := DefaultConfig()
	config.MaxAttempts = 3
	config.InitialDelay = 10 * time.Millisecond
	config.MaxDelay = 50 * time.Millisecond

	attempts := 0
	fn := func() error {
		attempts++
		if attempts%3 != 0 {
			return errors.New("temporary error")
		}
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = WithRetry(context.Background(), config, "bench_op", fn)
	}
}