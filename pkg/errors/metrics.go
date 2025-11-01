package errors
import (
	"fmt"
	"sync"
	"time"
)
type ErrorMetrics struct {
	mu sync.RWMutex
	errorCounts map[string]int
	errorRates map[string]*RateCounter
	errorPatterns map[string]*PatternTracker
}
type RateCounter struct {
	count     int
	window    time.Duration
	lastReset time.Time
}
type PatternTracker struct {
	occurrences []time.Time
	maxSize     int
}
func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		errorCounts:   make(map[string]int),
		errorRates:    make(map[string]*RateCounter),
		errorPatterns: make(map[string]*PatternTracker),
	}
}
func (em *ErrorMetrics) TrackError(err error) {
	if err == nil {
		return
	}
	em.mu.Lock()
	defer em.mu.Unlock()
	errorType := fmt.Sprintf("%T", err)
	em.errorCounts[errorType]++
	if _, exists := em.errorRates[errorType]; !exists {
		em.errorRates[errorType] = &RateCounter{
			window:    time.Minute,
			lastReset: time.Now(),
		}
	}
	em.errorRates[errorType].count++
	if _, exists := em.errorPatterns[errorType]; !exists {
		em.errorPatterns[errorType] = &PatternTracker{
			maxSize: 100,
		}
	}
	pattern := em.errorPatterns[errorType]
	pattern.occurrences = append(pattern.occurrences, time.Now())
	if len(pattern.occurrences) > pattern.maxSize {
		pattern.occurrences = pattern.occurrences[1:]
	}
}
func (em *ErrorMetrics) GetErrorRate(errorType string) float64 {
	em.mu.RLock()
	defer em.mu.RUnlock()
	if counter, exists := em.errorRates[errorType]; exists {
		elapsed := time.Since(counter.lastReset).Seconds()
		if elapsed == 0 {
			return 0
		}
		return float64(counter.count) / elapsed
	}
	return 0
}
func (em *ErrorMetrics) DetectErrorBurst(errorType string, threshold float64) bool {
	rate := em.GetErrorRate(errorType)
	return rate > threshold
}
func (em *ErrorMetrics) GetTopErrors(limit int) []ErrorStat {
	em.mu.RLock()
	defer em.mu.RUnlock()
	stats := make([]ErrorStat, 0, len(em.errorCounts))
	for errType, count := range em.errorCounts {
		stats = append(stats, ErrorStat{
			Type:  errType,
			Count: count,
			Rate:  em.GetErrorRate(errType),
		})
	}
	if len(stats) > limit {
		stats = stats[:limit]
	}
	return stats
}
type ErrorStat struct {
	Type  string
	Count int
	Rate  float64
}
func (em *ErrorMetrics) Reset(errorType string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	delete(em.errorCounts, errorType)
	delete(em.errorRates, errorType)
	delete(em.errorPatterns, errorType)
}
func (em *ErrorMetrics) GetPatternAnalysis(errorType string) *PatternAnalysis {
	em.mu.RLock()
	defer em.mu.RUnlock()
	pattern, exists := em.errorPatterns[errorType]
	if !exists || len(pattern.occurrences) == 0 {
		return nil
	}
	timestamps := pattern.occurrences
	intervals := make([]float64, len(timestamps)-1)
	for i := 1; i < len(timestamps); i++ {
		intervals[i-1] = timestamps[i].Sub(timestamps[i-1]).Seconds()
	}
	return &PatternAnalysis{
		ErrorType:        errorType,
		TotalOccurrences: len(timestamps),
		FirstOccurrence:  timestamps[0],
		LastOccurrence:   timestamps[len(timestamps)-1],
		AverageInterval:  calculateAverage(intervals),
		IsPeriodic:       detectPeriodicity(intervals),
	}
}
type PatternAnalysis struct {
	ErrorType        string
	TotalOccurrences int
	FirstOccurrence  time.Time
	LastOccurrence   time.Time
	AverageInterval  float64
	IsPeriodic       bool
}
func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
func detectPeriodicity(intervals []float64) bool {
	if len(intervals) < 3 {
		return false
	}
	variance := calculateVariance(intervals)
	return variance < 0.1
}
func calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := calculateAverage(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(values))
}