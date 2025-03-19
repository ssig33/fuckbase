package database

import (
	"sort"
	"testing"
)

// TestSortableIndexPagination tests the pagination functionality of a sortable index
func TestSortableIndexPagination(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Add some data with multiple fields
	type Employee struct {
		Name       string
		Department string
		HireDate   string
		Salary     int
	}

	// Add multiple employees with the same department
	err = db.Put("test_set", "emp1", Employee{Name: "Alice", Department: "Engineering", HireDate: "2022-05-15", Salary: 85000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp2", Employee{Name: "Bob", Department: "Engineering", HireDate: "2021-11-03", Salary: 95000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp3", Employee{Name: "Charlie", Department: "Engineering", HireDate: "2023-01-20", Salary: 75000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp4", Employee{Name: "David", Department: "Engineering", HireDate: "2022-03-10", Salary: 80000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp5", Employee{Name: "Eve", Department: "Engineering", HireDate: "2021-07-22", Salary: 90000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp6", Employee{Name: "Frank", Department: "Engineering", HireDate: "2022-09-05", Salary: 82000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp7", Employee{Name: "Grace", Department: "Engineering", HireDate: "2023-02-14", Salary: 78000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp8", Employee{Name: "Hannah", Department: "Engineering", HireDate: "2021-12-01", Salary: 88000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp9", Employee{Name: "Ian", Department: "Engineering", HireDate: "2022-06-30", Salary: 86000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp10", Employee{Name: "Julia", Department: "Engineering", HireDate: "2023-03-15", Salary: 79000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}

	// Create a sortable index on Department (primary) and HireDate (sort)
	sortableIndex, err := db.CreateSortableIndex("dept_hire_index", "test_set", "Department", []string{"HireDate"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// Test pagination with offset 0, limit 3
	keys, err := sortableIndex.QuerySortedWithPagination("Engineering", "HireDate", true, 0, 3)
	if err != nil {
		t.Fatalf("Failed to query sortable index with pagination: %v", err)
	}

	// Verify the keys are returned in the correct order with correct pagination
	// First page (offset 0, limit 3) should return the first 3 employees sorted by HireDate ascending
	expectedKeys := []string{"emp5", "emp2", "emp8"} // 2021-07-22, 2021-11-03, 2021-12-01
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys for first page, got %d", len(expectedKeys), len(keys))
	}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d for first page, got %s", expectedKeys[i], i, key)
		}
	}

	// Test second page (offset 3, limit 3)
	keys, err = sortableIndex.QuerySortedWithPagination("Engineering", "HireDate", true, 3, 3)
	if err != nil {
		t.Fatalf("Failed to query sortable index with pagination for second page: %v", err)
	}

	// Verify the second page
	expectedKeys = []string{"emp4", "emp1", "emp9"} // 2022-03-10, 2022-05-15, 2022-06-30
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys for second page, got %d", len(expectedKeys), len(keys))
	}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d for second page, got %s", expectedKeys[i], i, key)
		}
	}

	// Test third page (offset 6, limit 3)
	keys, err = sortableIndex.QuerySortedWithPagination("Engineering", "HireDate", true, 6, 3)
	if err != nil {
		t.Fatalf("Failed to query sortable index with pagination for third page: %v", err)
	}

	// Verify the third page
	expectedKeys = []string{"emp6", "emp3", "emp7"} // 2022-09-05, 2023-01-20, 2023-02-14
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys for third page, got %d", len(expectedKeys), len(keys))
	}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d for third page, got %s", expectedKeys[i], i, key)
		}
	}

	// Test fourth page (offset 9, limit 3) - should return only one result
	keys, err = sortableIndex.QuerySortedWithPagination("Engineering", "HireDate", true, 9, 3)
	if err != nil {
		t.Fatalf("Failed to query sortable index with pagination for fourth page: %v", err)
	}

	// Verify the fourth page
	expectedKeys = []string{"emp10"} // 2023-03-15
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys for fourth page, got %d", len(expectedKeys), len(keys))
	}
	if len(keys) > 0 && keys[0] != expectedKeys[0] {
		t.Errorf("Expected key %s for fourth page, got %s", expectedKeys[0], keys[0])
	}

	// Test with offset beyond the available results
	keys, err = sortableIndex.QuerySortedWithPagination("Engineering", "HireDate", true, 20, 3)
	if err != nil {
		t.Fatalf("Failed to query sortable index with pagination beyond available results: %v", err)
	}

	// Should return empty result
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys for offset beyond available results, got %d", len(keys))
	}

	// Test with descending order
	keys, err = sortableIndex.QuerySortedWithPagination("Engineering", "HireDate", false, 0, 3)
	if err != nil {
		t.Fatalf("Failed to query sortable index with pagination and descending order: %v", err)
	}

	// Verify the first page with descending order
	expectedKeys = []string{"emp10", "emp7", "emp3"} // 2023-03-15, 2023-02-14, 2023-01-20
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys for first page with descending order, got %d", len(expectedKeys), len(keys))
	}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d for first page with descending order, got %s", expectedKeys[i], i, key)
		}
	}

	// Test with multi-field pagination
	sortableIndex, err = db.CreateSortableIndex("dept_salary_index", "test_set", "Department", []string{"Salary"})
	if err != nil {
		t.Fatalf("Failed to create sortable index for multi-field pagination: %v", err)
	}

	// Test pagination with multi-field sorting
	keys, err = sortableIndex.QueryMultiSortedWithPagination("Engineering", []string{"Salary"}, []bool{true}, 0, 5)
	if err != nil {
		t.Fatalf("Failed to query sortable index with multi-field pagination: %v", err)
	}

	// Verify the first page with multi-field sorting
	expectedKeys = []string{"emp3", "emp7", "emp10", "emp4", "emp6"} // Sorted by Salary ascending
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys for first page with multi-field sorting, got %d", len(expectedKeys), len(keys))
	}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d for first page with multi-field sorting, got %s", expectedKeys[i], i, key)
		}
	}
}

// TestSortableIndexBasic tests the basic functionality of a sortable index
func TestSortableIndexBasic(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Add some data with multiple fields
	type Employee struct {
		Name       string
		Department string
		HireDate   string
		Salary     int
		Position   string
	}

	// Add employees with different departments and hire dates
	err = db.Put("test_set", "emp1", Employee{Name: "Alice", Department: "Engineering", HireDate: "2022-05-15", Salary: 85000, Position: "Software Engineer"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp2", Employee{Name: "Bob", Department: "Engineering", HireDate: "2021-11-03", Salary: 95000, Position: "Senior Engineer"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp3", Employee{Name: "Charlie", Department: "Engineering", HireDate: "2023-01-20", Salary: 75000, Position: "Junior Engineer"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp4", Employee{Name: "David", Department: "Sales", HireDate: "2022-03-10", Salary: 80000, Position: "Sales Representative"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp5", Employee{Name: "Eve", Department: "Sales", HireDate: "2021-07-22", Salary: 90000, Position: "Sales Manager"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp6", Employee{Name: "Frank", Department: "Marketing", HireDate: "2022-09-05", Salary: 82000, Position: "Marketing Specialist"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}

	// Create a sortable index on Department (primary) and HireDate (sort)
	sortableIndex, err := db.CreateSortableIndex("dept_hire_index", "test_set", "Department", []string{"HireDate"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// Test querying with sorting by HireDate in ascending order
	keys, err := sortableIndex.QuerySorted("Engineering", "HireDate", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}

	// Verify the keys are returned in the correct order (sorted by HireDate ascending)
	expectedKeys := []string{"emp2", "emp1", "emp3"} // Sorted by HireDate: 2021-11-03, 2022-05-15, 2023-01-20
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d, got %s", expectedKeys[i], i, key)
		}
	}

	// Test querying with sorting by HireDate in descending order
	keys, err = sortableIndex.QuerySorted("Engineering", "HireDate", false)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}

	// Verify the keys are returned in the correct order (sorted by HireDate descending)
	expectedKeys = []string{"emp3", "emp1", "emp2"} // Sorted by HireDate: 2023-01-20, 2022-05-15, 2021-11-03
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d, got %s", expectedKeys[i], i, key)
		}
	}

	// Test querying for a different department
	keys, err = sortableIndex.QuerySorted("Sales", "HireDate", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}

	// Verify the keys are returned in the correct order
	expectedKeys = []string{"emp5", "emp4"} // Sorted by HireDate: 2021-07-22, 2022-03-10
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d, got %s", expectedKeys[i], i, key)
		}
	}

	// Test querying for a department with only one entry
	keys, err = sortableIndex.QuerySorted("Marketing", "HireDate", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}

	expectedKeys = []string{"emp6"}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	if len(keys) > 0 && keys[0] != expectedKeys[0] {
		t.Errorf("Expected key %s, got %s", expectedKeys[0], keys[0])
	}

	// Test querying for a non-existent department
	keys, err = sortableIndex.QuerySorted("HR", "HireDate", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}

	if len(keys) != 0 {
		t.Errorf("Expected 0 keys for non-existent department, got %d", len(keys))
	}
}

// TestSortableIndexMultiField tests sorting by multiple fields
func TestSortableIndexMultiField(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Add some data with multiple fields
	type Employee struct {
		Name       string
		Department string
		HireDate   string
		Salary     int
		Position   string
	}

	// Add employees with the same department but different positions and hire dates
	err = db.Put("test_set", "emp1", Employee{Name: "Alice", Department: "Engineering", HireDate: "2022-05-15", Salary: 85000, Position: "Senior Engineer"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp2", Employee{Name: "Bob", Department: "Engineering", HireDate: "2021-11-03", Salary: 95000, Position: "Senior Engineer"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp3", Employee{Name: "Charlie", Department: "Engineering", HireDate: "2023-01-20", Salary: 75000, Position: "Junior Engineer"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp4", Employee{Name: "David", Department: "Engineering", HireDate: "2022-03-10", Salary: 80000, Position: "Junior Engineer"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp5", Employee{Name: "Eve", Department: "Engineering", HireDate: "2021-07-22", Salary: 90000, Position: "Lead Engineer"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}

	// Create a sortable index on Department (primary) and Position, HireDate (sort)
	sortableIndex, err := db.CreateSortableIndex("dept_pos_hire_index", "test_set", "Department", []string{"Position", "HireDate"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// Test querying with multi-field sorting (Position ascending, then HireDate descending)
	keys, err := sortableIndex.QueryMultiSorted("Engineering", []string{"Position", "HireDate"}, []bool{true, false})
	if err != nil {
		t.Fatalf("Failed to query sortable index with multi-field sorting: %v", err)
	}

	// Verify the keys are returned in the correct order
	// First Junior Engineers sorted by HireDate desc, then Lead Engineer, then Senior Engineers sorted by HireDate desc
	expectedKeys := []string{"emp3", "emp4", "emp5", "emp1", "emp2"}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d, got %s", expectedKeys[i], i, key)
		}
	}

	// Test with different sort orders
	keys, err = sortableIndex.QueryMultiSorted("Engineering", []string{"Position", "HireDate"}, []bool{false, true})
	if err != nil {
		t.Fatalf("Failed to query sortable index with multi-field sorting: %v", err)
	}

	// Verify the keys are returned in the correct order
	// First Senior Engineers sorted by HireDate asc, then Lead Engineer, then Junior Engineers sorted by HireDate asc
	expectedKeys = []string{"emp2", "emp1", "emp5", "emp4", "emp3"}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d, got %s", expectedKeys[i], i, key)
		}
	}
}

// TestSortableIndexMissingField tests the behavior when entries are missing sort fields
func TestSortableIndexMissingField(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Add some data with and without sort fields
	type FullEmployee struct {
		Name       string
		Department string
		HireDate   string
		Salary     int
	}

	type PartialEmployee struct {
		Name       string
		Department string
		// Missing HireDate
		Salary int
	}

	// Add employees with and without HireDate
	err = db.Put("test_set", "emp1", FullEmployee{Name: "Alice", Department: "Engineering", HireDate: "2022-05-15", Salary: 85000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp2", FullEmployee{Name: "Bob", Department: "Engineering", HireDate: "2021-11-03", Salary: 95000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp3", PartialEmployee{Name: "Charlie", Department: "Engineering", Salary: 75000}) // Missing HireDate
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp4", FullEmployee{Name: "David", Department: "Engineering", HireDate: "2023-01-20", Salary: 80000})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp5", PartialEmployee{Name: "Eve", Department: "Engineering", Salary: 90000}) // Missing HireDate
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}

	// Create a sortable index on Department (primary) and HireDate (sort)
	sortableIndex, err := db.CreateSortableIndex("dept_hire_index", "test_set", "Department", []string{"HireDate"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// Test querying with sorting by HireDate
	keys, err := sortableIndex.QuerySorted("Engineering", "HireDate", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}

	// Verify the keys are returned in the correct order
	// Entries with HireDate should be sorted first, then entries without HireDate
	expectedKeysWithHireDate := []string{"emp2", "emp1", "emp4"} // Sorted by HireDate
	expectedKeysWithoutHireDate := []string{"emp3", "emp5"}      // No particular order, just at the end

	if len(keys) != len(expectedKeysWithHireDate)+len(expectedKeysWithoutHireDate) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeysWithHireDate)+len(expectedKeysWithoutHireDate), len(keys))
	}

	// Check that entries with HireDate are sorted correctly
	for i, key := range keys[:len(expectedKeysWithHireDate)] {
		if key != expectedKeysWithHireDate[i] {
			t.Errorf("Expected key %s at position %d, got %s", expectedKeysWithHireDate[i], i, key)
		}
	}

	// Check that entries without HireDate are at the end
	remainingKeys := keys[len(expectedKeysWithHireDate):]
	sort.Strings(remainingKeys) // Sort for consistent comparison
	sort.Strings(expectedKeysWithoutHireDate)

	for i, key := range remainingKeys {
		if key != expectedKeysWithoutHireDate[i] {
			t.Errorf("Expected key %s at position %d among entries without HireDate, got %s", 
				expectedKeysWithoutHireDate[i], i, key)
		}
	}
}

// TestSortableIndexUpdate tests updating entries and ensuring the index is updated correctly
func TestSortableIndexUpdate(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Create a set
	_, err := db.CreateSet("test_set")
	if err != nil {
		t.Fatalf("Failed to create set: %v", err)
	}

	// Add some initial data
	type Employee struct {
		Name       string
		Department string
		HireDate   string
	}

	// Add initial employees
	err = db.Put("test_set", "emp1", Employee{Name: "Alice", Department: "Engineering", HireDate: "2022-05-15"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}
	
	err = db.Put("test_set", "emp2", Employee{Name: "Bob", Department: "Sales", HireDate: "2021-11-03"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}

	// Create a sortable index
	sortableIndex, err := db.CreateSortableIndex("dept_hire_index", "test_set", "Department", []string{"HireDate"})
	if err != nil {
		t.Fatalf("Failed to create sortable index: %v", err)
	}

	// Verify initial state
	keys, err := sortableIndex.QuerySorted("Engineering", "HireDate", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}
	if len(keys) != 1 || keys[0] != "emp1" {
		t.Errorf("Expected [emp1], got %v", keys)
	}

	keys, err = sortableIndex.QuerySorted("Sales", "HireDate", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}
	if len(keys) != 1 || keys[0] != "emp2" {
		t.Errorf("Expected [emp2], got %v", keys)
	}

	// Update an entry - change department
	err = db.Put("test_set", "emp1", Employee{Name: "Alice", Department: "Sales", HireDate: "2022-05-15"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}

	// Verify the index was updated
	keys, err = sortableIndex.QuerySorted("Engineering", "HireDate", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected no keys for Engineering, got %v", keys)
	}

	keys, err = sortableIndex.QuerySorted("Sales", "HireDate", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys for Sales, got %d", len(keys))
	}

	// The keys should be sorted by HireDate
	expectedKeys := []string{"emp2", "emp1"} // 2021-11-03, 2022-05-15
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d, got %s", expectedKeys[i], i, key)
		}
	}

	// Update an entry - change hire date
	err = db.Put("test_set", "emp2", Employee{Name: "Bob", Department: "Sales", HireDate: "2023-01-10"})
	if err != nil {
		t.Fatalf("Failed to put employee: %v", err)
	}

	// Verify the sort order was updated
	keys, err = sortableIndex.QuerySorted("Sales", "HireDate", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}

	// The keys should now be in a different order
	expectedKeys = []string{"emp1", "emp2"} // 2022-05-15, 2023-01-10
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key %s at position %d, got %s", expectedKeys[i], i, key)
		}
	}

	// Delete an entry
	err = db.Delete("test_set", "emp1")
	if err != nil {
		t.Fatalf("Failed to delete employee: %v", err)
	}

	// Verify the index was updated
	keys, err = sortableIndex.QuerySorted("Sales", "HireDate", true)
	if err != nil {
		t.Fatalf("Failed to query sortable index: %v", err)
	}
	if len(keys) != 1 || keys[0] != "emp2" {
		t.Errorf("Expected [emp2], got %v", keys)
	}
}