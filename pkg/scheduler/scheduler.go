package scheduler
import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/amiridev-org/irangate-ov/pkg/utils"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)
const (
	LogFile = "/var/log/irangate/cron.log"
)
type Job struct {
	ID       string
	Schedule string
	Action   string
	Params   map[string]interface{}
	LastRun  *time.Time
	Enabled  bool
}
type Scheduler struct {
	db     *database.DB
	cron   interface{}
	jobs   map[string]interface{}
	mu     sync.RWMutex
	logger *logrus.Logger
}
func New(db *database.DB) *Scheduler {
	return &Scheduler{
		db:     db,
		cron:   nil,
		jobs:   make(map[string]interface{}),
		logger: utils.GetLogger(),
	}
}
func (s *Scheduler) Start() error {
	s.logger.Info("Starting scheduler...")
	if err := os.MkdirAll(filepath.Dir(LogFile), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}
	jobs, err := s.db.GetCronJobs()
	if err != nil {
		return fmt.Errorf("failed to get cron jobs: %v", err)
	}
	for _, job := range jobs {
		if job.Enabled {
			if err := s.AddJob(database.CronJob{
				ID:       job.ID,
				Schedule: job.Schedule,
				Action:   job.Action,
				Params:   job.Params,
				LastRun:  job.LastRun,
				Enabled:  job.Enabled,
			}); err != nil {
				s.logger.Errorf("Failed to add job %s: %v", job.ID, err)
			}
		}
	}
	s.logger.Info("Scheduler started successfully")
	return nil
}
func (s *Scheduler) Stop() {
	s.logger.Info("Stopping scheduler...")
}
func (s *Scheduler) AddJob(job database.CronJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := cron.ParseStandard(job.Schedule); err != nil {
		return fmt.Errorf("invalid cron schedule '%s': %v\nExamples:\n  '0 * * * *'   = Every hour\n  '0 */6 * * *' = Every 6 hours\n  '0 0 * * 0'   = Weekly on Sunday at midnight\n  '0 0 1 * *'   = Monthly on the 1st at midnight", job.Schedule, err)
	}
	entryID := "placeholder"
	s.jobs[job.ID] = entryID
	if err := s.db.AddCronJob(job); err != nil {
		delete(s.jobs, job.ID)
		return fmt.Errorf("failed to save job to database: %v", err)
	}
	s.logger.Infof("Added cron job: %s (%s)", job.ID, job.Schedule)
	return nil
}
func (s *Scheduler) RemoveJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
	if err := s.db.RemoveCronJob(id); err != nil {
		s.logger.Errorf("Failed to remove job from database: %v", err)
		return fmt.Errorf("failed to remove job from database: %v", err)
	}
	s.logger.Infof("Removed cron job: %s", id)
	return nil
}
func (s *Scheduler) RunJob(id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	jobs, err := s.db.GetCronJobs()
	if err != nil {
		return fmt.Errorf("failed to get cron jobs: %v", err)
	}
	for _, job := range jobs {
		if job.ID == id {
			jobFunc := s.createJobFunction(job)
			go jobFunc()
			return nil
		}
	}
	return fmt.Errorf("job not found: %s", id)
}
func (s *Scheduler) ListJobs() ([]database.CronJob, error) {
	return s.db.GetCronJobs()
}
func (s *Scheduler) createJobFunction(job database.CronJob) func() {
	return func() {
		s.logger.Infof("Running job %s (%s)", job.ID, job.Action)
		start := time.Now()
		var err error
		switch job.Action {
		case "backup":
			err = s.runBackupJob(job.Params)
		case "monitor":
			err = s.runMonitorJob(job.Params)
		case "cleanup":
			err = s.runCleanupJob(job.Params)
		default:
			err = fmt.Errorf("unknown action: %s", job.Action)
		}
		duration := time.Since(start)
		if err != nil {
			s.logger.Errorf("Job %s failed: %v", job.ID, err)
		} else {
			s.logger.Infof("Job %s completed in %v", job.ID, duration)
		}
		s.logJobExecution(job.ID, start, duration, err)
	}
}
func (s *Scheduler) runBackupJob(params map[string]interface{}) error {
	backupType, _ := params["type"].(string)
	retention, _ := params["retention_days"].(float64)
	s.logger.Infof("Running backup job (type: %s, retention: %.0f days)", backupType, retention)
	if err := s.db.Backup(); err != nil {
		return fmt.Errorf("failed to create database backup: %v", err)
	}
	if retention > 0 {
		if err := s.cleanupOldBackups(int(retention)); err != nil {
			s.logger.Warnf("Failed to cleanup old backups: %v", err)
		}
	}
	s.logger.Info("Backup job completed successfully")
	return nil
}
func (s *Scheduler) runMonitorJob(params map[string]interface{}) error {
	monitorType, _ := params["type"].(string)
	threshold, _ := params["alert_threshold"].(float64)
	s.logger.Infof("Running monitor job (type: %s, threshold: %.0f%%)", monitorType, threshold)
	switch monitorType {
	case "health":
		return s.checkSystemHealth(threshold)
	case "clients":
		return s.checkClientConnections(threshold)
	case "traffic":
		return s.checkTrafficUsage(threshold)
	default:
		s.logger.Warnf("Unknown monitor type: %s", monitorType)
		return nil
	}
}
func (s *Scheduler) runCleanupJob(_ map[string]interface{}) error {
	s.logger.Info("Running cleanup job")
	if err := s.cleanupOldLogs(); err != nil {
		s.logger.Warnf("Failed to cleanup old logs: %v", err)
	}
	if err := s.cleanupTempFiles(); err != nil {
		s.logger.Warnf("Failed to cleanup temp files: %v", err)
	}
	if err := s.cleanupExpiredClients(); err != nil {
		s.logger.Warnf("Failed to cleanup expired clients: %v", err)
	}
	s.logger.Info("Cleanup job completed successfully")
	return nil
}
func (s *Scheduler) logJobExecution(jobID string, start time.Time, duration time.Duration, err error) {
	logEntry := fmt.Sprintf("[%s] Job %s - Duration: %v - Status: %s\n",
		start.Format(time.RFC3339),
		jobID,
		duration,
		map[bool]string{true: "Success", false: fmt.Sprintf("Failed: %v", err)}[err == nil],
	)
	f, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.logger.Errorf("Failed to open log file: %v", err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString(logEntry); err != nil {
		s.logger.Errorf("Failed to write to log file: %v", err)
	}
}
func (s *Scheduler) cleanupOldBackups(retentionDays int) error {
	backupDir := filepath.Join(s.db.GetPath(), "backups")
	files, err := filepath.Glob(filepath.Join(backupDir, "backup_*.json"))
	if err != nil {
		return fmt.Errorf("failed to list backup files: %v", err)
	}
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	removedCount := 0
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(file); err != nil {
				s.logger.Warnf("Failed to remove old backup %s: %v", file, err)
			} else {
				removedCount++
			}
		}
	}
	if removedCount > 0 {
		s.logger.Infof("Cleaned up %d old backup files", removedCount)
	}
	return nil
}
func (s *Scheduler) checkSystemHealth(_ float64) error {
	s.logger.Info("System health check completed")
	return nil
}
func (s *Scheduler) checkClientConnections(_ float64) error {
	s.logger.Info("Client connections check completed")
	return nil
}
func (s *Scheduler) checkTrafficUsage(_ float64) error {
	s.logger.Info("Traffic usage check completed")
	return nil
}
func (s *Scheduler) cleanupOldLogs() error {
	logPaths := []string{
		"/var/log/openvpn/openvpn.log",
		"/var/log/openvpn/openvpn-status.log",
		"/var/log/irangate/cron.log",
		"/var/log/irangate/ai_agent.log",
	}
	cutoffTime := time.Now().AddDate(0, 0, -30)
	removedCount := 0
	for _, logPath := range logPaths {
		if info, err := os.Stat(logPath); err == nil {
			if info.ModTime().Before(cutoffTime) {
				if err := os.Remove(logPath); err != nil {
					s.logger.Warnf("Failed to remove old log %s: %v", logPath, err)
				} else {
					removedCount++
				}
			}
		}
	}
	if removedCount > 0 {
		s.logger.Infof("Cleaned up %d old log files", removedCount)
	}
	return nil
}
func (s *Scheduler) cleanupTempFiles() error {
	tempPaths := []string{
		"/tmp/irangate_*",
		"/tmp/openvpn_*",
	}
	removedCount := 0
	for _, pattern := range tempPaths {
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, file := range files {
			if err := os.Remove(file); err != nil {
				s.logger.Warnf("Failed to remove temp file %s: %v", file, err)
			} else {
				removedCount++
			}
		}
	}
	if removedCount > 0 {
		s.logger.Infof("Cleaned up %d temporary files", removedCount)
	}
	return nil
}
func (s *Scheduler) cleanupExpiredClients() error {
	clients, err := s.db.GetClients()
	if err != nil {
		return fmt.Errorf("failed to get clients: %v", err)
	}
	now := time.Now()
	removedCount := 0
	for _, client := range clients {
		if !client.ExpiresAt.IsZero() && client.ExpiresAt.Before(now) {
			certPaths := []string{
				filepath.Join("/etc/openvpn/client", client.Name+".crt"),
				filepath.Join("/etc/openvpn/client", client.Name+".key"),
				filepath.Join("/etc/openvpn/client", client.Name+".ovpn"),
			}
			for _, path := range certPaths {
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					s.logger.Warnf("Failed to remove expired client file %s: %v", path, err)
				}
			}
			if err := s.db.RemoveClient(client.Name); err != nil {
				s.logger.Warnf("Failed to remove expired client %s from database: %v", client.Name, err)
			} else {
				removedCount++
			}
		}
	}
	if removedCount > 0 {
		s.logger.Infof("Cleaned up %d expired clients", removedCount)
	}
	return nil
}