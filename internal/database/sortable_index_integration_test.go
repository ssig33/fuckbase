package database

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// TestSortableIndexIntegrationWithPagination tests the integration between sortable indexes and pagination
func TestSortableIndexIntegrationWithPagination(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Add some data with multiple fields
	type Product struct {
		Name     string
		Category string
		Price    float64
		Rating   float64
		Stock    int
	}

	// Add a large number of products in the same category for pagination testing
	for i := 1; i <= 20; i++ {
		// Create products with different prices
		price := 100.0 + float64(i*50)
		rating := 3.0 + float64(i%5)/10.0
		stock := 10 + i*5
		
		err := db.Put("test_set", fmt.Sprintf("prod%d", i), Product{
			Name:     fmt.Sprintf("Product %d", i),
			Category: "Electronics",
			Price:    price,
			Rating:   rating,
			Stock:    stock,
		})
		if err != nil {
			t.Fatalf("Failed to put product: %v", err)
		}
	}

	// Create a sortable index on Category (primary) with Price (sort)
	sortableIndex, err := db.CreateSortableIndex("category_price_index", "test_set", "Category", []string{"Price"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// Test pagination with different page sizes and offsets
	
	// First page (offset 0, limit 5)
	keys, err := sortableIndex.QuerySortedWithPagination("Electronics", "Price", true, 0, 5)
	if err != nil {
		t.Fatalf("Failed to query first page: %v", err)
	}
	
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys for first page, got %d", len(keys))
	}
	
	// Verify the keys are in ascending order by price
	for i := 0; i < len(keys)-1; i++ {
		prodNum1, _ := strconv.Atoi(strings.TrimPrefix(keys[i], "prod"))
		prodNum2, _ := strconv.Atoi(strings.TrimPrefix(keys[i+1], "prod"))
		if prodNum1 > prodNum2 {
			t.Errorf("Products not in ascending price order: %s before %s", keys[i], keys[i+1])
		}
	}
	
	// Second page (offset 5, limit 5)
	keys, err = sortableIndex.QuerySortedWithPagination("Electronics", "Price", true, 5, 5)
	if err != nil {
		t.Fatalf("Failed to query second page: %v", err)
	}
	
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys for second page, got %d", len(keys))
	}
	
	// Third page (offset 10, limit 5)
	keys, err = sortableIndex.QuerySortedWithPagination("Electronics", "Price", true, 10, 5)
	if err != nil {
		t.Fatalf("Failed to query third page: %v", err)
	}
	
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys for third page, got %d", len(keys))
	}
	
	// Fourth page (offset 15, limit 5)
	keys, err = sortableIndex.QuerySortedWithPagination("Electronics", "Price", true, 15, 5)
	if err != nil {
		t.Fatalf("Failed to query fourth page: %v", err)
	}
	
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys for fourth page, got %d", len(keys))
	}
	
	// Beyond available results (offset 20, limit 5)
	keys, err = sortableIndex.QuerySortedWithPagination("Electronics", "Price", true, 20, 5)
	if err != nil {
		t.Fatalf("Failed to query beyond available results: %v", err)
	}
	
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys beyond available results, got %d", len(keys))
	}
	
	// Test with descending order
	keys, err = sortableIndex.QuerySortedWithPagination("Electronics", "Price", false, 0, 5)
	if err != nil {
		t.Fatalf("Failed to query with descending order: %v", err)
	}
	
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys with descending order, got %d", len(keys))
	}
	
	// Verify the keys are in descending order by price
	for i := 0; i < len(keys)-1; i++ {
		prodNum1, _ := strconv.Atoi(strings.TrimPrefix(keys[i], "prod"))
		prodNum2, _ := strconv.Atoi(strings.TrimPrefix(keys[i+1], "prod"))
		if prodNum1 < prodNum2 {
			t.Errorf("Products not in descending price order: %s before %s", keys[i], keys[i+1])
		}
	}
	
	// Test with different limit
	keys, err = sortableIndex.QuerySortedWithPagination("Electronics", "Price", true, 0, 10)
	if err != nil {
		t.Fatalf("Failed to query with different limit: %v", err)
	}
	
	if len(keys) != 10 {
		t.Errorf("Expected 10 keys with different limit, got %d", len(keys))
	}
	
	// Test with multi-field sorting and pagination
	sortableIndex, err = db.CreateSortableIndex("category_rating_price_index", "test_set", "Category", []string{"Rating", "Price"})
	if err != nil {
		t.Fatalf("Failed to create multi-field sortable index: %v", err)
	}
	
	// Test pagination with multi-field sorting
	keys, err = sortableIndex.QueryMultiSortedWithPagination("Electronics", []string{"Rating", "Price"}, []bool{true, false}, 0, 5)
	if err != nil {
		t.Fatalf("Failed to query with multi-field sorting and pagination: %v", err)
	}
	
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys with multi-field sorting and pagination, got %d", len(keys))
	}
	
	// Add a product to the set and verify pagination still works correctly
	err = db.Put("test_set", "prod21", Product{Name: "Product 21", Category: "Electronics", Price: 1500.0, Rating: 4.5, Stock: 100})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	// Query the first page again
	keys, err = sortableIndex.QuerySortedWithPagination("Electronics", "Rating", true, 0, 5)
	if err != nil {
		t.Fatalf("Failed to query first page after adding new product: %v", err)
	}
	
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys for first page after adding new product, got %d", len(keys))
	}
	
	// Delete a product and verify pagination adjusts correctly
	err = db.Delete("test_set", "prod1")
	if err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}
	
	// Query the first page again
	keys, err = sortableIndex.QuerySortedWithPagination("Electronics", "Rating", true, 0, 5)
	if err != nil {
		t.Fatalf("Failed to query first page after deleting product: %v", err)
	}
	
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys for first page after deleting product, got %d", len(keys))
	}
	
	// Verify the deleted product is not in the results
	for _, key := range keys {
		if key == "prod1" {
			t.Errorf("Deleted product found in query results")
		}
	}
}

// TestSortableIndexIntegrationWithSet tests the integration between sortable indexes and sets
func TestSortableIndexIntegrationWithSet(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Add some data with multiple fields
	type Product struct {
		Name     string
		Category string
		Price    float64
		Stock    int
		Rating   float64
	}

	// Add products with different categories, prices, and stocks
	err = db.Put("test_set", "prod1", Product{Name: "Laptop", Category: "Electronics", Price: 1200.00, Stock: 10, Rating: 4.5})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod2", Product{Name: "Smartphone", Category: "Electronics", Price: 800.00, Stock: 15, Rating: 4.7})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod3", Product{Name: "Headphones", Category: "Electronics", Price: 150.00, Stock: 30, Rating: 4.2})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod4", Product{Name: "T-shirt", Category: "Clothing", Price: 25.00, Stock: 100, Rating: 4.0})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod5", Product{Name: "Jeans", Category: "Clothing", Price: 50.00, Stock: 75, Rating: 4.3})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod6", Product{Name: "Book", Category: "Books", Price: 15.00, Stock: 50, Rating: 4.8})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Create a sortable index on Category (primary) with Price and Stock (sort)
	sortableIndex, err := db.CreateSortableIndex("category_price_stock_index", "test_set", "Category", []string{"Price", "Stock"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// Test querying with sorting by Price in ascending order
	keys, err := sortableIndex.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}

	// Verify the keys are returned in the correct order (sorted by Price ascending)
	expectedKeys := []string{"prod3", "prod2", "prod1"} // 150.00, 800.00, 1200.00
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d, got %s", expectedKeys[i], i, key)
		}
	}

	// Add a new product to the set
	err = db.Put("test_set", "prod7", Product{Name: "Tablet", Category: "Electronics", Price: 500.00, Stock: 20, Rating: 4.6})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Verify the index is automatically updated
	keys, err = sortableIndex.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index after adding new product: %v", err)
	}

	// Verify the keys are returned in the correct order with the new product
	expectedKeys = []string{"prod3", "prod7", "prod2", "prod1"} // 150.00, 500.00, 800.00, 1200.00
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys after adding new product, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d after adding new product, got %s", expectedKeys[i], i, key)
		}
	}

	// Update an existing product
	err = db.Put("test_set", "prod2", Product{Name: "Smartphone", Category: "Electronics", Price: 900.00, Stock: 15, Rating: 4.7})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Verify the index is updated with the new price
	keys, err = sortableIndex.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index after updating product: %v", err)
	}

	// Verify the keys are returned in the correct order with the updated product
	expectedKeys = []string{"prod3", "prod7", "prod2", "prod1"} // 150.00, 500.00, 900.00, 1200.00
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys after updating product, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d after updating product, got %s", expectedKeys[i], i, key)
		}
	}

	// Delete a product
	err = db.Delete("test_set", "prod7")
	if err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}

	// Verify the index is updated after deletion
	keys, err = sortableIndex.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index after deleting product: %v", err)
	}

	// Verify the keys are returned in the correct order without the deleted product
	expectedKeys = []string{"prod3", "prod2", "prod1"} // 150.00, 900.00, 1200.00
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys after deleting product, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d after deleting product, got %s", expectedKeys[i], i, key)
		}
	}

	// Test multi-field sorting
	keys, err = sortableIndex.QueryMultiSorted("Clothing", []string{"Price", "Stock"}, []bool{false, true})
	if err != nil {
		t.Fatalf("Failed to query sortable index with multi-field sorting: %v", err)
	}

	// Verify the keys are returned in the correct order (sorted by Price descending, then Stock ascending)
	expectedKeys = []string{"prod5", "prod4"} // 50.00, 25.00
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys for multi-field sorting, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d for multi-field sorting, got %s", expectedKeys[i], i, key)
		}
	}
}

// TestSortableIndexIntegrationWithMissingField tests the integration between sortable indexes and sets
// when values without the indexed fields are added to a set
func TestSortableIndexIntegrationWithMissingField(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Define types with and without sort fields
	type FullProduct struct {
		Name     string
		Category string
		Price    float64
		Stock    int
	}

	type PartialProduct struct {
		Name     string
		Category string
		// Missing Price
		Stock int
	}

	// Add products with and without Price
	err = db.Put("test_set", "prod1", FullProduct{Name: "Laptop", Category: "Electronics", Price: 1200.00, Stock: 10})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod2", FullProduct{Name: "Smartphone", Category: "Electronics", Price: 800.00, Stock: 15})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod3", PartialProduct{Name: "Headphones", Category: "Electronics", Stock: 30}) // Missing Price
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Create a sortable index on Category (primary) with Price (sort)
	sortableIndex, err := db.CreateSortableIndex("category_price_index", "test_set", "Category", []string{"Price"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// Test querying with sorting by Price
	keys, err := sortableIndex.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}

	// Verify the keys are returned in the correct order
	// Products with Price should be sorted first, then products without Price
	expectedKeysWithPrice := []string{"prod2", "prod1"} // 800.00, 1200.00
	expectedKeysWithoutPrice := []string{"prod3"}       // Missing Price

	if len(keys) != len(expectedKeysWithPrice)+len(expectedKeysWithoutPrice) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeysWithPrice)+len(expectedKeysWithoutPrice), len(keys))
	}

	// Check that products with Price are sorted correctly
	for i, key := range keys[:len(expectedKeysWithPrice)] {
		if key != expectedKeysWithPrice[i] {
			t.Errorf("Expected key %s at position %d, got %s", expectedKeysWithPrice[i], i, key)
		}
	}

	// Check that products without Price are at the end
	for i, key := range keys[len(expectedKeysWithPrice):] {
		if key != expectedKeysWithoutPrice[i] {
			t.Errorf("Expected key %s at position %d among products without Price, got %s", 
				expectedKeysWithoutPrice[i], i, key)
		}
	}

	// Add a new product without Price
	err = db.Put("test_set", "prod4", PartialProduct{Name: "Keyboard", Category: "Electronics", Stock: 25}) // Missing Price
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Verify the index is updated correctly
	keys, err = sortableIndex.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index after adding new product without Price: %v", err)
	}

	// Verify the keys are returned in the correct order
	expectedKeysWithPrice = []string{"prod2", "prod1"}         // 800.00, 1200.00
	expectedKeysWithoutPrice = []string{"prod3", "prod4"}      // Missing Price

	if len(keys) != len(expectedKeysWithPrice)+len(expectedKeysWithoutPrice) {
		t.Errorf("Expected %d keys after adding new product without Price, got %d", 
			len(expectedKeysWithPrice)+len(expectedKeysWithoutPrice), len(keys))
	}

	// Check that products with Price are sorted correctly
	for i, key := range keys[:len(expectedKeysWithPrice)] {
		if key != expectedKeysWithPrice[i] {
			t.Errorf("Expected key %s at position %d after adding new product without Price, got %s", 
				expectedKeysWithPrice[i], i, key)
		}
	}

	// Update a product to add the missing Price
	err = db.Put("test_set", "prod3", FullProduct{Name: "Headphones", Category: "Electronics", Price: 150.00, Stock: 30})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Verify the index is updated correctly
	keys, err = sortableIndex.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index after updating product to add Price: %v", err)
	}

	// Verify the keys are returned in the correct order
	expectedKeysWithPrice = []string{"prod3", "prod2", "prod1"} // 150.00, 800.00, 1200.00
	expectedKeysWithoutPrice = []string{"prod4"}                // Missing Price

	if len(keys) != len(expectedKeysWithPrice)+len(expectedKeysWithoutPrice) {
		t.Errorf("Expected %d keys after updating product to add Price, got %d", 
			len(expectedKeysWithPrice)+len(expectedKeysWithoutPrice), len(keys))
	}

	// Check that products with Price are sorted correctly
	for i, key := range keys[:len(expectedKeysWithPrice)] {
		if key != expectedKeysWithPrice[i] {
			t.Errorf("Expected key %s at position %d after updating product to add Price, got %s", 
				expectedKeysWithPrice[i], i, key)
		}
	}

	// Update a product to remove the Price
	err = db.Put("test_set", "prod1", PartialProduct{Name: "Laptop", Category: "Electronics", Stock: 10}) // Remove Price
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Verify the index is updated correctly
	keys, err = sortableIndex.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index after updating product to remove Price: %v", err)
	}

	// Verify the keys are returned in the correct order
	expectedKeysWithPrice = []string{"prod3", "prod2"}         // 150.00, 800.00
	expectedKeysWithoutPrice = []string{"prod1", "prod4"}      // Missing Price

	if len(keys) != len(expectedKeysWithPrice)+len(expectedKeysWithoutPrice) {
		t.Errorf("Expected %d keys after updating product to remove Price, got %d", 
			len(expectedKeysWithPrice)+len(expectedKeysWithoutPrice), len(keys))
	}

	// Check that products with Price are sorted correctly
	for i, key := range keys[:len(expectedKeysWithPrice)] {
		if key != expectedKeysWithPrice[i] {
			t.Errorf("Expected key %s at position %d after updating product to remove Price, got %s", 
				expectedKeysWithPrice[i], i, key)
		}
	}
}

// TestSortableIndexIntegrationWithCategoryChange tests the behavior when the primary field value changes
func TestSortableIndexIntegrationWithCategoryChange(t *testing.T) {
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

	// Add products in different categories
	err = db.Put("test_set", "prod1", Product{Name: "Laptop", Category: "Electronics", Price: 1200.00})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod2", Product{Name: "Smartphone", Category: "Electronics", Price: 800.00})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}
	
	err = db.Put("test_set", "prod3", Product{Name: "T-shirt", Category: "Clothing", Price: 25.00})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Create a sortable index on Category (primary) with Price (sort)
	sortableIndex, err := db.CreateSortableIndex("category_price_index", "test_set", "Category", []string{"Price"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// Verify initial state
	keys, err := sortableIndex.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}
	expectedKeys := []string{"prod2", "prod1"} // 800.00, 1200.00
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys for Electronics, got %d", len(expectedKeys), len(keys))
	}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d for Electronics, got %s", expectedKeys[i], i, key)
		}
	}

	keys, err = sortableIndex.QuerySorted("Clothing", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}
	if len(keys) != 1 || keys[0] != "prod3" {
		t.Errorf("Expected [prod3] for Clothing, got %v", keys)
	}

	// Change a product's category
	err = db.Put("test_set", "prod2", Product{Name: "Smartphone", Category: "Clothing", Price: 800.00})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Verify the index was updated for both categories
	keys, err = sortableIndex.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index after category change: %v", err)
	}
	if len(keys) != 1 || keys[0] != "prod1" {
		t.Errorf("Expected [prod1] for Electronics after category change, got %v", keys)
	}

	keys, err = sortableIndex.QuerySorted("Clothing", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index after category change: %v", err)
	}
	expectedKeys = []string{"prod3", "prod2"} // 25.00, 800.00
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys for Clothing after category change, got %d", len(expectedKeys), len(keys))
	}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d for Clothing after category change, got %s", 
				expectedKeys[i], i, key)
		}
	}

	// Change a product's category to a new category
	err = db.Put("test_set", "prod1", Product{Name: "Laptop", Category: "Computers", Price: 1200.00})
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Verify the index was updated for all categories
	keys, err = sortableIndex.QuerySorted("Electronics", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index after second category change: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected no keys for Electronics after second category change, got %v", keys)
	}

	keys, err = sortableIndex.QuerySorted("Computers", "Price", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index for new category: %v", err)
	}
	if len(keys) != 1 || keys[0] != "prod1" {
		t.Errorf("Expected [prod1] for Computers, got %v", keys)
	}
}