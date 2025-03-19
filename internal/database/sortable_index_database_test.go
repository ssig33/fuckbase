package database

import (
	"testing"
)

// TestDatabaseSortableIndexManagement tests the management of sortable indexes at the database level
func TestDatabaseSortableIndexManagement(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Add some data
	type Product struct {
		Name     string
		Category string
		Price    float64
		Stock    int
	}

	err = db.Put("test_set", "prod1", Product{Name: "Laptop", Category: "Electronics", Price: 1200.00, Stock: 10})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod2", Product{Name: "Smartphone", Category: "Electronics", Price: 800.00, Stock: 15})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Create a sortable index
	index1, err := db.CreateSortableIndex("category_price_index", "test_set", "Category", []string{"Price"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// Verify the index was created
	if index1 == nil {
		t.Fatalf("CreateSortableIndex returned nil index")
	}

	if index1.Name != "category_price_index" {
		t.Errorf("Expected index name 'category_price_index', got '%s'", index1.Name)
	}

	if index1.SetName != "test_set" {
		t.Errorf("Expected set name 'test_set', got '%s'", index1.SetName)
	}

	if index1.PrimaryField != "Category" {
		t.Errorf("Expected primary field 'Category', got '%s'", index1.PrimaryField)
	}

	if len(index1.SortFields) != 1 || index1.SortFields[0] != "Price" {
		t.Errorf("Expected sort fields ['Price'], got %v", index1.SortFields)
	}

	// Create another sortable index on the same set but with different fields
	index2, err := db.CreateSortableIndex("category_stock_index", "test_set", "Category", []string{"Stock"})
	if err != nil {
		t.Fatalf("Failed to create second sortable index: %v", err)
	}

	// Verify both indexes work correctly
	keys, err := index1.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query first sortable index: %v", err)
	}
	expectedKeys := []string{"prod2", "prod1"} // 800.00, 1200.00
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys from first index, got %d", len(expectedKeys), len(keys))
	}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d from first index, got %s", expectedKeys[i], i, key)
		}
	}

	keys, err = index2.QuerySorted("Electronics", "Stock", true)
	if err != nil {
		t.Fatalf("Failed to query second sortable index: %v", err)
	}
	expectedKeys = []string{"prod1", "prod2"} // 10, 15
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys from second index, got %d", len(expectedKeys), len(keys))
	}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d from second index, got %s", expectedKeys[i], i, key)
		}
	}

	// Test creating a sortable index with multiple sort fields
	index3, err := db.CreateSortableIndex("category_price_stock_index", "test_set", "Category", []string{"Price", "Stock"})
	if err != nil {
		t.Fatalf("Failed to create sortable index with multiple sort fields: %v", err)
	}

	// Verify the index with multiple sort fields
	if len(index3.SortFields) != 2 || index3.SortFields[0] != "Price" || index3.SortFields[1] != "Stock" {
		t.Errorf("Expected sort fields ['Price', 'Stock'], got %v", index3.SortFields)
	}

	// Test creating a sortable index with a non-existent set
	_, err = db.CreateSortableIndex("nonexistent_index", "nonexistent_set", "Category", []string{"Price"})
	if err == nil {
		t.Errorf("Expected error when creating sortable index on non-existent set, got nil")
	}

	// Test dropping a sortable index
	err = db.DropIndex("category_price_index")
	if err != nil {
		t.Fatalf("Failed to drop sortable index: %v", err)
	}

	// Verify the index was dropped
	_, err = db.GetIndex("category_price_index")
	if err == nil {
		t.Errorf("Expected error when getting dropped index, got nil")
	}

	// Verify other indexes still work
	keys, err = index2.QuerySorted("Electronics", "Stock", true)
	if err != nil {
		t.Fatalf("Failed to query second sortable index after dropping first: %v", err)
	}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys from second index after dropping first, got %d", len(expectedKeys), len(keys))
	}
}

// TestDatabaseSortableIndexRebuild tests rebuilding sortable indexes when data changes
func TestDatabaseSortableIndexRebuild(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Add some initial data
	type Product struct {
		Name     string
		Category string
		Price    float64
	}

	err = db.Put("test_set", "prod1", Product{Name: "Laptop", Category: "Electronics", Price: 1200.00})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod2", Product{Name: "Smartphone", Category: "Electronics", Price: 800.00})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Create a sortable index
	index, err := db.CreateSortableIndex("category_price_index", "test_set", "Category", []string{"Price"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// Verify initial state
	keys, err := index.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}
	expectedKeys := []string{"prod2", "prod1"} // 800.00, 1200.00
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys initially, got %d", len(expectedKeys), len(keys))
	}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d initially, got %s", expectedKeys[i], i, key)
		}
	}

	// Clear the set and add new data
	// Note: We're rebuilding the index after adding new data, so we don't need to clear the set
	
	err = db.Put("test_set", "prod3", Product{Name: "Tablet", Category: "Electronics", Price: 500.00})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod4", Product{Name: "Desktop", Category: "Electronics", Price: 1500.00})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod5", Product{Name: "Monitor", Category: "Electronics", Price: 300.00})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Rebuild the index
	err = db.RebuildIndex("category_price_index")
	if err != nil {
		t.Fatalf("Failed to rebuild sortable index: %v", err)
	}

	// Verify the index was rebuilt correctly
	keys, err = index.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index after rebuild: %v", err)
	}
	expectedKeys = []string{"prod5", "prod3", "prod4"} // 300.00, 500.00, 1500.00
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys after rebuild, got %d", len(expectedKeys), len(keys))
	}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d after rebuild, got %s", expectedKeys[i], i, key)
		}
	}

	// Test rebuilding a non-existent index
	err = db.RebuildIndex("nonexistent_index")
	if err == nil {
		t.Errorf("Expected error when rebuilding non-existent index, got nil")
	}
}

// TestDatabaseSortableIndexListAndGet tests listing and getting sortable indexes
func TestDatabaseSortableIndexListAndGet(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Add some data
	type Product struct {
		Name     string
		Category string
		Price    float64
	}

	err = db.Put("test_set", "prod1", Product{Name: "Laptop", Category: "Electronics", Price: 1200.00})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Create a regular index and a sortable index
	_, err = db.CreateIndex("category_index", "test_set", "Category")
	if err != nil {
		t.Fatalf("Failed to create regular index: %v", err)
	}

	_, err = db.CreateSortableIndex("category_price_index", "test_set", "Category", []string{"Price"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// List all indexes
	indexes := db.ListIndexes()
	if len(indexes) != 2 {
		t.Errorf("Expected 2 indexes, got %d", len(indexes))
	}

	// Check that both indexes are in the list
	foundRegular := false
	foundSortable := false
	for _, idx := range indexes {
		if idx == "category_index" {
			foundRegular = true
		} else if idx == "category_price_index" {
			foundSortable = true
		}
	}

	if !foundRegular {
		t.Errorf("Regular index 'category_index' not found in index list")
	}
	if !foundSortable {
		t.Errorf("Sortable index 'category_price_index' not found in index list")
	}

	// Get the sortable index
	index, err := db.GetIndex("category_price_index")
	if err != nil {
		t.Fatalf("Failed to get sortable index: %v", err)
	}

	// Verify it's a sortable index
	sortableIndex, ok := index.(*SortableIndex)
	if !ok {
		t.Fatalf("Expected index to be a SortableIndex, got %T", index)
	}

	if sortableIndex.PrimaryField != "Category" {
		t.Errorf("Expected primary field 'Category', got '%s'", sortableIndex.PrimaryField)
	}

	if len(sortableIndex.SortFields) != 1 || sortableIndex.SortFields[0] != "Price" {
		t.Errorf("Expected sort fields ['Price'], got %v", sortableIndex.SortFields)
	}

	// Get the regular index
	index, err = db.GetIndex("category_index")
	if err != nil {
		t.Fatalf("Failed to get regular index: %v", err)
	}

	// Verify it's a regular index
	regularIndex, ok := index.(*BasicIndex)
	if !ok {
		t.Fatalf("Expected index to be a BasicIndex, got %T", index)
	}

	if regularIndex.Field != "Category" {
		t.Errorf("Expected field 'Category', got '%s'", regularIndex.Field)
	}
}