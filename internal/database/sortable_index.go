package database

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

// SortableIndex represents an index that supports sorting on multiple fields
type SortableIndex struct {
	Name        string
	SetName     string
	PrimaryField string
	SortFields  []string
	Values      map[string][]string                  // Map from primary field value to list of keys
	SortValues  map[string]map[string]interface{}    // Map from key to sort field values
	mu          sync.RWMutex
}

// NewSortableIndex creates a new sortable index
func NewSortableIndex(name string, setName string, primaryField string, sortFields []string) *SortableIndex {
	return &SortableIndex{
		Name:        name,
		SetName:     setName,
		PrimaryField: primaryField,
		SortFields:  sortFields,
		Values:      make(map[string][]string),
		SortValues:  make(map[string]map[string]interface{}),
	}
}

// Build builds the sortable index by scanning all entries in the set
func (idx *SortableIndex) Build(set *Set) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Clear existing index data
	idx.Values = make(map[string][]string)
	idx.SortValues = make(map[string]map[string]interface{})

	// Scan all entries in the set
	return set.ForEach(func(key string, value []byte) error {
		// Extract the primary field value from the MessagePack encoded data
		primaryValue, err := idx.extractFieldValue(value, idx.PrimaryField)
		if err != nil {
			// If the primary field is not found, silently skip this entry
			if err.Error() == fmt.Sprintf("field not found in data: %s", idx.PrimaryField) {
				return nil
			}
			return err
		}

		// Add the key to the index
		idx.Values[primaryValue] = append(idx.Values[primaryValue], key)

		// Extract sort field values
		sortValues := make(map[string]interface{})
		for _, sortField := range idx.SortFields {
			sortValue, err := idx.extractFieldValue(value, sortField)
			if err == nil {
				sortValues[sortField] = sortValue
			}
			// If sort field is not found, we just don't add it to sortValues
		}

		// Only store sort values if at least one sort field was found
		if len(sortValues) > 0 {
			idx.SortValues[key] = sortValues
		}

		return nil
	})
}

// extractFieldValue extracts the value of a field from MessagePack encoded data
func (idx *SortableIndex) extractFieldValue(data []byte, fieldName string) (string, error) {
	var m map[string]interface{}
	if err := msgpack.Unmarshal(data, &m); err != nil {
		return "", fmt.Errorf("failed to decode MessagePack data: %w", err)
	}

	// Get the field value
	value, ok := m[fieldName]
	if !ok {
		return "", fmt.Errorf("field not found in data: %s", fieldName)
	}

	// Convert the value to a string
	switch v := value.(type) {
	case string:
		return v, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("unsupported field type: %T", value)
	}
}

// AddEntry adds an entry to the sortable index
func (idx *SortableIndex) AddEntry(key string, value []byte) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Extract the primary field value
	primaryValue, err := idx.extractFieldValue(value, idx.PrimaryField)
	if err != nil {
		// If the primary field is not found, silently skip this entry
		if err.Error() == fmt.Sprintf("field not found in data: %s", idx.PrimaryField) {
			return nil
		}
		return err
	}

	// Add the key to the index
	idx.Values[primaryValue] = append(idx.Values[primaryValue], key)

	// Extract sort field values
	sortValues := make(map[string]interface{})
	for _, sortField := range idx.SortFields {
		sortValue, err := idx.extractFieldValue(value, sortField)
		if err == nil {
			sortValues[sortField] = sortValue
		}
		// If sort field is not found, we just don't add it to sortValues
	}

	// Only store sort values if at least one sort field was found
	if len(sortValues) > 0 {
		idx.SortValues[key] = sortValues
	}

	return nil
}

// RemoveEntry removes an entry from the sortable index
func (idx *SortableIndex) RemoveEntry(key string, value []byte) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Extract the primary field value
	primaryValue, err := idx.extractFieldValue(value, idx.PrimaryField)
	if err != nil {
		// If the primary field is not found, silently skip this operation
		if err.Error() == fmt.Sprintf("field not found in data: %s", idx.PrimaryField) {
			return nil
		}
		return err
	}

	// Remove the key from the index
	keys, ok := idx.Values[primaryValue]
	if !ok {
		return nil
	}

	newKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		if k != key {
			newKeys = append(newKeys, k)
		}
	}

	if len(newKeys) == 0 {
		delete(idx.Values, primaryValue)
	} else {
		idx.Values[primaryValue] = newKeys
	}

	// Remove sort values for this key
	delete(idx.SortValues, key)

	return nil
}

// UpdateEntry updates an entry in the sortable index
func (idx *SortableIndex) UpdateEntry(key string, oldValue, newValue []byte) error {
	// Use RemoveEntry and AddEntry to ensure consistent behavior
	// This is more reliable than trying to update in place
	if err := idx.RemoveEntry(key, oldValue); err != nil {
		// If removal fails but it's not because the field is missing, return the error
		if err.Error() != fmt.Sprintf("field not found in data: %s", idx.PrimaryField) {
			return err
		}
		// Otherwise, continue with adding the new entry
	}
	
	return idx.AddEntry(key, newValue)
}

// Query queries the sortable index for keys matching the given primary value
func (idx *SortableIndex) Query(value string) ([]string, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Get the keys for the value
	keys, ok := idx.Values[value]
	if !ok {
		return []string{}, nil
	}

	// Make a copy of the keys to avoid modifying the original
	result := make([]string, len(keys))
	copy(result, keys)

	return result, nil
}

// QuerySorted queries the sortable index for keys matching the given primary value
// and sorts the results by the specified sort field
func (idx *SortableIndex) QuerySorted(value string, sortField string, ascending bool) ([]string, error) {
	// First get all matching keys
	keys, err := idx.Query(value)
	if err != nil {
		return nil, err
	}

	// If no keys found or sort field is not in the index, return as is
	if len(keys) == 0 || !idx.containsSortField(sortField) {
		return keys, nil
	}

	// Sort the keys based on the sort field
	return idx.sortKeysBySingleField(keys, sortField, ascending), nil
}

// QuerySortedWithPagination queries the sortable index with sorting and pagination
func (idx *SortableIndex) QuerySortedWithPagination(value string, sortField string, ascending bool, offset int, limit int) ([]string, error) {
	// First get sorted keys
	keys, err := idx.QuerySorted(value, sortField, ascending)
	if err != nil {
		return nil, err
	}

	// Apply pagination
	return idx.applyPagination(keys, offset, limit), nil
}

// QueryMultiSorted queries the sortable index for keys matching the given primary value
// and sorts the results by multiple sort fields
func (idx *SortableIndex) QueryMultiSorted(value string, sortFields []string, ascending []bool) ([]string, error) {
	// First get all matching keys
	keys, err := idx.Query(value)
	if err != nil {
		return nil, err
	}

	// If no keys found or no sort fields specified, return as is
	if len(keys) == 0 || len(sortFields) == 0 {
		return keys, nil
	}

	// Sort the keys based on multiple sort fields
	return idx.sortKeysByMultipleFields(keys, sortFields, ascending), nil
}

// QueryMultiSortedWithPagination queries the sortable index with multi-field sorting and pagination
func (idx *SortableIndex) QueryMultiSortedWithPagination(value string, sortFields []string, ascending []bool, offset int, limit int) ([]string, error) {
	// First get multi-sorted keys
	keys, err := idx.QueryMultiSorted(value, sortFields, ascending)
	if err != nil {
		return nil, err
	}

	// Apply pagination
	return idx.applyPagination(keys, offset, limit), nil
}

// applyPagination applies offset and limit to a list of keys
func (idx *SortableIndex) applyPagination(keys []string, offset int, limit int) []string {
	// Check if offset is beyond the available results
	if offset >= len(keys) {
		return []string{}
	}

	// Calculate end index
	end := offset + limit
	if end > len(keys) {
		end = len(keys)
	}

	// Return the paginated slice
	return keys[offset:end]
}

// containsSortField checks if the sort field is in the index
func (idx *SortableIndex) containsSortField(sortField string) bool {
	for _, field := range idx.SortFields {
		if field == sortField {
			return true
		}
	}
	return false
}

// sortKeysBySingleField sorts keys by a single sort field
func (idx *SortableIndex) sortKeysBySingleField(keys []string, sortField string, ascending bool) []string {
	// Create a copy of the keys to sort
	result := make([]string, len(keys))
	copy(result, keys)

	// Separate keys with and without sort values
	keysWithSortValue := make([]string, 0, len(keys))
	keysWithoutSortValue := make([]string, 0, len(keys))

	for _, key := range result {
		sortValues, ok := idx.SortValues[key]
		if ok && sortValues[sortField] != nil {
			keysWithSortValue = append(keysWithSortValue, key)
		} else {
			keysWithoutSortValue = append(keysWithoutSortValue, key)
		}
	}

	// Sort the keys with sort values
	idx.mu.RLock()
	sort.SliceStable(keysWithSortValue, func(i, j int) bool {
		keyI := keysWithSortValue[i]
		keyJ := keysWithSortValue[j]
		valueI := idx.SortValues[keyI][sortField]
		valueJ := idx.SortValues[keyJ][sortField]

		// Compare based on the type of the values
		result := idx.compareValues(valueI, valueJ)
		if !ascending {
			result = -result
		}
		return result < 0
	})
	idx.mu.RUnlock()

	// Combine sorted keys with keys without sort values
	sortedKeys := append(keysWithSortValue, keysWithoutSortValue...)
	return sortedKeys
}

// sortKeysByMultipleFields sorts keys by multiple sort fields
func (idx *SortableIndex) sortKeysByMultipleFields(keys []string, sortFields []string, ascending []bool) []string {
	// Create a copy of the keys to sort
	result := make([]string, len(keys))
	copy(result, keys)

	// Ensure ascending has the same length as sortFields
	asc := make([]bool, len(sortFields))
	copy(asc, ascending)
	for i := len(ascending); i < len(sortFields); i++ {
		asc = append(asc, true) // Default to ascending for missing values
	}

	// Separate keys with and without any sort values
	keysWithAnySortValue := make([]string, 0, len(keys))
	keysWithoutAnySortValue := make([]string, 0, len(keys))

	for _, key := range result {
		sortValues, ok := idx.SortValues[key]
		if ok && len(sortValues) > 0 {
			keysWithAnySortValue = append(keysWithAnySortValue, key)
		} else {
			keysWithoutAnySortValue = append(keysWithoutAnySortValue, key)
		}
	}

	// Sort the keys with any sort values
	idx.mu.RLock()
	sort.SliceStable(keysWithAnySortValue, func(i, j int) bool {
		keyI := keysWithAnySortValue[i]
		keyJ := keysWithAnySortValue[j]

		// Compare each sort field in order
		for k, sortField := range sortFields {
			valueI, okI := idx.SortValues[keyI][sortField]
			valueJ, okJ := idx.SortValues[keyJ][sortField]

			// If either key doesn't have this sort field, continue to the next field
			if !okI || !okJ {
				continue
			}

			// Compare the values
			result := idx.compareValues(valueI, valueJ)
			if result != 0 {
				if asc[k] {
					return result < 0
				}
				return result > 0
			}
		}

		// If all fields are equal, maintain original order
		return i < j
	})
	idx.mu.RUnlock()

	// Combine sorted keys with keys without any sort values
	sortedKeys := append(keysWithAnySortValue, keysWithoutAnySortValue...)
	return sortedKeys
}

// compareValues compares two values and returns:
// -1 if valueI < valueJ
// 0 if valueI == valueJ
// 1 if valueI > valueJ
func (idx *SortableIndex) compareValues(valueI, valueJ interface{}) int {
	// Handle nil values
	if valueI == nil && valueJ == nil {
		return 0
	}
	if valueI == nil {
		return -1
	}
	if valueJ == nil {
		return 1
	}

	// Try to convert both values to float64 for numeric comparison
	fI, fIOk := idx.toFloat64(valueI)
	fJ, fJOk := idx.toFloat64(valueJ)
	
	// If both values can be converted to float64, compare them as numbers
	if fIOk && fJOk {
		if fI < fJ {
			return -1
		} else if fI > fJ {
			return 1
		}
		return 0
	}
	
	// Otherwise, compare based on the type
	switch vI := valueI.(type) {
	case string:
		if vJ, ok := valueJ.(string); ok {
			if vI < vJ {
				return -1
			} else if vI > vJ {
				return 1
			}
			return 0
		}
	case bool:
		if vJ, ok := valueJ.(bool); ok {
			if !vI && vJ {
				return -1
			} else if vI && !vJ {
				return 1
			}
			return 0
		}
	}

	// If types are different or not comparable, compare string representations
	sI := fmt.Sprintf("%v", valueI)
	sJ := fmt.Sprintf("%v", valueJ)
	if sI < sJ {
		return -1
	} else if sI > sJ {
		return 1
	}
	return 0
}

// toFloat64 attempts to convert a value to float64
func (idx *SortableIndex) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		// Try to parse the string as a number
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// GetAllValues returns all unique primary field values in the index
func (idx *SortableIndex) GetAllValues() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	values := make([]string, 0, len(idx.Values))
	for value := range idx.Values {
		values = append(values, value)
	}

	return values
}

// Size returns the number of unique primary field values in the index
func (idx *SortableIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.Values)
}

// Clear clears the index
func (idx *SortableIndex) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.Values = make(map[string][]string)
	idx.SortValues = make(map[string]map[string]interface{})
}

// GetName returns the name of the index
func (idx *SortableIndex) GetName() string {
	return idx.Name
}

// GetSetName returns the name of the set this index is for
func (idx *SortableIndex) GetSetName() string {
	return idx.SetName
}

// GetField returns the primary field this index is on
func (idx *SortableIndex) GetField() string {
	return idx.PrimaryField
}

// GetType returns the type of this index
func (idx *SortableIndex) GetType() IndexType {
	return SortableIndexType
}