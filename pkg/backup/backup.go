package backup
import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)
type BackupConfig struct {
	BackupDir    string        `json:"backup_dir"`
	Interval     time.Duration `json:"interval"`
	MaxBackups   int           `json:"max_backups"`
	IncludeFiles []string      `json:"include_files"`
}
type BackupManager struct {
	config    BackupConfig
	stopChan  chan struct{}
	isRunning bool
}
type BackupMetadata struct {
	Timestamp time.Time         `json:"timestamp"`
	Files     []string          `json:"files"`
	Size      int64             `json:"size"`
	Version   string            `json:"version"`
	Checksums map[string]string `json:"checksums"`
}
func New(config BackupConfig) (*BackupManager, error) {
	if config.BackupDir == "" {
		return nil, fmt.Errorf("backup directory is required")
	}
	if config.Interval == 0 {
		config.Interval = 24 * time.Hour
	}
	if config.MaxBackups == 0 {
		config.MaxBackups = 7
	}
	if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %v", err)
	}
	return &BackupManager{
		config:   config,
		stopChan: make(chan struct{}),
	}, nil
}
func (m *BackupManager) Start() error {
	if m.isRunning {
		return fmt.Errorf("backup manager is already running")
	}
	m.isRunning = true
	ticker := time.NewTicker(m.config.Interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := m.CreateBackup(); err != nil {
					fmt.Printf("Backup failed: %v\n", err)
				}
			case <-m.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
	return nil
}
func (m *BackupManager) Stop() {
	if m.isRunning {
		close(m.stopChan)
		m.isRunning = false
	}
}
func (m *BackupManager) CreateBackup() error {
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(m.config.BackupDir, fmt.Sprintf("backup_%s.zip", timestamp))
	zipfile, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	defer zipfile.Close()
	archive := zip.NewWriter(zipfile)
	defer archive.Close()
	metadata := BackupMetadata{
		Timestamp: time.Now(),
		Files:     make([]string, 0),
		Version:   "1.0",
		Checksums: make(map[string]string),
	}
	for _, file := range m.config.IncludeFiles {
		if err := m.addFileToZip(archive, file, &metadata); err != nil {
			return err
		}
	}
	metadataJson, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	metadataWriter, err := archive.Create("metadata.json")
	if err != nil {
		return err
	}
	if _, err := metadataWriter.Write(metadataJson); err != nil {
		return err
	}
	return m.cleanupOldBackups()
}
func (m *BackupManager) RestoreBackup(backupFile string) error {
	reader, err := zip.OpenReader(backupFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	var metadata BackupMetadata
	for _, file := range reader.File {
		if file.Name == "metadata.json" {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			if err := json.NewDecoder(rc).Decode(&metadata); err != nil {
				return err
			}
			break
		}
	}
	if len(metadata.Checksums) > 0 {
		fmt.Println("Verifying backup integrity...")
		for _, file := range reader.File {
			if file.Name == "metadata.json" {
				continue
			}
			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open %s for validation: %v", file.Name, err)
			}
			hash := sha256.New()
			if _, err := io.Copy(hash, rc); err != nil {
				rc.Close()
				return fmt.Errorf("failed to calculate checksum for %s: %v", file.Name, err)
			}
			rc.Close()
			actualHash := fmt.Sprintf("%x", hash.Sum(nil))
			expectedHash, exists := metadata.Checksums[file.Name]
			if exists && actualHash != expectedHash {
				return fmt.Errorf("integrity check FAILED for %s: checksum mismatch (backup may be corrupted)", file.Name)
			}
		}
		fmt.Println("✓ Backup integrity verified successfully")
	}
	for _, file := range reader.File {
		if file.Name == "metadata.json" {
			continue
		}
		if err := m.restoreFile(file); err != nil {
			return err
		}
	}
	return nil
}
func (m *BackupManager) addFileToZip(archive *zip.Writer, file string, metadata *BackupMetadata) error {
	checksum, err := calculateFileSHA256(file)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum for %s: %v", file, err)
	}
	src, err := os.Open(file)
	if err != nil {
		return err
	}
	defer src.Close()
	info, err := src.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate
	header.Name = filepath.Base(file)
	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	if _, err := io.Copy(writer, src); err != nil {
		return err
	}
	metadata.Files = append(metadata.Files, file)
	metadata.Size += info.Size()
	metadata.Checksums[header.Name] = checksum
	return nil
}
func calculateFileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
func (m *BackupManager) restoreFile(file *zip.File) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	path := filepath.Join(m.config.BackupDir, "..", file.Name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, rc)
	return err
}
func (m *BackupManager) cleanupOldBackups() error {
	pattern := filepath.Join(m.config.BackupDir, "backup_*.zip")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(files) <= m.config.MaxBackups {
		return nil
	}
	type fileInfo struct {
		path    string
		modTime time.Time
	}
	fileInfos := make([]fileInfo, 0, len(files))
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{file, info.ModTime()})
	}
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime.Before(fileInfos[j].modTime)
	})
	for i := 0; i < len(fileInfos)-m.config.MaxBackups; i++ {
		if err := os.Remove(fileInfos[i].path); err != nil {
			fmt.Printf("Failed to remove old backup %s: %v\n", fileInfos[i].path, err)
		}
	}
	return nil
}
func (m *BackupManager) ListBackups() ([]BackupMetadata, error) {
	pattern := filepath.Join(m.config.BackupDir, "backup_*.zip")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	backups := make([]BackupMetadata, 0, len(files))
	for _, file := range files {
		reader, err := zip.OpenReader(file)
		if err != nil {
			continue
		}
		for _, f := range reader.File {
			if f.Name == "metadata.json" {
				rc, err := f.Open()
				if err != nil {
					continue
				}
				var metadata BackupMetadata
				if err := json.NewDecoder(rc).Decode(&metadata); err != nil {
					rc.Close()
					continue
				}
				rc.Close()
				backups = append(backups, metadata)
				break
			}
		}
		reader.Close()
	}
	return backups, nil
}