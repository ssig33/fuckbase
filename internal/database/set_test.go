package database

import (
	"testing"
)

func TestSetOperations(t *testing.T) {
	// Create a new set
	set := NewSet("test_set")
	if set.Name != "test_set" {
		t.Errorf("Expected set name to be 'test_set', got '%s'", set.Name)
	}

	// Test Put and Get
	type TestData struct {
		Name  string
		Value int
	}

	// Put a value
	data := TestData{Name: "test", Value: 42}
	err := set.Put("key1", data)
	if err != nil {
		t.Errorf("Failed to put value: %v", err)
	}

	// Get the value
	var result TestData
	err = set.Get("key1", &result)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}

	// Check the value
	if result.Name != data.Name || result.Value != data.Value {
		t.Errorf("Expected %+v, got %+v", data, result)
	}

	// Test Has
	if !set.Has("key1") {
		t.Errorf("Expected set to have key 'key1'")
	}
	if set.Has("nonexistent") {
		t.Errorf("Expected set to not have key 'nonexistent'")
	}

	// Test Delete
	err = set.Delete("key1")
	if err != nil {
		t.Errorf("Failed to delete key: %v", err)
	}
	if set.Has("key1") {
		t.Errorf("Expected set to not have key 'key1' after deletion")
	}

	// Test deleting a nonexistent key
	err = set.Delete("nonexistent")
	if err == nil {
		t.Errorf("Expected error when deleting nonexistent key")
	}

	// Test Keys
	set.Put("key1", "value1")
	set.Put("key2", "value2")
	set.Put("key3", "value3")

	keys := set.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Test Size
	if set.Size() != 3 {
		t.Errorf("Expected size 3, got %d", set.Size())
	}

	// Test Clear
	set.Clear()
	if set.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", set.Size())
	}
}

func TestSetForEach(t *testing.T) {
	// Create a new set
	set := NewSet("test_set")

	// Add some data
	set.Put("key1", "value1")
	set.Put("key2", "value2")
	set.Put("key3", "value3")

	// Test ForEach
	count := 0
	err := set.ForEach(func(key string, value []byte) error {
		count++
		return nil
	})

	if err != nil {
		t.Errorf("ForEach returned error: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected ForEach to iterate over 3 items, got %d", count)
	}

	// Test ForEach with error
	expectedError := "test error"
	err = set.ForEach(func(key string, value []byte) error {
		if key == "key2" {
			return &testError{expectedError}
		}
		return nil
	})

	if err == nil {
		t.Errorf("Expected ForEach to return error")
	} else if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

// testError is a custom error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}