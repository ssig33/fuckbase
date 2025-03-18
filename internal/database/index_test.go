package database

import (
	"testing"
)

// TestIndexWithMissingField tests the behavior when a value without the indexed field
// is added to a set with an index
func TestIndexWithMissingField(t *testing.T) {
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

	// Verify the index has the expected entries
	keys, err := index.Query("Alice", "", 0, 0)
	if err != nil {
		t.Fatalf("Failed to query index: %v", err)
	}
	if len(keys) != 1 || keys[0] != "key1" {
		t.Errorf("Expected index to have one entry for 'Alice', got %v", keys)
	}

	// Now add a value without the Name field
	type WithoutName struct {
		Age   int
		Email string
	}

	// This will add the data to the set but not update the index
	err = set.Put("key3", WithoutName{Age: 40, Email: "noname@example.com"})
	if err != nil {
		t.Fatalf("Failed to add data without Name field: %v", err)
	}

	// Verify the data was added to the set
	var result WithoutName
	err = set.Get("key3", &result)
	if err != nil {
		t.Errorf("Failed to get data without Name field: %v", err)
	}
	if result.Age != 40 || result.Email != "noname@example.com" {
		t.Errorf("Expected {Age: 40, Email: noname@example.com}, got %+v", result)
	}

	// Now manually try to update the index with this new entry
	// With our new implementation, this should silently skip the entry
	rawData, err := set.GetRaw("key3")
	if err != nil {
		t.Fatalf("Failed to get raw data: %v", err)
	}

	err = index.AddEntry("key3", rawData)
	if err != nil {
		t.Errorf("Failed to add entry without Name field to index: %v", err)
	}

	// Verify the entry was NOT added to the index
	// We'll check all values in the index to make sure key3 is not there
	allValues := index.GetAllValues()
	for _, value := range allValues {
		keys, err = index.Query(value, "", 0, 0)
		if err != nil {
			t.Fatalf("Failed to query index: %v", err)
		}
		for _, key := range keys {
			if key == "key3" {
				t.Errorf("Found key3 in index under value '%s', but it should not be indexed", value)
			}
		}
	}

	// Also verify that we can still query for the other values
	keys, err = index.Query("Alice", "", 0, 0)
	if err != nil {
		t.Fatalf("Failed to query index: %v", err)
	}
	if len(keys) != 1 || keys[0] != "key1" {
		t.Errorf("Expected index to have one entry for 'Alice', got %v", keys)
	}
}