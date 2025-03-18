package database

import (
	"fmt"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

// Set represents a collection of key-value pairs
type Set struct {
	Name string
	Data map[string][]byte // Key to MessagePack encoded value
	mu   sync.RWMutex
}

// NewSet creates a new set with the given name
func NewSet(name string) *Set {
	return &Set{
		Name: name,
		Data: make(map[string][]byte),
	}
}

// Put stores a value for a key
func (s *Set) Put(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Encode the value using MessagePack
	encoded, err := msgpack.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}

	s.Data[key] = encoded
	return nil
}

// Get retrieves a value for a key and decodes it into the provided destination
func (s *Set) Get(key string, dest interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	encoded, exists := s.Data[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	// Decode the value using MessagePack
	if err := msgpack.Unmarshal(encoded, dest); err != nil {
		return fmt.Errorf("failed to decode value: %w", err)
	}

	return nil
}

// GetRaw retrieves the raw MessagePack encoded value for a key
func (s *Set) GetRaw(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	encoded, exists := s.Data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return encoded, nil
}

// Delete removes a key-value pair
func (s *Set) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Data[key]; !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	delete(s.Data, key)
	return nil
}

// Has checks if a key exists
func (s *Set) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.Data[key]
	return exists
}

// Keys returns all keys in the set
func (s *Set) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.Data))
	for key := range s.Data {
		keys = append(keys, key)
	}

	return keys
}

// Size returns the number of key-value pairs in the set
func (s *Set) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.Data)
}

// ForEach iterates over all key-value pairs in the set
// The callback function receives the key and the raw MessagePack encoded value
func (s *Set) ForEach(callback func(key string, value []byte) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for key, value := range s.Data {
		if err := callback(key, value); err != nil {
			return err
		}
	}

	return nil
}

// Clear removes all key-value pairs from the set
func (s *Set) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data = make(map[string][]byte)
}