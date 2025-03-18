package database

import (
	"testing"
)

// TestIndexIntegrationWithMissingField tests the integration between sets and indexes
// when values without the indexed field are added to a set
func TestIndexIntegrationWithMissingField(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	set, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Add some data with the expected field
	type Person struct {
		Name  string
		Age   int
		Email string
	}

	// Add initial data with the Name field
	set.Put("key1", Person{Name: "Alice", Age: 30, Email: "alice@example.com"})
	set.Put("key2", Person{Name: "Bob", Age: 25, Email: "bob@example.com"})

	// Create an index on the Name field
	index, err := db.CreateIndex("name_index", "test_set", "Name")
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Now add a value without the Name field
	type WithoutName struct {
		Age   int
		Email string
	}

	// Add the data to the set
	err = set.Put("key3", WithoutName{Age: 40, Email: "noname@example.com"})
	if err != nil {
		t.Fatalf("Failed to add data without Name field: %v", err)
	}

	// Now manually update the index with this new entry
	rawData, err := set.GetRaw("key3")
	if err != nil {
		t.Fatalf("Failed to get raw data: %v", err)
	}

	err = index.AddEntry("key3", rawData)
	if err != nil {
		t.Errorf("Failed to add entry without Name field to index: %v", err)
	}

	// Verify the entry was added to the index under the "__MISSING__" value
	keys, err := index.Query("__MISSING__", "", 0, 0)
	if err != nil {
		t.Fatalf("Failed to query index: %v", err)
	}
	if len(keys) != 1 || keys[0] != "key3" {
		t.Errorf("Expected index to have one entry for '__MISSING__', got %v", keys)
	}

	// Add another value without the Name field
	err = set.Put("key4", WithoutName{Age: 50, Email: "another@example.com"})
	if err != nil {
		t.Fatalf("Failed to add second data without Name field: %v", err)
	}

	// Manually update the index again
	rawData, err = set.GetRaw("key4")
	if err != nil {
		t.Fatalf("Failed to get raw data: %v", err)
	}

	err = index.AddEntry("key4", rawData)
	if err != nil {
		t.Errorf("Failed to add second entry without Name field to index: %v", err)
	}

	// Verify both entries are now in the index under "__MISSING__"
	keys, err = index.Query("__MISSING__", "", 0, 0)
	if err != nil {
		t.Fatalf("Failed to query index: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("Expected index to have two entries for '__MISSING__', got %d entries", len(keys))
	}
	
	// Check if both keys are present
	foundKey3 := false
	foundKey4 := false
	for _, key := range keys {
		if key == "key3" {
			foundKey3 = true
		}
		if key == "key4" {
			foundKey4 = true
		}
	}
	if !foundKey3 || !foundKey4 {
		t.Errorf("Expected to find both 'key3' and 'key4' in the index, got %v", keys)
	}
}