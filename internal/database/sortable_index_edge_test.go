package database

import (
	"testing"
)

// TestSortableIndexEdgeCases tests edge cases for sortable indexes
func TestSortableIndexEdgeCases(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Test 1: Empty set with sortable index
	t.Run("EmptySetWithSortableIndex", func(t *testing.T) {
		// Create a sortable index on an empty set
		index, err := db.CreateSortableIndex("empty_index", "test_set", "category", []string{"price"})
		if err != nil {
			t.Fatalf("Failed to create sortable index on empty set: %v", err)
		}

		// Query the empty index
		keys, err := index.QuerySorted("any_category", "price", true)
		if err != nil {
			t.Fatalf("Failed to query empty sortable index: %v", err)
		}

		// Verify the result is empty
		if len(keys) != 0 {
			t.Errorf("Expected 0 keys from empty index, got %d", len(keys))
		}
	})

	// Test 2: Null values in sort fields
	t.Run("NullValuesInSortFields", func(t *testing.T) {
		// Add data with null/nil values for sort fields
		type Product struct {
			Category string
			Price    *float64 // Pointer to allow nil
			Rating   *float64 // Pointer to allow nil
		}

		// Create a value with nil price
		db.Put("test_set", "null_price", Product{
			Category: "Electronics",
			Price:    nil,
			Rating:   createFloat64Ptr(4.5),
		})

		// Create a value with nil rating
		db.Put("test_set", "null_rating", Product{
			Category: "Electronics",
			Price:    createFloat64Ptr(100.0),
			Rating:   nil,
		})

		// Create a value with both fields set
		db.Put("test_set", "both_set", Product{
			Category: "Electronics",
			Price:    createFloat64Ptr(200.0),
			Rating:   createFloat64Ptr(3.5),
		})

		// Create a sortable index
		index, err := db.CreateSortableIndex("null_test_index", "test_set", "Category", []string{"Price", "Rating"})
		if err != nil {
			t.Fatalf("Failed to create sortable index: %v", err)
		}

		// Query the index
		keys, err := index.QuerySorted("Electronics", "Price", true)
		if err != nil {
			t.Fatalf("Failed to query sortable index: %v", err)
		}

		// Verify that null values are handled correctly (should be at the end or beginning)
		if len(keys) != 3 {
			t.Errorf("Expected 3 keys, got %d", len(keys))
		}

		// The item with null price should be either at the beginning or end
		if keys[0] != "null_price" && keys[len(keys)-1] != "null_price" {
			t.Errorf("Expected null_price to be at the beginning or end, got keys: %v", keys)
		}
	})

	// Test 3: Very large number of items with the same primary field value
	t.Run("LargeNumberOfItemsWithSamePrimaryField", func(t *testing.T) {
		// Add a large number of items with the same category
		numItems := 100
		for i := 0; i < numItems; i++ {
			db.Put("test_set", "large_"+string(rune(i)), map[string]interface{}{
				"Category": "LargeCategory",
				"Price":    float64(i),
			})
		}

		// Create a sortable index
		index, err := db.CreateSortableIndex("large_index", "test_set", "Category", []string{"Price"})
		if err != nil {
			t.Fatalf("Failed to create sortable index: %v", err)
		}

		// Query all items
		keys, err := index.QuerySorted("LargeCategory", "Price", true)
		if err != nil {
			t.Fatalf("Failed to query sortable index: %v", err)
		}

		// Verify we got all items
		if len(keys) != numItems {
			t.Errorf("Expected %d keys, got %d", numItems, len(keys))
		}

		// Test pagination with various page sizes
		pageSizes := []int{1, 10, 25, 50, 99}
		for _, pageSize := range pageSizes {
			keys, err := index.QuerySortedWithPagination("LargeCategory", "Price", true, 0, pageSize)
			if err != nil {
				t.Fatalf("Failed to query with pagination (size %d): %v", pageSize, err)
			}

			if len(keys) != pageSize {
				t.Errorf("Expected %d keys for page size %d, got %d", pageSize, pageSize, len(keys))
			}
		}
	})

	// Test 4: Unicode characters in field values
	t.Run("UnicodeCharactersInFieldValues", func(t *testing.T) {
		// Add items with Unicode characters
		unicodeItems := []struct {
			key      string
			category string
			name     string
		}{
			{"unicode1", "Café", "Espresso"},
			{"unicode2", "Café", "Latte"},
			{"unicode3", "Résumé", "CV"},
			{"unicode4", "日本語", "漢字"},
			{"unicode5", "日本語", "ひらがな"},
		}

		for _, item := range unicodeItems {
			db.Put("test_set", item.key, map[string]interface{}{
				"Category": item.category,
				"Name":     item.name,
			})
		}

		// Create a sortable index
		index, err := db.CreateSortableIndex("unicode_index", "test_set", "Category", []string{"Name"})
		if err != nil {
			t.Fatalf("Failed to create sortable index: %v", err)
		}

		// Query for café items
		keys, err := index.QuerySorted("Café", "Name", true)
		if err != nil {
			t.Fatalf("Failed to query sortable index: %v", err)
		}

		// Verify we got the right items
		if len(keys) != 2 {
			t.Errorf("Expected 2 keys for Café, got %d", len(keys))
		}

		// Query for Japanese items
		keys, err = index.QuerySorted("日本語", "Name", true)
		if err != nil {
			t.Fatalf("Failed to query sortable index: %v", err)
		}

		// Verify we got the right items
		if len(keys) != 2 {
			t.Errorf("Expected 2 keys for 日本語, got %d", len(keys))
		}
	})
}

// Helper function to create a float64 pointer
func createFloat64Ptr(value float64) *float64 {
	return &value
}