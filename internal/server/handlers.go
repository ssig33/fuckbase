package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ssig33/fuckbase/internal/database"
	"github.com/ssig33/fuckbase/internal/logger"
)

// handleDatabaseCreate handles the /create endpoint
func (s *Server) handleDatabaseCreate(w http.ResponseWriter, r *http.Request) {
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
			s.handleDatabaseCreateImpl(w, r)
		})(w, r)
		return
	}

	// No admin auth required
	s.handleDatabaseCreateImpl(w, r)
}

// handleDatabaseCreateImpl implements the database creation logic
func (s *Server) handleDatabaseCreateImpl(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req CreateDatabaseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Validate request
	if req.Name == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Database name is required")
		return
	}

	// Check if database already exists
	if s.DBManager.DatabaseExists(req.Name) {
		writeErrorResponse(w, http.StatusConflict, "DB_ALREADY_EXISTS", "Database already exists")
		return
	}

	// Create auth config if provided
	var authConfig *database.AuthConfig
	if req.Auth.Username != "" && req.Auth.Password != "" {
		authConfig = &database.AuthConfig{
			Username: req.Auth.Username,
			Password: req.Auth.Password,
			Enabled:  true,
		}
	}

	// Create database
	_, err = s.DBManager.CreateDatabase(req.Name, authConfig)
	if err != nil {
		logger.Error("Failed to create database: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create database")
		return
	}

	logger.Info("Created database: %s", req.Name)

	// Return success response
	response := Response{
		Status:  "success",
		Message: "Database created successfully",
		Data: map[string]string{
			"database": req.Name,
		},
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleDatabaseDrop handles the /drop endpoint
func (s *Server) handleDatabaseDrop(w http.ResponseWriter, r *http.Request) {
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
			s.handleDatabaseDropImpl(w, r)
		})(w, r)
		return
	}

	// No admin auth required
	s.handleDatabaseDropImpl(w, r)
}

// handleDatabaseDropImpl implements the database drop logic
func (s *Server) handleDatabaseDropImpl(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req DropDatabaseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Validate request
	if req.Name == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Database name is required")
		return
	}

	// Check if database exists
	if !s.DBManager.DatabaseExists(req.Name) {
		writeErrorResponse(w, http.StatusNotFound, "DB_NOT_FOUND", "Database not found")
		return
	}

	// Drop database
	err = s.DBManager.DeleteDatabase(req.Name)
	if err != nil {
		logger.Error("Failed to drop database: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to drop database")
		return
	}

	logger.Info("Dropped database: %s", req.Name)

	// Return success response
	response := Response{
		Status:  "success",
		Message: "Database dropped successfully",
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleSetCreate handles the /set/create endpoint
func (s *Server) handleSetCreate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		logRequest(r, start, http.StatusOK)
	}()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req CreateSetRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Validate request
	if req.Database == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Database name is required")
		return
	}
	if req.Name == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Set name is required")
		return
	}

	// Get database
	db, err := s.DBManager.GetDatabase(req.Database)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "DB_NOT_FOUND", "Database not found")
		return
	}

	// Check database authentication
	username, password, hasAuth := ExtractDatabaseAuth(r)
	if !hasAuth {
		username = req.Auth.Username
		password = req.Auth.Password
	}
	if !db.Authenticate(username, password) {
		writeErrorResponse(w, http.StatusUnauthorized, "AUTH_FAILED", "Authentication failed")
		return
	}

	// Create set
	set, err := db.CreateSet(req.Name)
	if err != nil {
		logger.Error("Failed to create set: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create set")
		return
	}

	logger.Info("Created set: %s in database: %s", req.Name, req.Database)

	// Return success response
	response := Response{
		Status:  "success",
		Message: "Set created successfully",
		Data: map[string]string{
			"set": set.Name,
		},
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleServerInfo handles the /server/info endpoint
func (s *Server) handleServerInfo(w http.ResponseWriter, r *http.Request) {
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
			s.handleServerInfoImpl(w, r)
		})(w, r)
		return
	}

	// No admin auth required
	s.handleServerInfoImpl(w, r)
}

// handleServerInfoImpl implements the server info logic
func (s *Server) handleServerInfoImpl(w http.ResponseWriter, r *http.Request) {
	// Calculate uptime
	uptime := time.Since(s.startTime)
	days := int(uptime.Hours()) / 24
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	uptimeStr := fmt.Sprintf("%dd %dh %dm", days, hours, minutes)

	// Create response
	response := ServerInfoResponse{
		Status:         "success",
		Version:        "0.0.1",
		Uptime:         uptimeStr,
		DatabasesCount: s.DBManager.GetDatabaseCount(),
	}

	logger.Info("Server info requested")
	writeJSONResponse(w, http.StatusOK, response)
}

// handleSetGet handles the /set/get endpoint
func (s *Server) handleSetGet(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		logRequest(r, start, http.StatusOK)
	}()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req GetSetRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Validate request
	if req.Database == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Database name is required")
		return
	}
	if req.Set == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Set name is required")
		return
	}
	if req.Key == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Key is required")
		return
	}

	// Get database
	db, err := s.DBManager.GetDatabase(req.Database)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "DB_NOT_FOUND", "Database not found")
		return
	}

	// Check database authentication
	username, password, hasAuth := ExtractDatabaseAuth(r)
	if !hasAuth {
		username = req.Auth.Username
		password = req.Auth.Password
	}
	if !db.Authenticate(username, password) {
		writeErrorResponse(w, http.StatusUnauthorized, "AUTH_FAILED", "Authentication failed")
		return
	}

	// Get set
	set, err := db.GetSet(req.Set)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "SET_NOT_FOUND", "Set not found")
		return
	}

	// Get value
	var value interface{}
	if err := set.Get(req.Key, &value); err != nil {
		writeErrorResponse(w, http.StatusNotFound, "KEY_NOT_FOUND", "Key not found")
		return
	}

	logger.Info("Retrieved value for key: %s from set: %s in database: %s", req.Key, req.Set, req.Database)

	// Return success response
	response := Response{
		Status: "success",
		Data:   value,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleSetPut handles the /set/put endpoint
func (s *Server) handleSetPut(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		logRequest(r, start, http.StatusOK)
	}()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req PutSetRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Validate request
	if req.Database == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Database name is required")
		return
	}
	if req.Set == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Set name is required")
		return
	}
	if req.Key == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Key is required")
		return
	}
	if len(req.Value) == 0 {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Value is required")
		return
	}

	// Get database
	db, err := s.DBManager.GetDatabase(req.Database)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "DB_NOT_FOUND", "Database not found")
		return
	}

	// Check database authentication
	username, password, hasAuth := ExtractDatabaseAuth(r)
	if !hasAuth {
		username = req.Auth.Username
		password = req.Auth.Password
	}
	if !db.Authenticate(username, password) {
		writeErrorResponse(w, http.StatusUnauthorized, "AUTH_FAILED", "Authentication failed")
		return
	}

	// Get set
	set, err := db.GetSet(req.Set)
	if err != nil {
		// If set doesn't exist, create it
		set, err = db.CreateSet(req.Set)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create set")
			return
		}
	}

	// Parse the value from JSON
	var value interface{}
	if err := json.Unmarshal(req.Value, &value); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse value as JSON")
		return
	}

	// Store the value
	if err := set.Put(req.Key, value); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to store value")
		return
	}

	// Update indexes if any
	for _, index := range db.Indexes {
		if index.SetName == req.Set {
			// Get the raw value
			rawValue, err := set.GetRaw(req.Key)
			if err != nil {
				logger.Error("Failed to get raw value for index update: %v", err)
				continue
			}

			// Update the index
			if err := index.AddEntry(req.Key, rawValue); err != nil {
				logger.Error("Failed to update index: %v", err)
			}
		}
	}

	logger.Info("Stored value for key: %s in set: %s in database: %s", req.Key, req.Set, req.Database)

	// Return success response
	response := Response{
		Status:  "success",
		Message: "Data stored successfully",
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleSetDelete handles the /set/delete endpoint
func (s *Server) handleSetDelete(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		logRequest(r, start, http.StatusOK)
	}()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req DeleteSetRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Validate request
	if req.Database == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Database name is required")
		return
	}
	if req.Set == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Set name is required")
		return
	}
	if req.Key == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Key is required")
		return
	}

	// Get database
	db, err := s.DBManager.GetDatabase(req.Database)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "DB_NOT_FOUND", "Database not found")
		return
	}

	// Check database authentication
	username, password, hasAuth := ExtractDatabaseAuth(r)
	if !hasAuth {
		username = req.Auth.Username
		password = req.Auth.Password
	}
	if !db.Authenticate(username, password) {
		writeErrorResponse(w, http.StatusUnauthorized, "AUTH_FAILED", "Authentication failed")
		return
	}

	// Get set
	set, err := db.GetSet(req.Set)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "SET_NOT_FOUND", "Set not found")
		return
	}

	// Get the raw value before deleting (for index updates)
	rawValue, err := set.GetRaw(req.Key)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "KEY_NOT_FOUND", "Key not found")
		return
	}

	// Delete the key
	if err := set.Delete(req.Key); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete key")
		return
	}

	// Update indexes if any
	for _, index := range db.Indexes {
		if index.SetName == req.Set {
			// Remove the entry from the index
			if err := index.RemoveEntry(req.Key, rawValue); err != nil {
				logger.Error("Failed to update index: %v", err)
			}
		}
	}

	logger.Info("Deleted key: %s from set: %s in database: %s", req.Key, req.Set, req.Database)

	// Return success response
	response := Response{
		Status:  "success",
		Message: "Data deleted successfully",
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleSetList handles the /set/list endpoint
func (s *Server) handleSetList(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		logRequest(r, start, http.StatusOK)
	}()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req ListSetsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Validate request
	if req.Database == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Database name is required")
		return
	}

	// Get database
	db, err := s.DBManager.GetDatabase(req.Database)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "DB_NOT_FOUND", "Database not found")
		return
	}

	// Check database authentication
	username, password, hasAuth := ExtractDatabaseAuth(r)
	if !hasAuth {
		username = req.Auth.Username
		password = req.Auth.Password
	}
	if !db.Authenticate(username, password) {
		writeErrorResponse(w, http.StatusUnauthorized, "AUTH_FAILED", "Authentication failed")
		return
	}

	// Get list of sets
	sets := db.ListSets()

	logger.Info("Listed sets in database: %s", req.Database)

	// Return success response
	response := Response{
		Status: "success",
		Data: map[string][]string{
			"sets": sets,
		},
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleIndexCreate handles the /index/create endpoint
func (s *Server) handleIndexCreate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		logRequest(r, start, http.StatusOK)
	}()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req CreateIndexRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Validate request
	if req.Database == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Database name is required")
		return
	}
	if req.Set == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Set name is required")
		return
	}
	if req.Name == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Index name is required")
		return
	}
	if req.Field == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Field name is required")
		return
	}

	// Get database
	db, err := s.DBManager.GetDatabase(req.Database)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "DB_NOT_FOUND", "Database not found")
		return
	}

	// Check database authentication
	username, password, hasAuth := ExtractDatabaseAuth(r)
	if !hasAuth {
		username = req.Auth.Username
		password = req.Auth.Password
	}
	if !db.Authenticate(username, password) {
		writeErrorResponse(w, http.StatusUnauthorized, "AUTH_FAILED", "Authentication failed")
		return
	}

	// Create index
	index, err := db.CreateIndex(req.Name, req.Set, req.Field)
	if err != nil {
		logger.Error("Failed to create index: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create index")
		return
	}

	logger.Info("Created index: %s on field: %s for set: %s in database: %s", req.Name, req.Field, req.Set, req.Database)

	// Return success response
	response := Response{
		Status:  "success",
		Message: "Index created successfully",
		Data: map[string]string{
			"index": index.Name,
		},
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleIndexDrop handles the /index/drop endpoint
func (s *Server) handleIndexDrop(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		logRequest(r, start, http.StatusOK)
	}()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req DropIndexRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Validate request
	if req.Database == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Database name is required")
		return
	}
	if req.Set == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Set name is required")
		return
	}
	if req.Name == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Index name is required")
		return
	}

	// Get database
	db, err := s.DBManager.GetDatabase(req.Database)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "DB_NOT_FOUND", "Database not found")
		return
	}

	// Check database authentication
	username, password, hasAuth := ExtractDatabaseAuth(r)
	if !hasAuth {
		username = req.Auth.Username
		password = req.Auth.Password
	}
	if !db.Authenticate(username, password) {
		writeErrorResponse(w, http.StatusUnauthorized, "AUTH_FAILED", "Authentication failed")
		return
	}

	// Check if index exists
	_, err = db.GetIndex(req.Name)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "INDEX_NOT_FOUND", "Index not found")
		return
	}

	// Drop index
	if err := db.DeleteIndex(req.Name); err != nil {
		logger.Error("Failed to drop index: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to drop index")
		return
	}

	logger.Info("Dropped index: %s from database: %s", req.Name, req.Database)

	// Return success response
	response := Response{
		Status:  "success",
		Message: "Index dropped successfully",
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// handleIndexQuery handles the /index/query endpoint
func (s *Server) handleIndexQuery(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		logRequest(r, start, http.StatusOK)
	}()

	// Check if this is a POST request
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	var req QueryIndexRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse request body")
		return
	}

	// Validate request
	if req.Database == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Database name is required")
		return
	}
	if req.Set == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Set name is required")
		return
	}
	if req.Index == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Index name is required")
		return
	}
	if req.Value == "" {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Value is required")
		return
	}

	// Set default values if not provided
	if req.Limit <= 0 {
		req.Limit = 100 // Default limit
	}

	// Get database
	db, err := s.DBManager.GetDatabase(req.Database)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "DB_NOT_FOUND", "Database not found")
		return
	}

	// Check database authentication
	username, password, hasAuth := ExtractDatabaseAuth(r)
	if !hasAuth {
		username = req.Auth.Username
		password = req.Auth.Password
	}
	if !db.Authenticate(username, password) {
		writeErrorResponse(w, http.StatusUnauthorized, "AUTH_FAILED", "Authentication failed")
		return
	}

	// Get set
	set, err := db.GetSet(req.Set)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "SET_NOT_FOUND", "Set not found")
		return
	}

	// Get index
	index, err := db.GetIndex(req.Index)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "INDEX_NOT_FOUND", "Index not found")
		return
	}

	// Query the index
	keys, err := index.Query(req.Value, req.Limit, req.Offset)
	if err != nil {
		logger.Error("Failed to query index: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query index")
		return
	}

	// Get the values for the keys
	results := make([]map[string]interface{}, 0, len(keys))
	for _, key := range keys {
		var value interface{}
		if err := set.Get(key, &value); err != nil {
			logger.Error("Failed to get value for key %s: %v", key, err)
			continue
		}

		results = append(results, map[string]interface{}{
			"key":   key,
			"value": value,
		})
	}

	logger.Info("Queried index: %s with value: %s in set: %s in database: %s, found %d results",
		req.Index, req.Value, req.Set, req.Database, len(results))

	// Return success response
	response := Response{
		Status: "success",
		Data: map[string]interface{}{
			"count": len(results),
			"data":  results,
		},
	}
	writeJSONResponse(w, http.StatusOK, response)
}