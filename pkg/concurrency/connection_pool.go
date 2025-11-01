package concurrency
import (
	"context"
	"fmt"
	"sync"
	"time"
)
type ConnectionPool struct {
	mu          sync.RWMutex
	connections chan interface{}
	factory     func() (interface{}, error)
	close       func(interface{}) error
	timeout     time.Duration
	maxSize     int
	inUse       int
}
func NewConnectionPool(factory func() (interface{}, error),
	close func(interface{}) error,
	initialSize int,
	maxSize int,
	timeout time.Duration) (*ConnectionPool, error) {
	if initialSize > maxSize {
		return nil, fmt.Errorf("initial size cannot be greater than max size")
	}
	pool := &ConnectionPool{
		connections: make(chan interface{}, maxSize),
		factory:     factory,
		close:       close,
		timeout:     timeout,
		maxSize:     maxSize,
	}
	for i := 0; i < initialSize; i++ {
		conn, err := factory()
		if err != nil {
			return nil, fmt.Errorf("failed to create initial connection: %w", err)
		}
		pool.connections <- conn
	}
	return pool, nil
}
func (p *ConnectionPool) Acquire(ctx context.Context) (interface{}, error) {
	done := make(chan interface{})
	var err error
	go func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.inUse >= p.maxSize {
			err = fmt.Errorf("connection pool exhausted")
			return
		}
		select {
		case conn := <-p.connections:
			p.inUse++
			done <- conn
			return
		default:
		}
		conn, createErr := p.factory()
		if createErr != nil {
			err = fmt.Errorf("failed to create new connection: %w", createErr)
			return
		}
		p.inUse++
		done <- conn
	}()
	select {
	case conn := <-done:
		return conn, err
	case <-ctx.Done():
		return nil, fmt.Errorf("operation cancelled: %w", ctx.Err())
	case <-time.After(p.timeout):
		return nil, fmt.Errorf("operation timed out")
	}
}
func (p *ConnectionPool) Release(ctx context.Context, conn interface{}) error {
	done := make(chan struct{})
	var err error
	go func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.inUse--
		select {
		case p.connections <- conn:
			close(done)
			return
		default:
			if closeErr := p.close(conn); closeErr != nil {
				err = fmt.Errorf("failed to close connection: %w", closeErr)
				return
			}
		}
		close(done)
	}()
	select {
	case <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	case <-time.After(p.timeout):
		return fmt.Errorf("operation timed out")
	}
}
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.connections)
	for conn := range p.connections {
		if err := p.close(conn); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}
	return nil
}
func (p *ConnectionPool) GetStats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return PoolStats{
		InUse:     p.inUse,
		Available: len(p.connections),
		MaxSize:   p.maxSize,
	}
}
type PoolStats struct {
	InUse     int
	Available int
	MaxSize   int
}