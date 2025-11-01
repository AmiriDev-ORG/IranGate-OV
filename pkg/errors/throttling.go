package errors
import (
	"fmt"
	"sync"
	"time"
)
type ErrorThrottler struct {
	mu          sync.RWMutex
	windowSize  time.Duration
	maxErrors   int
	errorCounts map[string][]time.Time
	metrics     *ErrorMetrics
}
func NewErrorThrottler(windowSize time.Duration, maxErrors int) *ErrorThrottler {
	return &ErrorThrottler{
		windowSize:  windowSize,
		maxErrors:   maxErrors,
		errorCounts: make(map[string][]time.Time),
		metrics:     NewErrorMetrics(),
	}
}
func (t *ErrorThrottler) RecordError(key string, err error) {
	now := time.Now()
	windowStart := now.Add(-t.windowSize)
	t.mu.Lock()
	defer t.mu.Unlock()
	times := t.errorCounts[key]
	var newTimes []time.Time
	if times != nil {
		newTimes = make([]time.Time, 0, len(times)+1)
		for _, ts := range times {
			if ts.After(windowStart) {
				newTimes = append(newTimes, ts)
			}
		}
	} else {
		newTimes = make([]time.Time, 0, 1)
	}
	newTimes = append(newTimes, now)
	t.errorCounts[key] = newTimes
	if err != nil {
		t.metrics.TrackError(err)
	}
	newTimes = append(newTimes, now)
	t.errorCounts[key] = newTimes
	t.mu.Unlock()
}
func (t *ErrorThrottler) ShouldThrottle(key string) bool {
	now := time.Now()
	windowStart := now.Add(-t.windowSize)
	t.mu.Lock()
	count := 0
	if times, exists := t.errorCounts[key]; exists {
		newTimes := make([]time.Time, 0, len(times))
		for _, ts := range times {
			if ts.After(windowStart) {
				newTimes = append(newTimes, ts)
				count++
			}
		}
		if len(newTimes) != len(times) {
			t.errorCounts[key] = newTimes
		}
	}
	t.mu.Unlock()
	return count >= t.maxErrors
}
func (t *ErrorThrottler) RecordErrorSimple(key string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	t.errorCounts[key] = append(t.errorCounts[key], now)
	t.metrics.TrackError(err)
}
func (t *ErrorThrottler) GetMetrics() *ErrorMetrics {
	return t.metrics
}
type ErrorFilter struct {
	filters []ErrorFilterFunc
}
type ErrorFilterFunc func(error) bool
func NewErrorFilter(filters ...ErrorFilterFunc) *ErrorFilter {
	return &ErrorFilter{
		filters: filters,
	}
}
func (f *ErrorFilter) AddFilter(filter ErrorFilterFunc) {
	f.filters = append(f.filters, filter)
}
func (f *ErrorFilter) ShouldHandle(err error) bool {
	for _, filter := range f.filters {
		if !filter(err) {
			return false
		}
	}
	return true
}
var (
	FilterTemporaryErrors = func(err error) bool {
		_, isTemp := err.(*TemporaryError)
		return !isTemp
	}
	FilterTimeoutErrors = func(err error) bool {
		_, isTimeout := err.(*TimeoutError)
		return !isTimeout
	}
	FilterNonCritical = func(err error) bool {
		if critErr, ok := err.(interface{ IsCritical() bool }); ok {
			return critErr.IsCritical()
		}
		return true
	}
)
type ErrorAggregator struct {
	mu         sync.RWMutex
	aggregates map[string]*ErrorAggregate
	window     time.Duration
	threshold  int
}
type ErrorAggregate struct {
	Count       int
	FirstSeen   time.Time
	LastSeen    time.Time
	Occurrences []time.Time
	Sample      error
}
func NewErrorAggregator(window time.Duration, threshold int) *ErrorAggregator {
	return &ErrorAggregator{
		aggregates: make(map[string]*ErrorAggregate),
		window:     window,
		threshold:  threshold,
	}
}
func (a *ErrorAggregator) AddError(err error) {
	if err == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	key := fmt.Sprintf("%T", err)
	now := time.Now()
	if agg, exists := a.aggregates[key]; exists {
		agg.Count++
		agg.LastSeen = now
		agg.Occurrences = append(agg.Occurrences, now)
	} else {
		a.aggregates[key] = &ErrorAggregate{
			Count:       1,
			FirstSeen:   now,
			LastSeen:    now,
			Occurrences: []time.Time{now},
			Sample:      err,
		}
	}
	a.cleanup(now)
}
func (a *ErrorAggregator) GetAggregates() map[string]*ErrorAggregate {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make(map[string]*ErrorAggregate)
	for k, v := range a.aggregates {
		result[k] = v
	}
	return result
}
func (a *ErrorAggregator) cleanup(now time.Time) {
	windowStart := now.Add(-a.window)
	for key, agg := range a.aggregates {
		if agg.LastSeen.Before(windowStart) {
			delete(a.aggregates, key)
		}
	}
}