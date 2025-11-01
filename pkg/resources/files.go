package resources
import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)
type FileManager struct {
	mu       sync.RWMutex
	files    map[string]*ManagedFile
	pool     *ResourcePool
	maxFiles int
}
type ManagedFile struct {
	Path        string
	Handle      *os.File
	Size        int64
	LastAccess  time.Time
	AccessCount int64
	Locked      bool
}
func NewFileManager(pool *ResourcePool, maxFiles int) *FileManager {
	return &FileManager{
		files:    make(map[string]*ManagedFile),
		pool:     pool,
		maxFiles: maxFiles,
	}
}
func (fm *FileManager) OpenFile(ctx context.Context, path string) (*os.File, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	if mf, exists := fm.files[path]; exists {
		mf.LastAccess = time.Now()
		mf.AccessCount++
		return mf.Handle, nil
	}
	if len(fm.files) >= fm.maxFiles {
		fm.closeLeastRecentlyUsed()
		if len(fm.files) >= fm.maxFiles {
			return nil, fmt.Errorf("too many open files")
		}
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	err = fm.pool.Allocate(ctx, path, ResourceTypeFile, uint64(info.Size()), func() error {
		return file.Close()
	})
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to allocate resource: %w", err)
	}
	mf := &ManagedFile{
		Path:       path,
		Handle:     file,
		Size:       info.Size(),
		LastAccess: time.Now(),
	}
	fm.files[path] = mf
	return file, nil
}
func (fm *FileManager) CloseFile(ctx context.Context, path string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	mf, exists := fm.files[path]
	if !exists {
		return fmt.Errorf("file not found: %s", path)
	}
	if mf.Locked {
		return fmt.Errorf("file is locked: %s", path)
	}
	if err := fm.pool.Release(ctx, path); err != nil {
		return fmt.Errorf("failed to release resource: %w", err)
	}
	delete(fm.files, path)
	return nil
}
func (fm *FileManager) closeLeastRecentlyUsed() {
	type fileAccess struct {
		path       string
		lastAccess time.Time
	}
	var files []fileAccess
	for path, mf := range fm.files {
		if !mf.Locked {
			files = append(files, fileAccess{path, mf.LastAccess})
		}
	}
	for i := 0; i < len(files)/2; i++ {
		if err := fm.CloseFile(context.Background(), files[i].path); err != nil {
			fmt.Printf("Error closing file %s: %v\n", files[i].path, err)
		}
	}
}
func (fm *FileManager) LockFile(path string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	mf, exists := fm.files[path]
	if !exists {
		return fmt.Errorf("file not found: %s", path)
	}
	mf.Locked = true
	return nil
}
func (fm *FileManager) UnlockFile(path string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	mf, exists := fm.files[path]
	if !exists {
		return fmt.Errorf("file not found: %s", path)
	}
	mf.Locked = false
	return nil
}
func (fm *FileManager) GetFileStats() map[string]FileStats {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	stats := make(map[string]FileStats)
	for path, mf := range fm.files {
		stats[path] = FileStats{
			Size:        mf.Size,
			AccessCount: mf.AccessCount,
			LastAccess:  mf.LastAccess,
			Locked:      mf.Locked,
		}
	}
	return stats
}
type FileStats struct {
	Size        int64
	AccessCount int64
	LastAccess  time.Time
	Locked      bool
}