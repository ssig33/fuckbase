package s3

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ssig33/fuckbase/internal/database"
	"github.com/ssig33/fuckbase/internal/logger"
)

// BackupMetadata represents metadata about a backup
type BackupMetadata struct {
	Timestamp   time.Time `json:"timestamp"`
	Version     string    `json:"version"`
	DatabaseCount int     `json:"database_count"`
	SetCount    int       `json:"set_count"`
	EntryCount  int       `json:"entry_count"`
}

// DatabaseBackup represents a backup of a single database
type DatabaseBackup struct {
	Name    string                     `json:"name"`
	Sets    map[string]SetBackup       `json:"sets"`
	Indexes map[string]IndexBackup     `json:"indexes"`
	Auth    *database.AuthConfig       `json:"auth,omitempty"`
}

// SetBackup represents a backup of a single set
type SetBackup struct {
	Name  string                 `json:"name"`
	Data  map[string]interface{} `json:"data"`
}

// IndexBackup represents a backup of a single index
type IndexBackup struct {
	Name        string   `json:"name"`
	SetName     string   `json:"set_name"`
	Field       string   `json:"field"`
	Type        int      `json:"type"`
	SortFields  []string `json:"sort_fields,omitempty"`
}

// FullBackup represents a full backup of all databases
type FullBackup struct {
	Metadata  BackupMetadata           `json:"metadata"`
	Databases map[string]DatabaseBackup `json:"databases"`
}

// BackupManager handles database backups to S3
type BackupManager struct {
	s3Client *Client
	dbManager *database.Manager
}

// NewBackupManager creates a new backup manager
func NewBackupManager(s3Client *Client, dbManager *database.Manager) *BackupManager {
	return &BackupManager{
		s3Client:  s3Client,
		dbManager: dbManager,
	}
}

// BackupDatabase backs up a single database to S3
func (bm *BackupManager) BackupDatabase(dbName string) error {
	db, err := bm.dbManager.GetDatabase(dbName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	backup, err := bm.createDatabaseBackup(db)
	if err != nil {
		return fmt.Errorf("failed to create database backup: %w", err)
	}

	// Convert backup to JSON
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup data: %w", err)
	}

	// Upload to S3
	objectName := GenerateBackupObjectName(dbName)
	err = bm.s3Client.UploadFile(objectName, data, "application/json")
	if err != nil {
		return fmt.Errorf("failed to upload backup: %w", err)
	}

	logger.Info("Successfully backed up database %s to S3", dbName)
	return nil
}

// BackupAllDatabases backs up all databases to S3
func (bm *BackupManager) BackupAllDatabases() error {
	dbNames := bm.dbManager.ListDatabases()
	
	fullBackup := FullBackup{
		Metadata: BackupMetadata{
			Timestamp:     time.Now().UTC(),
			Version:       "1.0.0",
			DatabaseCount: len(dbNames),
		},
		Databases: make(map[string]DatabaseBackup),
	}

	var totalSets, totalEntries int

	// Backup each database
	for _, dbName := range dbNames {
		db, err := bm.dbManager.GetDatabase(dbName)
		if err != nil {
			logger.Error("Failed to get database %s: %v", dbName, err)
			continue
		}

		backup, err := bm.createDatabaseBackup(db)
		if err != nil {
			logger.Error("Failed to create backup for database %s: %v", dbName, err)
			continue
		}

		fullBackup.Databases[dbName] = backup

		// Count sets and entries
		totalSets += len(backup.Sets)
		for _, set := range backup.Sets {
			totalEntries += len(set.Data)
		}
	}

	fullBackup.Metadata.SetCount = totalSets
	fullBackup.Metadata.EntryCount = totalEntries

	// Convert backup to JSON
	data, err := json.MarshalIndent(fullBackup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup data: %w", err)
	}

	// Upload to S3
	objectName := GenerateFullBackupObjectName()
	err = bm.s3Client.UploadFile(objectName, data, "application/json")
	if err != nil {
		return fmt.Errorf("failed to upload backup: %w", err)
	}

	logger.Info("Successfully backed up all %d databases to S3", len(dbNames))
	return nil
}

// createDatabaseBackup creates a backup of a database
func (bm *BackupManager) createDatabaseBackup(db *database.Database) (DatabaseBackup, error) {
	backup := DatabaseBackup{
		Name:    db.Name,
		Sets:    make(map[string]SetBackup),
		Indexes: make(map[string]IndexBackup),
		Auth:    db.Auth,
	}

	// Backup sets
	for _, setName := range db.ListSets() {
		set, err := db.GetSet(setName)
		if err != nil {
			return DatabaseBackup{}, fmt.Errorf("failed to get set %s: %w", setName, err)
		}

		setBackup := SetBackup{
			Name: set.Name,
			Data: make(map[string]interface{}),
		}

		// Get all keys in the set
		keys := set.Keys()
		for _, key := range keys {
			var value interface{}
			if err := set.Get(key, &value); err != nil {
				logger.Error("Failed to get value for key %s in set %s: %v", key, setName, err)
				continue
			}
			setBackup.Data[key] = value
		}

		backup.Sets[setName] = setBackup
	}

	// Backup indexes
	for _, indexName := range db.ListIndexes() {
		index, err := db.GetIndex(indexName)
		if err != nil {
			return DatabaseBackup{}, fmt.Errorf("failed to get index %s: %w", indexName, err)
		}

		indexBackup := IndexBackup{
			Name:    index.GetName(),
			SetName: index.GetSetName(),
			Field:   index.GetField(),
			Type:    int(index.GetType()),
		}

		// Add sort fields for sortable indexes
		if index.GetType() == database.SortableIndexType {
			if sortableIndex, ok := index.(*database.SortableIndex); ok {
				indexBackup.SortFields = sortableIndex.SortFields
			}
		}

		backup.Indexes[indexName] = indexBackup
	}

	return backup, nil
}

// RestoreDatabase restores a database from S3
func (bm *BackupManager) RestoreDatabase(objectName string) error {
	// Download from S3
	data, err := bm.s3Client.DownloadFile(objectName)
	if err != nil {
		return fmt.Errorf("failed to download backup: %w", err)
	}

	// Parse backup data
	var backup DatabaseBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return fmt.Errorf("failed to unmarshal backup data: %w", err)
	}

	// Check if database already exists
	if bm.dbManager.DatabaseExists(backup.Name) {
		// Delete existing database
		if err := bm.dbManager.DeleteDatabase(backup.Name); err != nil {
			return fmt.Errorf("failed to delete existing database: %w", err)
		}
	}

	// Create database
	db, err := bm.dbManager.CreateDatabase(backup.Name, backup.Auth)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Restore sets
	for _, setBackup := range backup.Sets {
		set, err := db.CreateSet(setBackup.Name)
		if err != nil {
			logger.Error("Failed to create set %s: %v", setBackup.Name, err)
			continue
		}

		// Restore data
		for key, value := range setBackup.Data {
			if err := set.Put(key, value); err != nil {
				logger.Error("Failed to put value for key %s in set %s: %v", key, setBackup.Name, err)
				continue
			}
		}
	}

	// Restore indexes
	for _, indexBackup := range backup.Indexes {
		var err error
		
		// Create the appropriate type of index
		if indexBackup.Type == int(database.BasicIndexType) {
			_, err = db.CreateIndex(indexBackup.Name, indexBackup.SetName, indexBackup.Field)
		} else if indexBackup.Type == int(database.SortableIndexType) {
			_, err = db.CreateSortableIndex(indexBackup.Name, indexBackup.SetName, indexBackup.Field, indexBackup.SortFields)
		} else {
			logger.Error("Unknown index type %d for index %s", indexBackup.Type, indexBackup.Name)
			continue
		}
		
		if err != nil {
			logger.Error("Failed to create index %s: %v", indexBackup.Name, err)
			continue
		}
	}

	logger.Info("Successfully restored database %s from S3", backup.Name)
	return nil
}

// RestoreAllDatabases restores all databases from a full backup
func (bm *BackupManager) RestoreAllDatabases(objectName string) error {
	// Download from S3
	data, err := bm.s3Client.DownloadFile(objectName)
	if err != nil {
		return fmt.Errorf("failed to download backup: %w", err)
	}

	// Parse backup data
	var fullBackup FullBackup
	if err := json.Unmarshal(data, &fullBackup); err != nil {
		return fmt.Errorf("failed to unmarshal backup data: %w", err)
	}

	// Delete all existing databases
	for _, dbName := range bm.dbManager.ListDatabases() {
		if err := bm.dbManager.DeleteDatabase(dbName); err != nil {
			logger.Error("Failed to delete existing database %s: %v", dbName, err)
		}
	}

	// Restore each database
	for dbName, dbBackup := range fullBackup.Databases {
		// Create database
		db, err := bm.dbManager.CreateDatabase(dbName, dbBackup.Auth)
		if err != nil {
			logger.Error("Failed to create database %s: %v", dbName, err)
			continue
		}

		// Restore sets
		for _, setBackup := range dbBackup.Sets {
			set, err := db.CreateSet(setBackup.Name)
			if err != nil {
				logger.Error("Failed to create set %s: %v", setBackup.Name, err)
				continue
			}

			// Restore data
			for key, value := range setBackup.Data {
				if err := set.Put(key, value); err != nil {
					logger.Error("Failed to put value for key %s in set %s: %v", key, setBackup.Name, err)
					continue
				}
			}
		}

		// Restore indexes
		for _, indexBackup := range dbBackup.Indexes {
			var err error
			
			// Create the appropriate type of index
			if indexBackup.Type == int(database.BasicIndexType) {
				_, err = db.CreateIndex(indexBackup.Name, indexBackup.SetName, indexBackup.Field)
			} else if indexBackup.Type == int(database.SortableIndexType) {
				_, err = db.CreateSortableIndex(indexBackup.Name, indexBackup.SetName, indexBackup.Field, indexBackup.SortFields)
			} else {
				logger.Error("Unknown index type %d for index %s", indexBackup.Type, indexBackup.Name)
				continue
			}
			
			if err != nil {
				logger.Error("Failed to create index %s: %v", indexBackup.Name, err)
				continue
			}
		}
	}

	logger.Info("Successfully restored %d databases from S3", len(fullBackup.Databases))
	return nil
}

// ListBackups lists all backups in S3
func (bm *BackupManager) ListBackups() ([]string, error) {
	return bm.s3Client.ListFiles("backups/")
}

// ListDatabaseBackups lists all backups for a specific database
func (bm *BackupManager) ListDatabaseBackups(dbName string) ([]string, error) {
	return bm.s3Client.ListFiles(fmt.Sprintf("backups/%s/", dbName))
}

// ListFullBackups lists all full backups
func (bm *BackupManager) ListFullBackups() ([]string, error) {
	return bm.s3Client.ListFiles("backups/full/")
}