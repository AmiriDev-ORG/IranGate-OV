package errors
import (
	"context"
	"fmt"
	"time"
)
type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	Timeout         time.Duration
}
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:     3,
	InitialInterval: 100 * time.Millisecond,
	MaxInterval:     10 * time.Second,
	Multiplier:      2.0,
	Timeout:         30 * time.Second,
}
type RetryableFunc func(ctx context.Context) error
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	switch err.(type) {
	case *TemporaryError:
		return true
	case *ConnectionError:
		return true
	case *TimeoutError:
		return true
	default:
		return false
	}
}
func WithRetry(ctx context.Context, fn RetryableFunc, config RetryConfig) error {
	var lastErr error
	attempt := 0
	interval := config.InitialInterval
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}
	for {
		if attempt >= config.MaxAttempts {
			return fmt.Errorf("max retry attempts (%d) exceeded: %w", config.MaxAttempts, lastErr)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry operation cancelled: %w", ctx.Err())
		default:
		}
		err := fn(ctx)
		if err == nil {
			return nil
		}
		lastErr = err
		if !IsRetryable(err) {
			return fmt.Errorf("non-retryable error encountered: %w", err)
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("retry operation cancelled during wait: %w", ctx.Err())
		case <-timer.C:
		}
		interval = time.Duration(float64(interval) * config.Multiplier)
		if interval > config.MaxInterval {
			interval = config.MaxInterval
		}
		attempt++
		LogRetryAttempt(ctx, attempt, err)
	}
}
type CircuitBreaker struct {
	failures     int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
	state        CircuitState
}
type CircuitState int
const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}
func (cb *CircuitBreaker) Execute(ctx context.Context, fn RetryableFunc) error {
	if !cb.allowRequest() {
		return &CircuitOpenError{message: "circuit breaker is open"}
	}
	err := fn(ctx)
	cb.recordResult(err)
	return err
}
func (cb *CircuitBreaker) allowRequest() bool {
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}
func (cb *CircuitBreaker) recordResult(err error) {
	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
	} else {
		cb.failures = 0
		cb.state = StateClosed
	}
}
func LogRetryAttempt(ctx context.Context, attempt int, err error) {
	logger := GetLogger(ctx)
	logger.Warn(fmt.Sprintf("Retry attempt %d failed: %v", attempt, err))
}