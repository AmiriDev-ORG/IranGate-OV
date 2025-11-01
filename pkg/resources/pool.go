package resources
import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)
type ResourcePool struct {
	mu              sync.RWMutex
	resources       map[string]*ManagedResource
	memoryLimit     uint64
	memoryUsage     uint64
	cleanupInterval time.Duration
	metrics         *ResourceMetrics
	shutdown        chan struct{}
}
type ManagedResource struct {
	ID           string
	Type         ResourceType
	CreatedAt    time.Time
	LastAccessed time.Time
	Size         uint64
	InUse        bool
	RefCount     int
	Cleanup      func() error
}
type ResourceType string
const (
	ResourceTypeMemory     ResourceType = "memory"
	ResourceTypeConnection ResourceType = "connection"
	ResourceTypeFile       ResourceType = "file"
	ResourceTypeHandle     ResourceType = "handle"
)
type ResourceMetrics struct {
	TotalAllocated uint64
	TotalFreed     uint64
	PeakUsage      uint64
	CurrentUsage   uint64
	LeakSuspects   map[string]*ManagedResource
}
func NewResourcePool(memoryLimit uint64, cleanupInterval time.Duration) *ResourcePool {
	pool := &ResourcePool{
		resources:       make(map[string]*ManagedResource),
		memoryLimit:     memoryLimit,
		cleanupInterval: cleanupInterval,
		metrics: &ResourceMetrics{
			LeakSuspects: make(map[string]*ManagedResource),
		},
		shutdown: make(chan struct{}),
	}
	go pool.backgroundCleanup()
	go pool.monitorMemory()
	return pool
}
func (p *ResourcePool) Allocate(ctx context.Context, id string, resType ResourceType, size uint64, cleanup func() error) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.memoryUsage+size > p.memoryLimit {
		p.emergencyCleanup()
		if p.memoryUsage+size > p.memoryLimit {
			return fmt.Errorf("memory limit exceeded")
		}
	}
	resource := &ManagedResource{
		ID:           id,
		Type:         resType,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		Size:         size,
		InUse:        true,
		RefCount:     1,
		Cleanup:      cleanup,
	}
	p.resources[id] = resource
	p.memoryUsage += size
	p.updateMetrics()
	return nil
}
func (p *ResourcePool) Release(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	resource, exists := p.resources[id]
	if !exists {
		return fmt.Errorf("resource not found: %s", id)
	}
	resource.RefCount--
	if resource.RefCount == 0 {
		resource.InUse = false
	}
	return nil
}
func (p *ResourcePool) backgroundCleanup() {
	ticker := time.NewTicker(p.cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.cleanup(false)
		case <-p.shutdown:
			p.cleanup(true)
			return
		}
	}
}
func (p *ResourcePool) cleanup(force bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, resource := range p.resources {
		if force || (!resource.InUse && time.Since(resource.LastAccessed) > p.cleanupInterval) {
			if err := resource.Cleanup(); err != nil {
				fmt.Printf("Error cleaning up resource %s: %v\n", id, err)
			}
			delete(p.resources, id)
			p.memoryUsage -= resource.Size
		}
	}
	p.updateMetrics()
}
func (p *ResourcePool) emergencyCleanup() {
	for id, resource := range p.resources {
		if !resource.InUse {
			if err := resource.Cleanup(); err != nil {
				fmt.Printf("Error in emergency cleanup of resource %s: %v\n", id, err)
			}
			delete(p.resources, id)
			p.memoryUsage -= resource.Size
		}
	}
	runtime.GC()
}
func (p *ResourcePool) monitorMemory() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.checkForLeaks()
		case <-p.shutdown:
			return
		}
	}
}
func (p *ResourcePool) checkForLeaks() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	threshold := 30 * time.Minute
	for id, resource := range p.resources {
		if resource.InUse && time.Since(resource.LastAccessed) > threshold {
			p.metrics.LeakSuspects[id] = resource
		}
	}
}
func (p *ResourcePool) updateMetrics() {
	p.metrics.CurrentUsage = p.memoryUsage
	if p.memoryUsage > p.metrics.PeakUsage {
		p.metrics.PeakUsage = p.memoryUsage
	}
}
func (p *ResourcePool) GetMetrics() *ResourceMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return &ResourceMetrics{
		TotalAllocated: p.metrics.TotalAllocated,
		TotalFreed:     p.metrics.TotalFreed,
		PeakUsage:      p.metrics.PeakUsage,
		CurrentUsage:   p.metrics.CurrentUsage,
		LeakSuspects:   p.metrics.LeakSuspects,
	}
}
func (p *ResourcePool) Shutdown(ctx context.Context) error {
	close(p.shutdown)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second * 5):
		p.cleanup(true)
	}
	return nil
}