package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ssig33/fuckbase/internal/config"
	"github.com/ssig33/fuckbase/internal/database"
	"github.com/ssig33/fuckbase/internal/logger"
	"github.com/ssig33/fuckbase/internal/s3"
)

// Server represents the HTTP server for FuckBase
type Server struct {
	Config         *config.ServerConfig
	DBManager      *database.Manager
	httpServer     *http.Server
	adminAuth      *AdminAuth
	s3Client       *s3.Client
	backupManager  *s3.BackupManager
	backupTicker   *time.Ticker
	stopBackupChan chan struct{}
}

// NewServer creates a new server with the given configuration and database manager
func NewServer(cfg *config.ServerConfig, dbManager *database.Manager) *Server {
	server := &Server{
		Config:         cfg,
		DBManager:      dbManager,
		adminAuth:      NewAdminAuth(cfg.AdminAuth),
		stopBackupChan: make(chan struct{}),
	}

	// Initialize S3 client if enabled
	if cfg.S3Config != nil && cfg.S3Config.Enabled {
		var err error
		server.s3Client, err = s3.NewClient(cfg.S3Config)
		if err != nil {
			logger.Error("Failed to initialize S3 client: %v", err)
		} else {
			logger.Info("S3 client initialized successfully")
			server.backupManager = s3.NewBackupManager(server.s3Client, dbManager)
		}
	}

	return server
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create a new router
	router := http.NewServeMux()

	// Register API endpoints
	s.registerEndpoints(router)

	// Create the HTTP server
	addr := fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start automatic backups if S3 is enabled
	if s.backupManager != nil && s.Config.BackupInterval > 0 {
		s.startAutomaticBackups()
	}

	// Start the server
	logger.Info("Starting server on %s", addr)
	return s.httpServer.ListenAndServe()
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	logger.Info("Stopping server")
	
	// Stop automatic backups
	if s.backupTicker != nil {
		s.stopBackupChan <- struct{}{}
		s.backupTicker.Stop()
	}
	
	return s.httpServer.Shutdown(ctx)
}

// startAutomaticBackups starts a ticker to perform automatic backups
func (s *Server) startAutomaticBackups() {
	interval := time.Duration(s.Config.BackupInterval) * time.Minute
	s.backupTicker = time.NewTicker(interval)
	
	go func() {
		logger.Info("Starting automatic backups every %d minutes", s.Config.BackupInterval)
		for {
			select {
			case <-s.backupTicker.C:
				logger.Info("Running scheduled backup")
				if err := s.backupManager.BackupAllDatabases(); err != nil {
					logger.Error("Scheduled backup failed: %v", err)
				}
			case <-s.stopBackupChan:
				logger.Info("Stopping automatic backups")
				return
			}
		}
	}()
}

// registerEndpoints registers all API endpoints
func (s *Server) registerEndpoints(router *http.ServeMux) {
	// Database management endpoints
	router.HandleFunc("/create", s.handleDatabaseCreate)
	router.HandleFunc("/drop", s.handleDatabaseDrop)

	// Set operations
	router.HandleFunc("/set/create", s.handleSetCreate)
	router.HandleFunc("/set/get", s.handleSetGet)
	router.HandleFunc("/set/put", s.handleSetPut)
	router.HandleFunc("/set/delete", s.handleSetDelete)
	router.HandleFunc("/set/list", s.handleSetList)

	// Index operations
	router.HandleFunc("/index/create", s.handleIndexCreate)
	router.HandleFunc("/index/drop", s.handleIndexDrop)
	router.HandleFunc("/index/query", s.handleIndexQuery)

	// Server info
	router.HandleFunc("/server/info", s.handleServerInfo)
	
	// Backup and restore operations (only if S3 is enabled)
	if s.backupManager != nil {
		router.HandleFunc("/backup/create", s.handleBackupCreate)
		router.HandleFunc("/backup/list", s.handleBackupList)
		router.HandleFunc("/backup/restore", s.handleBackupRestore)
	}
}

// logRequest logs information about an HTTP request
func logRequest(r *http.Request, start time.Time, statusCode int) {
	duration := time.Since(start)
	logger.Info("%s %s %d %s", r.Method, r.URL.Path, statusCode, duration)
}

// writeJSONResponse writes a JSON response with the given status code
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.Error("Failed to encode JSON response: %v", err)
		}
	}
}

// writeErrorResponse writes an error response with the given status code and error message
func writeErrorResponse(w http.ResponseWriter, statusCode int, code string, message string) {
	response := ErrorResponse{
		Status:  "error",
		Code:    code,
		Message: message,
	}
	writeJSONResponse(w, statusCode, response)
}