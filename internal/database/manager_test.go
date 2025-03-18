package database

import (
	"testing"
)

func TestManagerOperations(t *testing.T) {
	// Create a new manager
	manager := NewManager()
	if len(manager.Databases) != 0 {
		t.Errorf("Expected 0 databases, got %d", len(manager.Databases))
	}

	// Test CreateDatabase
	db, err := manager.CreateDatabase("test_db", nil)
	if err != nil {
		t.Errorf("Failed to create database: %v", err)
	}
	if db.Name != "test_db" {
		t.Errorf("Expected database name to be 'test_db', got '%s'", db.Name)
	}

	// Test creating a database with the same name
	_, err = manager.CreateDatabase("test_db", nil)
	if err == nil {
		t.Errorf("Expected error when creating a database with the same name")
	}

	// Test GetDatabase
	db, err = manager.GetDatabase("test_db")
	if err != nil {
		t.Errorf("Failed to get database: %v", err)
	}
	if db.Name != "test_db" {
		t.Errorf("Expected database name to be 'test_db', got '%s'", db.Name)
	}

	// Test getting a nonexistent database
	_, err = manager.GetDatabase("nonexistent")
	if err == nil {
		t.Errorf("Expected error when getting a nonexistent database")
	}

	// Test DatabaseExists
	if !manager.DatabaseExists("test_db") {
		t.Errorf("Expected database 'test_db' to exist")
	}
	if manager.DatabaseExists("nonexistent") {
		t.Errorf("Expected database 'nonexistent' to not exist")
	}

	// Test ListDatabases
	databases := manager.ListDatabases()
	if len(databases) != 1 {
		t.Errorf("Expected 1 database, got %d", len(databases))
	}
	if databases[0] != "test_db" {
		t.Errorf("Expected database name to be 'test_db', got '%s'", databases[0])
	}

	// Test GetDatabaseCount
	count := manager.GetDatabaseCount()
	if count != 1 {
		t.Errorf("Expected 1 database, got %d", count)
	}

	// Test DeleteDatabase
	err = manager.DeleteDatabase("test_db")
	if err != nil {
		t.Errorf("Failed to delete database: %v", err)
	}
	if manager.DatabaseExists("test_db") {
		t.Errorf("Expected database 'test_db' to not exist after deletion")
	}

	// Test deleting a nonexistent database
	err = manager.DeleteDatabase("nonexistent")
	if err == nil {
		t.Errorf("Expected error when deleting a nonexistent database")
	}
}

func TestManagerAuthentication(t *testing.T) {
	// Create a new manager
	manager := NewManager()

	// Create a database without authentication
	manager.CreateDatabase("db_no_auth", nil)

	// Create a database with authentication
	auth := &AuthConfig{
		Username: "admin",
		Password: "password",
		Enabled:  true,
	}
	manager.CreateDatabase("db_with_auth", auth)

	// Test authentication for database without auth
	authenticated, err := manager.AuthenticateDatabase("db_no_auth", "any", "any")
	if err != nil {
		t.Errorf("Failed to authenticate: %v", err)
	}
	if !authenticated {
		t.Errorf("Expected authentication to succeed for database without auth")
	}

	// Test authentication for database with auth
	authenticated, err = manager.AuthenticateDatabase("db_with_auth", "admin", "password")
	if err != nil {
		t.Errorf("Failed to authenticate: %v", err)
	}
	if !authenticated {
		t.Errorf("Expected authentication to succeed with valid credentials")
	}

	// Test authentication with invalid credentials
	authenticated, err = manager.AuthenticateDatabase("db_with_auth", "admin", "wrong")
	if err != nil {
		t.Errorf("Failed to authenticate: %v", err)
	}
	if authenticated {
		t.Errorf("Expected authentication to fail with invalid credentials")
	}

	// Test authentication for nonexistent database
	_, err = manager.AuthenticateDatabase("nonexistent", "admin", "password")
	if err == nil {
		t.Errorf("Expected error when authenticating against a nonexistent database")
	}
}