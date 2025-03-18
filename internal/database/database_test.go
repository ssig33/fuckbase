package database

import (
	"testing"
)

func TestDatabaseOperations(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)
	if db.Name != "test_db" {
		t.Errorf("Expected database name to be 'test_db', got '%s'", db.Name)
	}

	// Test CreateSet
	set, err := db.CreateSet("test_set")
	if err != nil {
		t.Errorf("Failed to create set: %v", err)
	}
	if set.Name != "test_set" {
		t.Errorf("Expected set name to be 'test_set', got '%s'", set.Name)
	}

	// Test creating a set with the same name
	_, err = db.CreateSet("test_set")
	if err == nil {
		t.Errorf("Expected error when creating a set with the same name")
	}

	// Test GetSet
	set, err = db.GetSet("test_set")
	if err != nil {
		t.Errorf("Failed to get set: %v", err)
	}
	if set.Name != "test_set" {
		t.Errorf("Expected set name to be 'test_set', got '%s'", set.Name)
	}

	// Test getting a nonexistent set
	_, err = db.GetSet("nonexistent")
	if err == nil {
		t.Errorf("Expected error when getting a nonexistent set")
	}

	// Test ListSets
	sets := db.ListSets()
	if len(sets) != 1 {
		t.Errorf("Expected 1 set, got %d", len(sets))
	}
	if sets[0] != "test_set" {
		t.Errorf("Expected set name to be 'test_set', got '%s'", sets[0])
	}

	// Test DeleteSet
	err = db.DeleteSet("test_set")
	if err != nil {
		t.Errorf("Failed to delete set: %v", err)
	}
	sets = db.ListSets()
	if len(sets) != 0 {
		t.Errorf("Expected 0 sets after deletion, got %d", len(sets))
	}

	// Test deleting a nonexistent set
	err = db.DeleteSet("nonexistent")
	if err == nil {
		t.Errorf("Expected error when deleting a nonexistent set")
	}
}

func TestDatabaseAuthentication(t *testing.T) {
	// Create a database without authentication
	db := NewDatabase("test_db", nil)
	if !db.Authenticate("any", "any") {
		t.Errorf("Expected authentication to succeed when auth is not enabled")
	}

	// Create a database with authentication
	auth := &AuthConfig{
		Username: "admin",
		Password: "password",
		Enabled:  true,
	}
	db = NewDatabase("test_db", auth)

	// Test valid credentials
	if !db.Authenticate("admin", "password") {
		t.Errorf("Expected authentication to succeed with valid credentials")
	}

	// Test invalid username
	if db.Authenticate("wrong", "password") {
		t.Errorf("Expected authentication to fail with invalid username")
	}

	// Test invalid password
	if db.Authenticate("admin", "wrong") {
		t.Errorf("Expected authentication to fail with invalid password")
	}
}

func TestDatabaseIndex(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	set, _ := db.CreateSet("test_set")

	// Add some data to the set
	type TestData struct {
		Name  string
		Value int
	}
	set.Put("key1", TestData{Name: "Alice", Value: 30})
	set.Put("key2", TestData{Name: "Bob", Value: 25})
	set.Put("key3", TestData{Name: "Charlie", Value: 35})

	// Create an index on the Name field
	index, err := db.CreateIndex("name_index", "test_set", "Name")
	if err != nil {
		t.Errorf("Failed to create index: %v", err)
	}
	if index.Name != "name_index" {
		t.Errorf("Expected index name to be 'name_index', got '%s'", index.Name)
	}

	// Test creating an index with the same name
	_, err = db.CreateIndex("name_index", "test_set", "Value")
	if err == nil {
		t.Errorf("Expected error when creating an index with the same name")
	}

	// Test GetIndex
	index, err = db.GetIndex("name_index")
	if err != nil {
		t.Errorf("Failed to get index: %v", err)
	}
	if index.Name != "name_index" {
		t.Errorf("Expected index name to be 'name_index', got '%s'", index.Name)
	}

	// Test getting a nonexistent index
	_, err = db.GetIndex("nonexistent")
	if err == nil {
		t.Errorf("Expected error when getting a nonexistent index")
	}

	// Test ListIndexes
	indexes := db.ListIndexes()
	if len(indexes) != 1 {
		t.Errorf("Expected 1 index, got %d", len(indexes))
	}
	if indexes[0] != "name_index" {
		t.Errorf("Expected index name to be 'name_index', got '%s'", indexes[0])
	}

	// Test DeleteIndex
	err = db.DeleteIndex("name_index")
	if err != nil {
		t.Errorf("Failed to delete index: %v", err)
	}
	indexes = db.ListIndexes()
	if len(indexes) != 0 {
		t.Errorf("Expected 0 indexes after deletion, got %d", len(indexes))
	}

	// Test deleting a nonexistent index
	err = db.DeleteIndex("nonexistent")
	if err == nil {
		t.Errorf("Expected error when deleting a nonexistent index")
	}
}