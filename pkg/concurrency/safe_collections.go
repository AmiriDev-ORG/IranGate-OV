package concurrency
import (
	"context"
	"fmt"
	"sync"
	"time"
)
type SafeCounter struct {
	mu      sync.RWMutex
	count   int64
	timeout time.Duration
}
func NewSafeCounter(timeout time.Duration) *SafeCounter {
	return &SafeCounter{
		timeout: timeout,
	}
}
func (sc *SafeCounter) Increment(ctx context.Context) error {
	done := make(chan struct{})
	var err error
	go func() {
		if !sc.mu.TryLock() {
			err = fmt.Errorf("lock acquisition timeout")
			return
		}
		defer sc.mu.Unlock()
		sc.count++
		close(done)
	}()
	select {
	case <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	case <-time.After(sc.timeout):
		return fmt.Errorf("operation timed out")
	}
}
func (sc *SafeCounter) GetCount(ctx context.Context) (int64, error) {
	done := make(chan int64)
	var err error
	go func() {
		if !sc.mu.TryRLock() {
			err = fmt.Errorf("lock acquisition timeout")
			return
		}
		defer sc.mu.RUnlock()
		done <- sc.count
	}()
	select {
	case count := <-done:
		return count, err
	case <-ctx.Done():
		return 0, fmt.Errorf("operation cancelled: %w", ctx.Err())
	case <-time.After(sc.timeout):
		return 0, fmt.Errorf("operation timed out")
	}
}
type SafeMap struct {
	mu      sync.RWMutex
	data    map[string]interface{}
	timeout time.Duration
}
func NewSafeMap(timeout time.Duration) *SafeMap {
	return &SafeMap{
		data:    make(map[string]interface{}),
		timeout: timeout,
	}
}
func (sm *SafeMap) Set(ctx context.Context, key string, value interface{}) error {
	done := make(chan struct{})
	var err error
	go func() {
		if !sm.mu.TryLock() {
			err = fmt.Errorf("lock acquisition timeout")
			return
		}
		defer sm.mu.Unlock()
		sm.data[key] = value
		close(done)
	}()
	select {
	case <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	case <-time.After(sm.timeout):
		return fmt.Errorf("operation timed out")
	}
}
func (sm *SafeMap) Get(ctx context.Context, key string) (interface{}, bool, error) {
	done := make(chan struct{})
	var (
		value interface{}
		ok    bool
		err   error
	)
	go func() {
		if !sm.mu.TryRLock() {
			err = fmt.Errorf("lock acquisition timeout")
			return
		}
		defer sm.mu.RUnlock()
		value, ok = sm.data[key]
		close(done)
	}()
	select {
	case <-done:
		return value, ok, err
	case <-ctx.Done():
		return nil, false, fmt.Errorf("operation cancelled: %w", ctx.Err())
	case <-time.After(sm.timeout):
		return nil, false, fmt.Errorf("operation timed out")
	}
}
func (sm *SafeMap) Delete(ctx context.Context, key string) error {
	done := make(chan struct{})
	var err error
	go func() {
		if !sm.mu.TryLock() {
			err = fmt.Errorf("lock acquisition timeout")
			return
		}
		defer sm.mu.Unlock()
		delete(sm.data, key)
		close(done)
	}()
	select {
	case <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	case <-time.After(sm.timeout):
		return fmt.Errorf("operation timed out")
	}
}