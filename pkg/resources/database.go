package resources
import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)
type DBConnectionPool struct {
	mu          sync.RWMutex
	pool        *ResourcePool
	connections map[string]*ManagedConnection
	config      DBConfig
}
type DBConfig struct {
	MaxConnections int
	IdleTimeout    time.Duration
	MaxLifetime    time.Duration
	ConnectTimeout time.Duration
}
type ManagedConnection struct {
	ID               string
	Conn             *sql.DB
	CreatedAt        time.Time
	LastUsed         time.Time
	TransactionCount int64
	QueryCount       int64
	ErrorCount       int64
}
func NewDBConnectionPool(pool *ResourcePool, config DBConfig) *DBConnectionPool {
	return &DBConnectionPool{
		pool:        pool,
		connections: make(map[string]*ManagedConnection),
		config:      config,
	}
}
func (dp *DBConnectionPool) GetConnection(ctx context.Context, dsn string) (*sql.DB, error) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	if conn, exists := dp.connections[dsn]; exists {
		conn.LastUsed = time.Now()
		return conn.Conn, nil
	}
	if len(dp.connections) >= dp.config.MaxConnections {
		return nil, fmt.Errorf("connection limit reached")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db.SetConnMaxLifetime(dp.config.MaxLifetime)
	db.SetConnMaxIdleTime(dp.config.IdleTimeout)
	connID := fmt.Sprintf("db_%s_%d", dsn, time.Now().UnixNano())
	err = dp.pool.Allocate(ctx, connID, ResourceTypeConnection, 1024, func() error {
		return db.Close()
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to allocate resource: %w", err)
	}
	conn := &ManagedConnection{
		ID:        connID,
		Conn:      db,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}
	dp.connections[dsn] = conn
	return db, nil
}
func (dp *DBConnectionPool) ReleaseConnection(ctx context.Context, dsn string) error {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	conn, exists := dp.connections[dsn]
	if !exists {
		return fmt.Errorf("connection not found: %s", dsn)
	}
	if err := dp.pool.Release(ctx, conn.ID); err != nil {
		return fmt.Errorf("failed to release resource: %w", err)
	}
	delete(dp.connections, dsn)
	return nil
}
func (dp *DBConnectionPool) GetConnectionStats() map[string]ConnectionStats {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	stats := make(map[string]ConnectionStats)
	for dsn, conn := range dp.connections {
		stats[dsn] = ConnectionStats{
			CreatedAt:        conn.CreatedAt,
			LastUsed:         conn.LastUsed,
			TransactionCount: conn.TransactionCount,
			QueryCount:       conn.QueryCount,
			ErrorCount:       conn.ErrorCount,
		}
	}
	return stats
}
type ConnectionStats struct {
	CreatedAt        time.Time
	LastUsed         time.Time
	TransactionCount int64
	QueryCount       int64
	ErrorCount       int64
}
func (dp *DBConnectionPool) IncrementStats(dsn string, statsType string) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	if conn, exists := dp.connections[dsn]; exists {
		switch statsType {
		case "transaction":
			conn.TransactionCount++
		case "query":
			conn.QueryCount++
		case "error":
			conn.ErrorCount++
		}
	}
}
func (dp *DBConnectionPool) CleanupIdleConnections(ctx context.Context) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	now := time.Now()
	for dsn, conn := range dp.connections {
		if now.Sub(conn.LastUsed) > dp.config.IdleTimeout {
			if err := dp.ReleaseConnection(ctx, dsn); err != nil {
				fmt.Printf("Error releasing connection %s: %v\n", dsn, err)
			}
		}
	}
}