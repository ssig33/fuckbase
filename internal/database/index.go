package database

import (
	"fmt"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

// BasicIndex represents a basic index on a single field in a set
type BasicIndex struct {
	Name    string
	SetName string
	Field   string
	Values  map[string][]string // Map from field value to list of keys
	mu      sync.RWMutex
}

// NewIndex creates a new basic index
func NewIndex(name string, setName string, field string) *BasicIndex {
	return &BasicIndex{
		Name:    name,
		SetName: setName,
		Field:   field,
		Values:  make(map[string][]string),
	}
}

// Build builds the index by scanning all entries in the set
// Entries without the indexed field are silently skipped
func (idx *BasicIndex) Build(set *Set) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Clear existing index data
	idx.Values = make(map[string][]string)

	// Scan all entries in the set
	return set.ForEach(func(key string, value []byte) error {
		// Extract the field value from the MessagePack encoded data
		fieldValue, err := idx.extractFieldValue(value)
		if err != nil {
			// If the field is not found, silently skip this entry
			if err.Error() == fmt.Sprintf("field not found in data: %s", idx.Field) {
				return nil
			}
			return err
		}

		// Add the key to the index
		idx.Values[fieldValue] = append(idx.Values[fieldValue], key)
		return nil
	})
}

// extractFieldValue extracts the value of the indexed field from MessagePack encoded data
func (idx *BasicIndex) extractFieldValue(data []byte) (string, error) {
	var m map[string]interface{}
	if err := msgpack.Unmarshal(data, &m); err != nil {
		return "", fmt.Errorf("failed to decode MessagePack data: %w", err)
	}

	// Get the field value
	value, ok := m[idx.Field]
	if !ok {
		return "", fmt.Errorf("field not found in data: %s", idx.Field)
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

// AddEntry adds an entry to the index
// If the field is not found in the data, the entry is silently skipped (not added to the index)
func (idx *BasicIndex) AddEntry(key string, value []byte) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Extract the field value
	fieldValue, err := idx.extractFieldValue(value)
	if err != nil {
		// If the field is not found, silently skip this entry
		if err.Error() == fmt.Sprintf("field not found in data: %s", idx.Field) {
			return nil
		}
		return err
	}

	// Add the key to the index
	idx.Values[fieldValue] = append(idx.Values[fieldValue], key)
	return nil
}

// RemoveEntry removes an entry from the index
// If the field is not found in the data, the operation is silently skipped
func (idx *BasicIndex) RemoveEntry(key string, value []byte) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Extract the field value
	fieldValue, err := idx.extractFieldValue(value)
	if err != nil {
		// If the field is not found, silently skip this operation
		if err.Error() == fmt.Sprintf("field not found in data: %s", idx.Field) {
			return nil // Silently skip as per the comment
		}
		return err
	}

	// Remove the key from the index
	keys, ok := idx.Values[fieldValue]
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
		delete(idx.Values, fieldValue)
	} else {
		idx.Values[fieldValue] = newKeys
	}

	return nil
}

// UpdateEntry updates an entry in the index
// If the field is not found in either the old or new data, those operations are silently skipped
func (idx *BasicIndex) UpdateEntry(key string, oldValue, newValue []byte) error {
	// Remove the old entry
	if err := idx.RemoveEntry(key, oldValue); err != nil {
		return err
	}

	// Add the new entry
	return idx.AddEntry(key, newValue)
}

// Query queries the index for keys matching the given value
func (idx *BasicIndex) Query(value string) ([]string, error) {
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

// GetAllValues returns all unique values in the index
func (idx *BasicIndex) GetAllValues() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	values := make([]string, 0, len(idx.Values))
	for value := range idx.Values {
		values = append(values, value)
	}

	return values
}

// Size returns the number of unique values in the index
func (idx *BasicIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.Values)
}

// Clear clears the index
func (idx *BasicIndex) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.Values = make(map[string][]string)
}

// GetName returns the name of the index
func (idx *BasicIndex) GetName() string {
	return idx.Name
}

// GetSetName returns the name of the set this index is for
func (idx *BasicIndex) GetSetName() string {
	return idx.SetName
}

// GetField returns the field this index is on
func (idx *BasicIndex) GetField() string {
	return idx.Field
}

// GetType returns the type of this index
func (idx *BasicIndex) GetType() IndexType {
	return BasicIndexType
}