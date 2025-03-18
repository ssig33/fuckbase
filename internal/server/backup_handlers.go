package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ssig33/fuckbase/internal/logger"
)

// handleBackupCreate handles the /backup/create endpoint
func (s *Server) handleBackupCreate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		logRequest(r, start, http.StatusOK)
	}()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Check admin authentication if enabled
	if s.adminAuth.Config != nil && s.adminAuth.Config.Enabled {
		s.adminAuth.RequireAdminAuth(func(w http.ResponseWriter, r *http.Request) {
			s.handleBackupCreateImpl(w, r)
		})(w, r)
		return
	}

	// No admin auth required
	s.handleBackupCreateImpl(w, r)
}

// handleBackupCreateImpl implements the backup creation logic
func (s *Server) handleBackupCreateImpl(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req CreateBackupRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Check if S3 is enabled
	if s.backupManager == nil {
		writeErrorResponse(w, http.StatusServiceUnavailable, "S3_NOT_ENABLED", "S3 backup is not enabled")
		return
	}

	var backupErr error
	var message string

	// If database is specified, backup only that database
	if req.Database != "" {
		// Check if database exists
		if !s.DBManager.DatabaseExists(req.Database) {
			writeErrorResponse(w, http.StatusNotFound, "DB_NOT_FOUND", "Database not found")
			return
		}

		// Backup the database
		backupErr = s.backupManager.BackupDatabase(req.Database)
		message = "Database backed up successfully"
		logger.Info("Backing up database: %s", req.Database)
	} else {
		// Backup all databases
		backupErr = s.backupManager.BackupAllDatabases()
		message = "All databases backed up successfully"
		logger.Info("Backing up all databases")
	}

	if backupErr != nil {
		logger.Error("Backup failed: %v", backupErr)
		writeErrorResponse(w, http.StatusInternalServerError, "BACKUP_FAILED", "Failed to create backup: "+backupErr.Error())
		return
	}

	// Return success response
	response := Response{
		Status:  "success",
		Message: message,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleBackupList handles the /backup/list endpoint
func (s *Server) handleBackupList(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		logRequest(r, start, http.StatusOK)
	}()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Check admin authentication if enabled
	if s.adminAuth.Config != nil && s.adminAuth.Config.Enabled {
		s.adminAuth.RequireAdminAuth(func(w http.ResponseWriter, r *http.Request) {
			s.handleBackupListImpl(w, r)
		})(w, r)
		return
	}

	// No admin auth required
	s.handleBackupListImpl(w, r)
}

// handleBackupListImpl implements the backup listing logic
func (s *Server) handleBackupListImpl(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req ListBackupsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Check if S3 is enabled
	if s.backupManager == nil {
		writeErrorResponse(w, http.StatusServiceUnavailable, "S3_NOT_ENABLED", "S3 backup is not enabled")
		return
	}

	var backups []string
	var listErr error

	// If database is specified, list backups for that database
	if req.Database != "" {
		// Check if database exists
		if !s.DBManager.DatabaseExists(req.Database) {
			writeErrorResponse(w, http.StatusNotFound, "DB_NOT_FOUND", "Database not found")
			return
		}

		// List backups for the database
		backups, listErr = s.backupManager.ListDatabaseBackups(req.Database)
		logger.Info("Listing backups for database: %s", req.Database)
	} else {
		// List all backups
		backups, listErr = s.backupManager.ListBackups()
		logger.Info("Listing all backups")
	}

	if listErr != nil {
		logger.Error("Failed to list backups: %v", listErr)
		writeErrorResponse(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list backups: "+listErr.Error())
		return
	}

	// Convert to backup info objects
	backupInfos := make([]BackupInfo, 0, len(backups))
	for _, backup := range backups {
		// Get file info
		info, err := s.s3Client.GetFileInfo(backup)
		if err != nil {
			logger.Error("Failed to get info for backup %s: %v", backup, err)
			continue
		}

		// Parse database name from path
		dbName := ""
		if strings.HasPrefix(backup, "backups/full/") {
			dbName = "all"
		} else {
			parts := strings.Split(backup, "/")
			if len(parts) >= 2 {
				dbName = parts[1]
			}
		}

		// Parse timestamp from filename
		timestamp := time.Now() // Default to now if parsing fails
		parts := strings.Split(backup, "/")
		if len(parts) > 0 {
			fileName := parts[len(parts)-1]
			if len(fileName) > 15 { // Assuming format: YYYYMMDD-HHMMSS.json
				dateStr := strings.TrimSuffix(fileName, ".json")
				if t, err := time.Parse("20060102-150405", dateStr); err == nil {
					timestamp = t
				}
			}
		}

		backupInfos = append(backupInfos, BackupInfo{
			Name:      backup,
			Timestamp: timestamp,
			Size:      info.Size,
			Database:  dbName,
		})
	}

	// Return success response
	response := ListBackupsResponse{
		Status:  "success",
		Backups: backupInfos,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleBackupRestore handles the /backup/restore endpoint
func (s *Server) handleBackupRestore(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		logRequest(r, start, http.StatusOK)
	}()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Check admin authentication if enabled
	if s.adminAuth.Config != nil && s.adminAuth.Config.Enabled {
		s.adminAuth.RequireAdminAuth(func(w http.ResponseWriter, r *http.Request) {
			s.handleBackupRestoreImpl(w, r)
		})(w, r)
		return
	}

	// No admin auth required
	s.handleBackupRestoreImpl(w, r)
}

// handleBackupRestoreImpl implements the backup restore logic
func (s *Server) handleBackupRestoreImpl(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req RestoreBackupRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Validate request
	if req.BackupName == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Backup name is required")
		return
	}

	// Check if S3 is enabled
	if s.backupManager == nil {
		writeErrorResponse(w, http.StatusServiceUnavailable, "S3_NOT_ENABLED", "S3 backup is not enabled")
		return
	}

	var restoreErr error
	var message string

	// Check if it's a full backup or a single database backup
	if strings.HasPrefix(req.BackupName, "backups/full/") {
		// Restore all databases
		restoreErr = s.backupManager.RestoreAllDatabases(req.BackupName)
		message = "All databases restored successfully"
		logger.Info("Restoring all databases from backup: %s", req.BackupName)
	} else {
		// Restore single database
		restoreErr = s.backupManager.RestoreDatabase(req.BackupName)
		message = "Database restored successfully"
		logger.Info("Restoring database from backup: %s", req.BackupName)
	}

	if restoreErr != nil {
		logger.Error("Restore failed: %v", restoreErr)
		writeErrorResponse(w, http.StatusInternalServerError, "RESTORE_FAILED", "Failed to restore backup: "+restoreErr.Error())
		return
	}

	// Return success response
	response := Response{
		Status:  "success",
		Message: message,
	}
	writeJSONResponse(w, http.StatusOK, response)
}