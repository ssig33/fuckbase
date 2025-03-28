package server

import (
	"encoding/json"
	"time"
)

// Response is the base response structure
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse is the error response structure
type ErrorResponse struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CreateDatabaseRequest is the request structure for creating a database
type CreateDatabaseRequest struct {
	Name string `json:"name"`
	Auth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
	AdminAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"admin_auth"`
}

// DropDatabaseRequest is the request structure for dropping a database
type DropDatabaseRequest struct {
	Name      string `json:"name"`
	AdminAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"admin_auth"`
}

// CreateSetRequest is the request structure for creating a set
type CreateSetRequest struct {
	Database string `json:"database"`
	Name     string `json:"name"`
	Auth     struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// GetSetRequest is the request structure for getting a value from a set
type GetSetRequest struct {
	Database string `json:"database"`
	Set      string `json:"set"`
	Key      string `json:"key"`
	Auth     struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// PutSetRequest is the request structure for putting a value into a set
type PutSetRequest struct {
	Database string          `json:"database"`
	Set      string          `json:"set"`
	Key      string          `json:"key"`
	Value    json.RawMessage `json:"value"`
	Auth     struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// DeleteSetRequest is the request structure for deleting a value from a set
type DeleteSetRequest struct {
	Database string `json:"database"`
	Set      string `json:"set"`
	Key      string `json:"key"`
	Auth     struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// ListSetsRequest is the request structure for listing sets in a database
type ListSetsRequest struct {
	Database string `json:"database"`
	Auth     struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// CreateIndexRequest is the request structure for creating an index
type CreateIndexRequest struct {
	Database string `json:"database"`
	Set      string `json:"set"`
	Name     string `json:"name"`
	Field    string `json:"field"`
	Auth     struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// DropIndexRequest is the request structure for dropping an index
type DropIndexRequest struct {
	Database string `json:"database"`
	Set      string `json:"set"`
	Name     string `json:"name"`
	Auth     struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// QueryIndexRequest is the request structure for querying an index
type QueryIndexRequest struct {
	Database string `json:"database"`
	Set      string `json:"set"`
	Index    string `json:"index"`
	Value    string `json:"value"`
	Auth     struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// ServerInfoRequest is the request structure for getting server info
type ServerInfoRequest struct {
	AdminAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"admin_auth"`
}

// ServerInfoResponse is the response structure for server info
type ServerInfoResponse struct {
	Status         string `json:"status"`
	Version        string `json:"version"`
	Uptime         string `json:"uptime"`
	DatabasesCount int    `json:"databases_count"`
}

// CreateBackupRequest is the request structure for creating a backup
type CreateBackupRequest struct {
	Database  string `json:"database,omitempty"` // If empty, backup all databases
	AdminAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"admin_auth"`
}

// ListBackupsRequest is the request structure for listing backups
type ListBackupsRequest struct {
	Database  string `json:"database,omitempty"` // If empty, list all backups
	AdminAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"admin_auth"`
}

// BackupInfo represents information about a backup
type BackupInfo struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size"`
	Database  string    `json:"database,omitempty"`
}

// ListBackupsResponse is the response structure for listing backups
type ListBackupsResponse struct {
	Status  string       `json:"status"`
	Backups []BackupInfo `json:"backups"`
}

// RestoreBackupRequest is the request structure for restoring a backup
type RestoreBackupRequest struct {
	BackupName string `json:"backup_name"`
	AdminAuth  struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"admin_auth"`
}

// CreateSortableIndexRequest is the request structure for creating a sortable index
type CreateSortableIndexRequest struct {
	Database     string   `json:"database"`
	Set          string   `json:"set"`
	Name         string   `json:"name"`
	PrimaryField string   `json:"primary_field"`
	SortFields   []string `json:"sort_fields"`
	Auth         struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// SortField represents a sort field and its order
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}

// Pagination represents pagination parameters
type Pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// QuerySortedIndexRequest is the request structure for querying a sortable index with single field sorting
type QuerySortedIndexRequest struct {
	Database   string     `json:"database"`
	Set        string     `json:"set"`
	Index      string     `json:"index"`
	Value      string     `json:"value"`
	Sort       SortField  `json:"sort"`
	Pagination Pagination `json:"pagination,omitempty"`
	Auth       struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// QueryMultiSortedIndexRequest is the request structure for querying a sortable index with multi-field sorting
type QueryMultiSortedIndexRequest struct {
	Database   string      `json:"database"`
	Set        string      `json:"set"`
	Index      string      `json:"index"`
	Value      string      `json:"value"`
	Sort       []SortField `json:"sort"`
	Pagination Pagination  `json:"pagination,omitempty"`
	Auth       struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}