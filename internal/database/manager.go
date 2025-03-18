package database

import (
	"fmt"
	"sync"

	"github.com/ssig33/fuckbase/internal/logger"
)

// Manager manages multiple databases
type Manager struct {
	Databases map[string]*Database
	mu        sync.RWMutex
}

// NewManager creates a new database manager
func NewManager() *Manager {
	return &Manager{
		Databases: make(map[string]*Database),
	}
}

// CreateDatabase creates a new database with the given name and optional authentication
func (m *Manager) CreateDatabase(name string, auth *AuthConfig) (*Database, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.Databases[name]; exists {
		return nil, fmt.Errorf("database already exists: %s", name)
	}

	db := NewDatabase(name, auth)
	m.Databases[name] = db
	logger.Info("Created database: %s", name)
	return db, nil
}

// GetDatabase returns a database by name
func (m *Manager) GetDatabase(name string) (*Database, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	db, exists := m.Databases[name]
	if !exists {
		return nil, fmt.Errorf("database not found: %s", name)
	}

	return db, nil
}

// DeleteDatabase deletes a database
func (m *Manager) DeleteDatabase(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.Databases[name]; !exists {
		return fmt.Errorf("database not found: %s", name)
	}

	delete(m.Databases, name)
	logger.Info("Deleted database: %s", name)
	return nil
}

// ListDatabases returns a list of all database names
func (m *Manager) ListDatabases() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	databases := make([]string, 0, len(m.Databases))
	for name := range m.Databases {
		databases = append(databases, name)
	}

	return databases
}

// DatabaseExists checks if a database exists
func (m *Manager) DatabaseExists(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.Databases[name]
	return exists
}

// AuthenticateDatabase authenticates a user against a database
func (m *Manager) AuthenticateDatabase(name, username, password string) (bool, error) {
	db, err := m.GetDatabase(name)
	if err != nil {
		return false, err
	}

	return db.Authenticate(username, password), nil
}

// GetDatabaseCount returns the number of databases
func (m *Manager) GetDatabaseCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.Databases)
}