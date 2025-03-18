package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ssig33/fuckbase/internal/config"
	"github.com/ssig33/fuckbase/internal/database"
	"github.com/ssig33/fuckbase/internal/logger"
)

func init() {
	// Initialize logger for tests
	logger.InitLogger("info", "stdout")
}

func TestServerEndpoints(t *testing.T) {
	// Create a new server
	cfg := config.NewServerConfig()
	dbManager := database.NewManager()
	srv := NewServer(cfg, dbManager)

	// Test database creation
	t.Run("CreateDatabase", func(t *testing.T) {
		// Create a request
		reqBody := CreateDatabaseRequest{
			Name: "test_db",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rr := httptest.NewRecorder()

		// Call the handler
		srv.handleDatabaseCreate(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check the response body
		var resp Response
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Errorf("Failed to parse response body: %v", err)
		}
		if resp.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", resp.Status)
		}
		if resp.Message != "Database created successfully" {
			t.Errorf("Expected message 'Database created successfully', got '%s'", resp.Message)
		}
	})

	// Test creating a database with the same name
	t.Run("CreateDuplicateDatabase", func(t *testing.T) {
		// Create a request
		reqBody := CreateDatabaseRequest{
			Name: "test_db",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rr := httptest.NewRecorder()

		// Call the handler
		srv.handleDatabaseCreate(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusConflict {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusConflict)
		}

		// Check the response body
		var resp ErrorResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Errorf("Failed to parse response body: %v", err)
		}
		if resp.Status != "error" {
			t.Errorf("Expected status 'error', got '%s'", resp.Status)
		}
		if resp.Code != "DB_ALREADY_EXISTS" {
			t.Errorf("Expected code 'DB_ALREADY_EXISTS', got '%s'", resp.Code)
		}
	})

	// Test set creation
	t.Run("CreateSet", func(t *testing.T) {
		// Create a request
		reqBody := CreateSetRequest{
			Database: "test_db",
			Name:     "test_set",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/set/create", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rr := httptest.NewRecorder()

		// Call the handler
		srv.handleSetCreate(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check the response body
		var resp Response
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Errorf("Failed to parse response body: %v", err)
		}
		if resp.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", resp.Status)
		}
		if resp.Message != "Set created successfully" {
			t.Errorf("Expected message 'Set created successfully', got '%s'", resp.Message)
		}
	})

	// Test server info
	t.Run("ServerInfo", func(t *testing.T) {
		// Create a request
		req := httptest.NewRequest(http.MethodPost, "/server/info", nil)
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rr := httptest.NewRecorder()

		// Call the handler
		srv.handleServerInfo(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check the response body
		var resp ServerInfoResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Errorf("Failed to parse response body: %v", err)
		}
		if resp.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", resp.Status)
		}
		if resp.DatabasesCount != 1 {
			t.Errorf("Expected 1 database, got %d", resp.DatabasesCount)
		}
	})

	// Test database deletion
	t.Run("DeleteDatabase", func(t *testing.T) {
		// Create a request
		reqBody := DropDatabaseRequest{
			Name: "test_db",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/drop", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rr := httptest.NewRecorder()

		// Call the handler
		srv.handleDatabaseDrop(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check the response body
		var resp Response
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Errorf("Failed to parse response body: %v", err)
		}
		if resp.Status != "success" {
			t.Errorf("Expected status 'success', got '%s'", resp.Status)
		}
		if resp.Message != "Database dropped successfully" {
			t.Errorf("Expected message 'Database dropped successfully', got '%s'", resp.Message)
		}
	})

	// Test deleting a nonexistent database
	t.Run("DeleteNonexistentDatabase", func(t *testing.T) {
		// Create a request
		reqBody := DropDatabaseRequest{
			Name: "nonexistent",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/drop", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rr := httptest.NewRecorder()

		// Call the handler
		srv.handleDatabaseDrop(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
		}

		// Check the response body
		var resp ErrorResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Errorf("Failed to parse response body: %v", err)
		}
		if resp.Status != "error" {
			t.Errorf("Expected status 'error', got '%s'", resp.Status)
		}
		if resp.Code != "DB_NOT_FOUND" {
			t.Errorf("Expected code 'DB_NOT_FOUND', got '%s'", resp.Code)
		}
	})
}

func TestMethodNotAllowed(t *testing.T) {
	// Create a new server
	cfg := config.NewServerConfig()
	dbManager := database.NewManager()
	srv := NewServer(cfg, dbManager)

	// Test GET request to /create
	req := httptest.NewRequest(http.MethodGet, "/create", nil)
	rr := httptest.NewRecorder()
	srv.handleDatabaseCreate(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}

	// Check the response body
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response body: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", resp.Status)
	}
	if resp.Code != "METHOD_NOT_ALLOWED" {
		t.Errorf("Expected code 'METHOD_NOT_ALLOWED', got '%s'", resp.Code)
	}
}

func TestInvalidRequest(t *testing.T) {
	// Create a new server
	cfg := config.NewServerConfig()
	dbManager := database.NewManager()
	srv := NewServer(cfg, dbManager)

	// Test invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.handleDatabaseCreate(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Check the response body
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response body: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", resp.Status)
	}
	if resp.Code != "INVALID_REQUEST" {
		t.Errorf("Expected code 'INVALID_REQUEST', got '%s'", resp.Code)
	}
}

func TestMissingRequiredFields(t *testing.T) {
	// Create a new server
	cfg := config.NewServerConfig()
	dbManager := database.NewManager()
	srv := NewServer(cfg, dbManager)

	// Test missing database name
	reqBody := CreateDatabaseRequest{
		Name: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.handleDatabaseCreate(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Check the response body
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response body: %v", err)
	}
	if resp.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", resp.Status)
	}
	if resp.Code != "INVALID_REQUEST" {
		t.Errorf("Expected code 'INVALID_REQUEST', got '%s'", resp.Code)
	}
}