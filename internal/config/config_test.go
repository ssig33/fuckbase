package config

import (
	"os"
	"testing"
)

func TestNewServerConfig(t *testing.T) {
	cfg := NewServerConfig()

	// Check default values
	if cfg.Port != 8080 {
		t.Errorf("Expected default port to be 8080, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("Expected default host to be '0.0.0.0', got '%s'", cfg.Host)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("Expected default data directory to be './data', got '%s'", cfg.DataDir)
	}
	if cfg.AdminAuth == nil || cfg.AdminAuth.Enabled {
		t.Errorf("Expected admin auth to be disabled by default")
	}
	if cfg.S3Config == nil || cfg.S3Config.Enabled {
		t.Errorf("Expected S3 config to be disabled by default")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default log level to be 'info', got '%s'", cfg.LogLevel)
	}
	if cfg.LogFile != "stdout" {
		t.Errorf("Expected default log file to be 'stdout', got '%s'", cfg.LogFile)
	}
	if cfg.BackupInterval != 60 {
		t.Errorf("Expected default backup interval to be 60, got %d", cfg.BackupInterval)
	}
}

func TestParseEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("FUCKBASE_PORT", "9090")
	os.Setenv("FUCKBASE_HOST", "127.0.0.1")
	os.Setenv("FUCKBASE_DATA_DIR", "/tmp/data")
	os.Setenv("FUCKBASE_ADMIN_USERNAME", "admin")
	os.Setenv("FUCKBASE_ADMIN_PASSWORD", "password")
	os.Setenv("FUCKBASE_S3_ENDPOINT", "https://s3.example.com")
	os.Setenv("FUCKBASE_S3_BUCKET", "my-bucket")
	os.Setenv("FUCKBASE_S3_ACCESS_KEY", "access-key")
	os.Setenv("FUCKBASE_S3_SECRET_KEY", "secret-key")
	os.Setenv("FUCKBASE_S3_REGION", "us-west-2")
	os.Setenv("FUCKBASE_BACKUP_INTERVAL", "120")
	os.Setenv("FUCKBASE_LOG_LEVEL", "debug")
	os.Setenv("FUCKBASE_LOG_FILE", "/tmp/fuckbase.log")

	// Create a new config and parse environment variables
	cfg := NewServerConfig()
	cfg.ParseEnv()

	// Check parsed values
	if cfg.Port != 9090 {
		t.Errorf("Expected port to be 9090, got %d", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("Expected host to be '127.0.0.1', got '%s'", cfg.Host)
	}
	if cfg.DataDir != "/tmp/data" {
		t.Errorf("Expected data directory to be '/tmp/data', got '%s'", cfg.DataDir)
	}
	if !cfg.AdminAuth.Enabled || cfg.AdminAuth.Username != "admin" || cfg.AdminAuth.Password != "password" {
		t.Errorf("Expected admin auth to be enabled with username 'admin' and password 'password'")
	}
	if !cfg.S3Config.Enabled || cfg.S3Config.Endpoint != "https://s3.example.com" || cfg.S3Config.Bucket != "my-bucket" || cfg.S3Config.AccessKey != "access-key" || cfg.S3Config.SecretKey != "secret-key" || cfg.S3Config.Region != "us-west-2" {
		t.Errorf("Expected S3 config to be enabled with correct values")
	}
	if cfg.BackupInterval != 120 {
		t.Errorf("Expected backup interval to be 120, got %d", cfg.BackupInterval)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected log level to be 'debug', got '%s'", cfg.LogLevel)
	}
	if cfg.LogFile != "/tmp/fuckbase.log" {
		t.Errorf("Expected log file to be '/tmp/fuckbase.log', got '%s'", cfg.LogFile)
	}

	// Clean up environment variables
	os.Unsetenv("FUCKBASE_PORT")
	os.Unsetenv("FUCKBASE_HOST")
	os.Unsetenv("FUCKBASE_DATA_DIR")
	os.Unsetenv("FUCKBASE_ADMIN_USERNAME")
	os.Unsetenv("FUCKBASE_ADMIN_PASSWORD")
	os.Unsetenv("FUCKBASE_S3_ENDPOINT")
	os.Unsetenv("FUCKBASE_S3_BUCKET")
	os.Unsetenv("FUCKBASE_S3_ACCESS_KEY")
	os.Unsetenv("FUCKBASE_S3_SECRET_KEY")
	os.Unsetenv("FUCKBASE_S3_REGION")
	os.Unsetenv("FUCKBASE_BACKUP_INTERVAL")
	os.Unsetenv("FUCKBASE_LOG_LEVEL")
	os.Unsetenv("FUCKBASE_LOG_FILE")
}

func TestInvalidEnvValues(t *testing.T) {
	// Set invalid environment variables
	os.Setenv("FUCKBASE_PORT", "invalid")
	os.Setenv("FUCKBASE_BACKUP_INTERVAL", "invalid")

	// Create a new config and parse environment variables
	cfg := NewServerConfig()
	cfg.ParseEnv()

	// Check that invalid values are ignored
	if cfg.Port != 8080 {
		t.Errorf("Expected port to remain at default 8080, got %d", cfg.Port)
	}
	if cfg.BackupInterval != 60 {
		t.Errorf("Expected backup interval to remain at default 60, got %d", cfg.BackupInterval)
	}

	// Clean up environment variables
	os.Unsetenv("FUCKBASE_PORT")
	os.Unsetenv("FUCKBASE_BACKUP_INTERVAL")
}

func TestPartialEnvConfig(t *testing.T) {
	// Set partial environment variables
	os.Setenv("FUCKBASE_PORT", "9090")
	os.Setenv("FUCKBASE_ADMIN_USERNAME", "admin")
	// Missing admin password
	os.Setenv("FUCKBASE_S3_ENDPOINT", "https://s3.example.com")
	// Missing other S3 config

	// Create a new config and parse environment variables
	cfg := NewServerConfig()
	cfg.ParseEnv()

	// Check that port is updated
	if cfg.Port != 9090 {
		t.Errorf("Expected port to be 9090, got %d", cfg.Port)
	}

	// Check that admin auth is not enabled due to missing password
	if cfg.AdminAuth.Enabled {
		t.Errorf("Expected admin auth to remain disabled due to missing password")
	}

	// Check that S3 config is not enabled due to missing values
	if cfg.S3Config.Enabled {
		t.Errorf("Expected S3 config to remain disabled due to missing values")
	}

	// Clean up environment variables
	os.Unsetenv("FUCKBASE_PORT")
	os.Unsetenv("FUCKBASE_ADMIN_USERNAME")
	os.Unsetenv("FUCKBASE_S3_ENDPOINT")
}