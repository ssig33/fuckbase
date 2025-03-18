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

// Database represents a FuckBase database
type Database struct {
	Name    string
	Sets    map[string]*Set
	Indexes map[string]*Index
	Auth    *AuthConfig
	mu      sync.RWMutex
}

// NewDatabase creates a new database with the given name and optional authentication
func NewDatabase(name string, auth *AuthConfig) *Database {
	return &Database{
		Name:    name,
		Sets:    make(map[string]*Set),
		Indexes: make(map[string]*Index),
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

// CreateIndex creates a new index for a set
func (db *Database) CreateIndex(name string, setName string, field string) (*Index, error) {
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

// GetIndex returns an index by name
func (db *Database) GetIndex(name string) (*Index, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	index, exists := db.Indexes[name]
	if !exists {
		return nil, fmt.Errorf("index not found: %s", name)
	}

	return index, nil
}

// DeleteIndex deletes an index from the database
func (db *Database) DeleteIndex(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.Indexes[name]; !exists {
		return fmt.Errorf("index not found: %s", name)
	}

	delete(db.Indexes, name)
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