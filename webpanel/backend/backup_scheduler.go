package main
import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
)
type BackupSchedule struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	CronExpression string     `json:"cron_expression"`
	Enabled        bool       `json:"enabled"`
	RetentionDays  int        `json:"retention_days"`
	BackupPath     string     `json:"backup_path"`
	IncludePaths   []string   `json:"include_paths"`
	LastRun        *time.Time `json:"last_run,omitempty"`
	NextRun        *time.Time `json:"next_run,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
type BackupRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	CronExpression string   `json:"cron_expression"`
	Enabled        bool     `json:"enabled"`
	RetentionDays  int      `json:"retention_days"`
	BackupPath     string   `json:"backup_path"`
	IncludePaths   []string `json:"include_paths"`
}
type BackupFile struct {
	Filename   string    `json:"filename"`
	Size       int64     `json:"size"`
	CreatedAt  time.Time `json:"created_at"`
	BackupType string    `json:"backup_type"`
	ScheduleID string    `json:"schedule_id"`
}
type BackupExecution struct {
	ID          string    `json:"id"`
	ScheduleID  string    `json:"schedule_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	FilesBacked int       `json:"files_backed"`
	TotalSize   int64     `json:"total_size"`
	Duration    int64     `json:"duration"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}
const BackupSchedulerDataDir = "/opt/irangate/webpanel/backups"
var cronScheduler *cron.Cron
func getBackupSchedulesHandler(w http.ResponseWriter, r *http.Request) {
	schedules, err := loadAllBackupSchedules()
	if err != nil {
		sendError(w, "Failed to load backup schedules", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Backup schedules retrieved successfully", schedules)
}
func createBackupScheduleHandler(w http.ResponseWriter, r *http.Request) {
	var req BackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.CronExpression == "" {
		sendError(w, "Name and cron expression are required", http.StatusBadRequest)
		return
	}
	if _, err := cron.ParseStandard(req.CronExpression); err != nil {
		sendError(w, fmt.Sprintf("Invalid cron expression: %v", err), http.StatusBadRequest)
		return
	}
	if req.RetentionDays <= 0 {
		req.RetentionDays = 30
	}
	if req.BackupPath == "" {
		req.BackupPath = filepath.Join(BackupSchedulerDataDir, "backups")
	}
	if len(req.IncludePaths) == 0 {
		req.IncludePaths = getDefaultBackupPaths()
	}
	schedule := BackupSchedule{
		ID:             generateScheduleID(req.Name),
		Name:           req.Name,
		Description:    req.Description,
		CronExpression: req.CronExpression,
		Enabled:        req.Enabled,
		RetentionDays:  req.RetentionDays,
		BackupPath:     req.BackupPath,
		IncludePaths:   req.IncludePaths,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := saveBackupSchedule(&schedule); err != nil {
		sendError(w, "Failed to save backup schedule", http.StatusInternalServerError)
		return
	}
	if schedule.Enabled {
		scheduleBackup(&schedule)
	}
	sendSuccess(w, "Backup schedule created successfully", schedule)
}
func updateBackupScheduleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduleID := vars["id"]
	var req BackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	schedule, err := loadBackupSchedule(scheduleID)
	if err != nil {
		sendError(w, "Backup schedule not found", http.StatusNotFound)
		return
	}
	if req.CronExpression != "" {
		if _, err := cron.ParseStandard(req.CronExpression); err != nil {
			sendError(w, fmt.Sprintf("Invalid cron expression: %v", err), http.StatusBadRequest)
			return
		}
	}
	if req.Name != "" {
		schedule.Name = req.Name
	}
	if req.Description != "" {
		schedule.Description = req.Description
	}
	if req.CronExpression != "" {
		schedule.CronExpression = req.CronExpression
	}
	if req.BackupPath != "" {
		schedule.BackupPath = req.BackupPath
	}
	if req.IncludePaths != nil {
		schedule.IncludePaths = req.IncludePaths
	}
	if req.RetentionDays > 0 {
		schedule.RetentionDays = req.RetentionDays
	}
	schedule.Enabled = req.Enabled
	schedule.UpdatedAt = time.Now()
	if err := saveBackupSchedule(schedule); err != nil {
		sendError(w, "Failed to update backup schedule", http.StatusInternalServerError)
		return
	}
	if schedule.Enabled {
		scheduleBackup(schedule)
	}
	sendSuccess(w, "Backup schedule updated successfully", schedule)
}
func deleteBackupScheduleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduleID := vars["id"]
	schedule, err := loadBackupSchedule(scheduleID)
	if err != nil {
		sendError(w, "Backup schedule not found", http.StatusNotFound)
		return
	}
	removeScheduleFromCron(schedule.ID)
	scheduleFile := filepath.Join(BackupSchedulerDataDir, "schedules", scheduleID+".json")
	if err := os.Remove(scheduleFile); err != nil {
		sendError(w, "Failed to delete backup schedule", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Backup schedule deleted successfully", map[string]interface{}{
		"schedule_id": scheduleID,
	})
}
func runBackupScheduleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduleID := vars["id"]
	schedule, err := loadBackupSchedule(scheduleID)
	if err != nil {
		sendError(w, "Backup schedule not found", http.StatusNotFound)
		return
	}
	go func() {
		executeBackup(schedule)
	}()
	sendSuccess(w, "Backup execution started", map[string]interface{}{
		"schedule_id": scheduleID,
		"status":      "started",
	})
}
func listBackupsHandler(w http.ResponseWriter, r *http.Request) {
	scheduleID := r.URL.Query().Get("schedule_id")
	backups, err := listBackupFiles(scheduleID)
	if err != nil {
		sendError(w, "Failed to list backup files", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Backup files retrieved successfully", backups)
}
func deleteBackupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		sendError(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	backupPath := filepath.Join(BackupSchedulerDataDir, "backups", filename)
	if err := os.Remove(backupPath); err != nil {
		sendError(w, "Failed to delete backup file", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Backup file deleted successfully", map[string]interface{}{
		"filename": filename,
	})
}
func downloadBackupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		sendError(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	backupPath := filepath.Join(BackupSchedulerDataDir, "backups", filename)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		sendError(w, "Backup file not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	http.ServeFile(w, r, backupPath)
}
func generateScheduleID(name string) string {
	return fmt.Sprintf("schedule_%s_%d", sanitizeFilename(name), time.Now().Unix())
}
func getDefaultBackupPaths() []string {
	return []string{
		"/opt/irangate/database",
		"/etc/openvpn",
		"/opt/irangate/clients",
	}
}
func saveBackupSchedule(schedule *BackupSchedule) error {
	schedulesDir := filepath.Join(BackupSchedulerDataDir, "schedules")
	if err := os.MkdirAll(schedulesDir, 0755); err != nil {
		return err
	}
	scheduleFile := filepath.Join(schedulesDir, schedule.ID+".json")
	data, err := json.MarshalIndent(schedule, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(scheduleFile, data, 0644)
}
func loadBackupSchedule(scheduleID string) (*BackupSchedule, error) {
	scheduleFile := filepath.Join(BackupSchedulerDataDir, "schedules", scheduleID+".json")
	data, err := os.ReadFile(scheduleFile)
	if err != nil {
		return nil, err
	}
	var schedule BackupSchedule
	if err := json.Unmarshal(data, &schedule); err != nil {
		return nil, err
	}
	return &schedule, nil
}
func loadAllBackupSchedules() ([]*BackupSchedule, error) {
	schedulesDir := filepath.Join(BackupSchedulerDataDir, "schedules")
	if err := os.MkdirAll(schedulesDir, 0755); err != nil {
		return nil, err
	}
	pattern := filepath.Join(schedulesDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	var schedules []*BackupSchedule
	for _, file := range files {
		schedule, err := loadBackupScheduleFromFile(file)
		if err != nil {
			continue
		}
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}
func loadBackupScheduleFromFile(filename string) (*BackupSchedule, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var schedule BackupSchedule
	if err := json.Unmarshal(data, &schedule); err != nil {
		return nil, err
	}
	return &schedule, nil
}
func scheduleBackup(schedule *BackupSchedule) {
	removeScheduleFromCron(schedule.ID)
	_, err := cronScheduler.AddFunc(schedule.CronExpression, func() {
		executeBackup(schedule)
	})
	if err != nil {
		fmt.Printf("Failed to schedule backup %s: %v\n", schedule.ID, err)
	}
}
func removeScheduleFromCron(scheduleID string) {
}
func executeBackup(schedule *BackupSchedule) {
	execution := BackupExecution{
		ID:         fmt.Sprintf("exec_%d", time.Now().Unix()),
		ScheduleID: schedule.ID,
		StartedAt:  time.Now(),
		Status:     "running",
	}
	defer func() {
		execution.CompletedAt = time.Now()
		execution.Duration = execution.CompletedAt.Sub(execution.StartedAt).Milliseconds()
		saveBackupExecution(&execution)
		now := time.Now()
		schedule.LastRun = &now
		saveBackupSchedule(schedule)
	}()
	backupDir := filepath.Join(BackupSchedulerDataDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		execution.Status = "failed"
		execution.Message = fmt.Sprintf("Failed to create backup directory: %v", err)
		return
	}
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s_%s.zip", schedule.ID, timestamp)
	backupPath := filepath.Join(backupDir, filename)
	if err := createBackupArchive(backupPath, schedule.IncludePaths); err != nil {
		execution.Status = "failed"
		execution.Message = fmt.Sprintf("Failed to create backup archive: %v", err)
		return
	}
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		execution.Status = "failed"
		execution.Message = fmt.Sprintf("Failed to get backup file info: %v", err)
		return
	}
	execution.Status = "success"
	execution.Message = fmt.Sprintf("Backup completed successfully: %s", filename)
	execution.TotalSize = fileInfo.Size()
	execution.FilesBacked = len(schedule.IncludePaths)
	cleanupOldBackups(schedule)
}
func createBackupArchive(archivePath string, includePaths []string) error {
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()
	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()
	for _, path := range includePaths {
		if err := addPathToArchive(zipWriter, path, ""); err != nil {
			fmt.Printf("Warning: Failed to add path %s to archive: %v\n", path, err)
		}
	}
	return nil
}
func addPathToArchive(zipWriter *zip.Writer, sourcePath, basePath string) error {
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		if basePath != "" {
			relPath = filepath.Join(basePath, relPath)
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}
		_, err = writer.Write([]byte{})
		if err != nil {
			return err
		}
		return nil
	})
}
func cleanupOldBackups(schedule *BackupSchedule) {
	backupDir := filepath.Join(BackupSchedulerDataDir, "backups")
	pattern := filepath.Join(backupDir, fmt.Sprintf("backup_%s_*.zip", schedule.ID))
	files, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			info1, err1 := os.Stat(files[i])
			info2, err2 := os.Stat(files[j])
			if err1 != nil || err2 != nil {
				continue
			}
			if info1.ModTime().After(info2.ModTime()) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
	cutoff := time.Now().AddDate(0, 0, -schedule.RetentionDays)
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(file)
		}
	}
}
func listBackupFiles(scheduleID string) ([]*BackupFile, error) {
	backupDir := filepath.Join(BackupSchedulerDataDir, "backups")
	var pattern string
	if scheduleID != "" {
		pattern = filepath.Join(backupDir, fmt.Sprintf("backup_%s_*.zip", scheduleID))
	} else {
		pattern = filepath.Join(backupDir, "backup_*.zip")
	}
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	var backupFiles []*BackupFile
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		filename := filepath.Base(file)
		parts := strings.Split(filename, "_")
		var extractedScheduleID string
		if len(parts) >= 3 {
			extractedScheduleID = parts[1]
		}
		backupFile := &BackupFile{
			Filename:   filename,
			Size:       info.Size(),
			CreatedAt:  info.ModTime(),
			BackupType: "scheduled",
			ScheduleID: extractedScheduleID,
		}
		backupFiles = append(backupFiles, backupFile)
	}
	return backupFiles, nil
}
func saveBackupExecution(execution *BackupExecution) error {
	executionsFile := filepath.Join(BackupSchedulerDataDir, "executions.json")
	var executions []*BackupExecution
	if data, err := os.ReadFile(executionsFile); err == nil {
		json.Unmarshal(data, &executions)
	}
	executions = append(executions, execution)
	if len(executions) > 1000 {
		executions = executions[len(executions)-1000:]
	}
	data, _ := json.MarshalIndent(executions, "", "  ")
	return os.WriteFile(executionsFile, data, 0644)
}