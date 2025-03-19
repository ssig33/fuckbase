package database

import (
	"fmt"
	"sync"
)

// AuthConfig represents the authentication configuration for a database
type AuthConfig struct {
	Username string
	Password string
	Enabled  bool
}

// IndexType represents the type of an index
type IndexType int

const (
	BasicIndexType IndexType = iota
	SortableIndexType
)

// Index is an interface that all index types must implement
type Index interface {
	Build(set *Set) error
	AddEntry(key string, value []byte) error
	RemoveEntry(key string, value []byte) error
	UpdateEntry(key string, oldValue, newValue []byte) error
	Query(value string) ([]string, error)
	GetAllValues() []string
	Size() int
	Clear()
	GetName() string
	GetSetName() string
	GetField() string
	GetType() IndexType
}

// Database represents a FuckBase database
type Database struct {
	Name    string
	Sets    map[string]*Set
	Indexes map[string]Index
	Auth    *AuthConfig
	mu      sync.RWMutex
}

// NewDatabase creates a new database with the given name and optional authentication
func NewDatabase(name string, auth *AuthConfig) *Database {
	return &Database{
		Name:    name,
		Sets:    make(map[string]*Set),
		Indexes: make(map[string]Index),
		Auth:    auth,
	}
}

// CreateSet creates a new set in the database
func (db *Database) CreateSet(name string) (*Set, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.Sets[name]; exists {
		return nil, fmt.Errorf("set already exists: %s", name)
	}

	set := NewSet(name)
	db.Sets[name] = set
	return set, nil
}

// GetSet returns a set by name
func (db *Database) GetSet(name string) (*Set, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	set, exists := db.Sets[name]
	if !exists {
		return nil, fmt.Errorf("set not found: %s", name)
	}

	return set, nil
}

// DeleteSet deletes a set from the database
func (db *Database) DeleteSet(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.Sets[name]; !exists {
		return fmt.Errorf("set not found: %s", name)
	}

	delete(db.Sets, name)
	return nil
}

// ListSets returns a list of all set names in the database
func (db *Database) ListSets() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	sets := make([]string, 0, len(db.Sets))
	for name := range db.Sets {
		sets = append(sets, name)
	}

	return sets
}

// CreateIndex creates a new basic index for a set
func (db *Database) CreateIndex(name string, setName string, field string) (*BasicIndex, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.Indexes[name]; exists {
		return nil, fmt.Errorf("index already exists: %s", name)
	}

	set, exists := db.Sets[setName]
	if !exists {
		return nil, fmt.Errorf("set not found: %s", setName)
	}

	index := NewIndex(name, setName, field)
	
	// Build the index by scanning all entries in the set
	if err := index.Build(set); err != nil {
		return nil, fmt.Errorf("failed to build index: %w", err)
	}

	db.Indexes[name] = index
	return index, nil
}

// CreateSortableIndex creates a new sortable index for a set
func (db *Database) CreateSortableIndex(name string, setName string, primaryField string, sortFields []string) (*SortableIndex, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.Indexes[name]; exists {
		return nil, fmt.Errorf("index already exists: %s", name)
	}

	set, exists := db.Sets[setName]
	if !exists {
		return nil, fmt.Errorf("set not found: %s", setName)
	}

	index := NewSortableIndex(name, setName, primaryField, sortFields)
	
	// Build the index by scanning all entries in the set
	if err := index.Build(set); err != nil {
		return nil, fmt.Errorf("failed to build sortable index: %w", err)
	}

	db.Indexes[name] = index
	return index, nil
}

// GetIndex returns an index by name
func (db *Database) GetIndex(name string) (Index, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	index, exists := db.Indexes[name]
	if !exists {
		return nil, fmt.Errorf("index not found: %s", name)
	}

	return index, nil
}

// DropIndex deletes an index from the database
func (db *Database) DropIndex(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.Indexes[name]; !exists {
		return fmt.Errorf("index not found: %s", name)
	}

	delete(db.Indexes, name)
	return nil
}

// RebuildIndex rebuilds an existing index
func (db *Database) RebuildIndex(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	index, exists := db.Indexes[name]
	if !exists {
		return fmt.Errorf("index not found: %s", name)
	}

	// Get the set for this index
	var setName string
	switch idx := index.(type) {
	case *BasicIndex:
		setName = idx.SetName
	case *SortableIndex:
		setName = idx.SetName
	default:
		return fmt.Errorf("unknown index type for index: %s", name)
	}

	set, exists := db.Sets[setName]
	if !exists {
		return fmt.Errorf("set not found for index: %s", setName)
	}

	// Rebuild the index
	if err := index.Build(set); err != nil {
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	return nil
}

// ListIndexes returns a list of all index names in the database
func (db *Database) ListIndexes() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	indexes := make([]string, 0, len(db.Indexes))
	for name := range db.Indexes {
		indexes = append(indexes, name)
	}

	return indexes
}

// Authenticate authenticates a user against the database's authentication configuration
func (db *Database) Authenticate(username, password string) bool {
	if db.Auth == nil || !db.Auth.Enabled {
		return true
	}

	return db.Auth.Username == username && db.Auth.Password == password
}

// Put adds or updates a value in a set and updates all related indexes
func (db *Database) Put(setName string, key string, value interface{}) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Get the set
	set, exists := db.Sets[setName]
	if !exists {
		return fmt.Errorf("set not found: %s", setName)
	}

	// Get the old value if it exists
	var oldValue []byte
	if set.Has(key) {
		var err error
		oldValue, err = set.GetRaw(key)
		if err != nil {
			return fmt.Errorf("failed to get old value: %w", err)
		}
	}

	// Add the new value to the set
	if err := set.Put(key, value); err != nil {
		return fmt.Errorf("failed to put value: %w", err)
	}

	// Get the new value
	newValue, err := set.GetRaw(key)
	if err != nil {
		return fmt.Errorf("failed to get new value: %w", err)
	}

	// Update all indexes that reference this set
	for _, index := range db.Indexes {
		var indexSetName string
		switch idx := index.(type) {
		case *BasicIndex:
			indexSetName = idx.SetName
		case *SortableIndex:
			indexSetName = idx.SetName
		default:
			continue // Skip unknown index types
		}

		// Only update indexes for this set
		if indexSetName == setName {
			if oldValue == nil {
				// This is a new entry
				if err := index.AddEntry(key, newValue); err != nil {
					return fmt.Errorf("failed to add entry to index: %w", err)
				}
			} else {
				// This is an update
				if err := index.UpdateEntry(key, oldValue, newValue); err != nil {
					return fmt.Errorf("failed to update entry in index: %w", err)
				}
			}
		}
	}

	return nil
}

// Delete removes a value from a set and updates all related indexes
func (db *Database) Delete(setName string, key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Get the set
	set, exists := db.Sets[setName]
	if !exists {
		return fmt.Errorf("set not found: %s", setName)
	}

	// Get the value before deleting
	oldValue, err := set.GetRaw(key)
	if err != nil {
		return fmt.Errorf("failed to get value: %w", err)
	}

	// Delete the value from the set
	if err := set.Delete(key); err != nil {
		return fmt.Errorf("failed to delete value: %w", err)
	}

	// Update all indexes that reference this set
	for _, index := range db.Indexes {
		var indexSetName string
		switch idx := index.(type) {
		case *BasicIndex:
			indexSetName = idx.SetName
		case *SortableIndex:
			indexSetName = idx.SetName
		default:
			continue // Skip unknown index types
		}

		// Only update indexes for this set
		if indexSetName == setName {
			if err := index.RemoveEntry(key, oldValue); err != nil {
				return fmt.Errorf("failed to remove entry from index: %w", err)
			}
		}
	}

	return nil
}