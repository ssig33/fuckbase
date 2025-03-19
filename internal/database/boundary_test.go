package database

import (
	"fmt"
	"strings"
	"testing"
)

// TestDatabaseBoundaryConditions tests boundary conditions for database operations
func TestDatabaseBoundaryConditions(t *testing.T) {
	// Create a new database
	db := NewDatabase("test_db", nil)

	// Test 1: Very long set name
	t.Run("VeryLongSetName", func(t *testing.T) {
		// Create a set with a very long name (1000 characters)
		longName := strings.Repeat("a", 1000)
		set, err := db.CreateSet(longName)
		if err != nil {
			t.Fatalf("Failed to create set with long name: %v", err)
		}

		// Verify the set was created with the correct name
		if set.Name != longName {
			t.Errorf("Expected set name to be %s, got %s", longName, set.Name)
		}

		// Verify we can get the set by its long name
		retrievedSet, err := db.GetSet(longName)
		if err != nil {
			t.Errorf("Failed to get set with long name: %v", err)
		}
		if retrievedSet.Name != longName {
			t.Errorf("Expected retrieved set name to be %s, got %s", longName, retrievedSet.Name)
		}
	})

	// Test 2: Very long key name
	t.Run("VeryLongKeyName", func(t *testing.T) {
		// Create a set
		set, err := db.CreateSet("test_set")
		if err != nil {
			t.Fatalf("Failed to create set: %v", err)
		}

		// Create a key with a very long name (1000 characters)
		longKey := strings.Repeat("k", 1000)
		value := "test_value"

		// Put the value with the long key
		err = set.Put(longKey, value)
		if err != nil {
			t.Fatalf("Failed to put value with long key: %v", err)
		}

		// Verify we can get the value by its long key
		var retrievedValue string
		err = set.Get(longKey, &retrievedValue)
		if err != nil {
			t.Errorf("Failed to get value with long key: %v", err)
		}
		if retrievedValue != value {
			t.Errorf("Expected retrieved value to be %s, got %s", value, retrievedValue)
		}
	})

	// Test 3: Very large value
	t.Run("VeryLargeValue", func(t *testing.T) {
		// Create a set
		set, err := db.CreateSet("large_value_set")
		if err != nil {
			t.Fatalf("Failed to create set: %v", err)
		}

		// Create a large value (1MB string)
		largeValue := strings.Repeat("v", 1024*1024)

		// Put the large value
		err = set.Put("large_key", largeValue)
		if err != nil {
			t.Fatalf("Failed to put large value: %v", err)
		}

		// Verify we can get the large value
		var retrievedValue string
		err = set.Get("large_key", &retrievedValue)
		if err != nil {
			t.Errorf("Failed to get large value: %v", err)
		}
		if len(retrievedValue) != len(largeValue) {
			t.Errorf("Expected retrieved value length to be %d, got %d", len(largeValue), len(retrievedValue))
		}
		if retrievedValue != largeValue {
			t.Errorf("Retrieved value does not match original large value")
		}
	})

	// Test 4: Many keys in a set
	t.Run("ManyKeysInSet", func(t *testing.T) {
		// Create a set
		set, err := db.CreateSet("many_keys_set")
		if err != nil {
			t.Fatalf("Failed to create set: %v", err)
		}

		// Add many keys (1000)
		numKeys := 1000
		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("key%d", i)
			value := fmt.Sprintf("value%d", i)
			err := set.Put(key, value)
			if err != nil {
				t.Fatalf("Failed to put value for key %s: %v", key, err)
			}
		}

		// Verify the set size
		if set.Size() != numKeys {
			t.Errorf("Expected set size to be %d, got %d", numKeys, set.Size())
		}

		// Verify we can get all keys
		keys := set.Keys()
		if len(keys) != numKeys {
			t.Errorf("Expected %d keys, got %d", numKeys, len(keys))
		}

		// Verify we can get all values
		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("key%d", i)
			expectedValue := fmt.Sprintf("value%d", i)
			var retrievedValue string
			err := set.Get(key, &retrievedValue)
			if err != nil {
				t.Errorf("Failed to get value for key %s: %v", key, err)
				continue
			}
			if retrievedValue != expectedValue {
				t.Errorf("Expected value for key %s to be %s, got %s", key, expectedValue, retrievedValue)
			}
		}
	})

	// Test 5: Many sets in a database
	t.Run("ManySetsInDatabase", func(t *testing.T) {
		// Create many sets (100)
		numSets := 100
		for i := 0; i < numSets; i++ {
			setName := fmt.Sprintf("set%d", i)
			_, err := db.CreateSet(setName)
			if err != nil {
				t.Fatalf("Failed to create set %s: %v", setName, err)
			}
		}

		// Verify we can list all sets
		sets := db.ListSets()
		
		// Count the sets that match our naming pattern (there might be other sets from other tests)
		patternCount := 0
		for _, setName := range sets {
			if strings.HasPrefix(setName, "set") {
				patternCount++
			}
		}
		
		if patternCount < numSets {
			t.Errorf("Expected at least %d sets with pattern 'set', got %d", numSets, patternCount)
		}

		// Verify we can get each set
		for i := 0; i < numSets; i++ {
			setName := fmt.Sprintf("set%d", i)
			set, err := db.GetSet(setName)
			if err != nil {
				t.Errorf("Failed to get set %s: %v", setName, err)
				continue
			}
			if set.Name != setName {
				t.Errorf("Expected set name to be %s, got %s", setName, set.Name)
			}
		}
	})

	// Test 6: Many indexes in a database
	t.Run("ManyIndexesInDatabase", func(t *testing.T) {
		// Create a set for indexing
		set, err := db.CreateSet("index_test_set")
		if err != nil {
			t.Fatalf("Failed to create set: %v", err)
		}

		// Add some data
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key%d", i)
			value := map[string]interface{}{
				"field0": fmt.Sprintf("value%d", i),
				"field1": i,
				"field2": i * 2,
				"field3": fmt.Sprintf("tag%d", i % 3),
				"field4": i % 2 == 0,
			}
			err := set.Put(key, value)
			if err != nil {
				t.Fatalf("Failed to put value: %v", err)
			}
		}

		// Create many indexes (50)
		numIndexes := 50
		for i := 0; i < numIndexes; i++ {
			fieldName := fmt.Sprintf("field%d", i % 5) // Use one of 5 fields
			indexName := fmt.Sprintf("index%d", i)
			_, err := db.CreateIndex(indexName, "index_test_set", fieldName)
			if err != nil {
				t.Fatalf("Failed to create index %s: %v", indexName, err)
			}
		}

		// Verify we can list all indexes
		indexes := db.ListIndexes()
		if len(indexes) < numIndexes {
			t.Errorf("Expected at least %d indexes, got %d", numIndexes, len(indexes))
		}

		// Verify we can get each index
		for i := 0; i < numIndexes; i++ {
			indexName := fmt.Sprintf("index%d", i)
			index, err := db.GetIndex(indexName)
			if err != nil {
				t.Errorf("Failed to get index %s: %v", indexName, err)
				continue
			}
			if index.GetName() != indexName {
				t.Errorf("Expected index name to be %s, got %s", indexName, index.GetName())
			}
		}
	})
}