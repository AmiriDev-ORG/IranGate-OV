package concurrency
import (
	"context"
	"fmt"
	"sync"
	"time"
)
type ResourceManager struct {
	mu          sync.RWMutex
	resources   map[string]*Resource
	timeout     time.Duration
	maxRetries  int
	maxLockTime time.Duration
}
type Resource struct {
	mu       sync.RWMutex
	locked   bool
	lockedAt time.Time
	lockedBy string
	waitList []chan struct{}
}
func NewResourceManager(timeout time.Duration, maxRetries int, maxLockTime time.Duration) *ResourceManager {
	return &ResourceManager{
		resources:   make(map[string]*Resource),
		timeout:     timeout,
		maxRetries:  maxRetries,
		maxLockTime: maxLockTime,
	}
}
func (rm *ResourceManager) AcquireResource(ctx context.Context, resourceID string, requestorID string) error {
	for attempt := 0; attempt < rm.maxRetries; attempt++ {
		if err := rm.tryAcquireResource(ctx, resourceID, requestorID); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		case <-time.After(time.Second * time.Duration(attempt+1)):
			continue
		}
	}
	return fmt.Errorf("failed to acquire resource after %d attempts", rm.maxRetries)
}
func (rm *ResourceManager) tryAcquireResource(ctx context.Context, resourceID string, requestorID string) error {
	done := make(chan struct{})
	var err error
	go func() {
		rm.mu.Lock()
		resource, exists := rm.resources[resourceID]
		if !exists {
			resource = &Resource{}
			rm.resources[resourceID] = resource
		}
		rm.mu.Unlock()
		if !resource.mu.TryLock() {
			err = fmt.Errorf("lock acquisition timeout")
			return
		}
		defer resource.mu.Unlock()
		if resource.locked {
			if time.Since(resource.lockedAt) > rm.maxLockTime {
				resource.locked = false
			} else {
				err = fmt.Errorf("resource already locked by %s", resource.lockedBy)
				return
			}
		}
		resource.locked = true
		resource.lockedAt = time.Now()
		resource.lockedBy = requestorID
		close(done)
	}()
	select {
	case <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	case <-time.After(rm.timeout):
		return fmt.Errorf("operation timed out")
	}
}
func (rm *ResourceManager) ReleaseResource(ctx context.Context, resourceID string, requestorID string) error {
	done := make(chan struct{})
	var err error
	go func() {
		rm.mu.RLock()
		resource, exists := rm.resources[resourceID]
		rm.mu.RUnlock()
		if !exists {
			err = fmt.Errorf("resource not found")
			return
		}
		if !resource.mu.TryLock() {
			err = fmt.Errorf("lock acquisition timeout")
			return
		}
		defer resource.mu.Unlock()
		if !resource.locked {
			err = fmt.Errorf("resource not locked")
			return
		}
		if resource.lockedBy != requestorID {
			err = fmt.Errorf("resource locked by different requestor")
			return
		}
		resource.locked = false
		resource.lockedBy = ""
		for _, ch := range resource.waitList {
			close(ch)
		}
		resource.waitList = nil
		close(done)
	}()
	select {
	case <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	case <-time.After(rm.timeout):
		return fmt.Errorf("operation timed out")
	}
}
func (rm *ResourceManager) WaitForResource(ctx context.Context, resourceID string) error {
	done := make(chan struct{})
	var err error
	go func() {
		rm.mu.RLock()
		resource, exists := rm.resources[resourceID]
		rm.mu.RUnlock()
		if !exists {
			err = fmt.Errorf("resource not found")
			return
		}
		if !resource.mu.TryRLock() {
			err = fmt.Errorf("lock acquisition timeout")
			return
		}
		defer resource.mu.RUnlock()
		if !resource.locked {
			close(done)
			return
		}
		wait := make(chan struct{})
		resource.waitList = append(resource.waitList, wait)
		<-wait
		close(done)
	}()
	select {
	case <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	case <-time.After(rm.timeout):
		return fmt.Errorf("operation timed out")
	}
}
func (rm *ResourceManager) GetResourceStatus(resourceID string) (bool, string, error) {
	rm.mu.RLock()
	resource, exists := rm.resources[resourceID]
	rm.mu.RUnlock()
	if !exists {
		return false, "", fmt.Errorf("resource not found")
	}
	resource.mu.RLock()
	defer resource.mu.RUnlock()
	return resource.locked, resource.lockedBy, nil
}