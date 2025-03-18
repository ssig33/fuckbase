package config

import (
	"flag"
	"os"
	"strconv"
)

// ServerConfig represents the configuration for the FuckBase server
type ServerConfig struct {
	Port           int
	Host           string
	DataDir        string
	AdminAuth      *AdminAuthConfig
	S3Config       *S3Config
	LogLevel       string
	LogFile        string
	BackupInterval int
}

// AdminAuthConfig represents the configuration for admin authentication
type AdminAuthConfig struct {
	Username string
	Password string
	Enabled  bool
}

// S3Config represents the configuration for S3 integration
type S3Config struct {
	Endpoint    string
	Bucket      string
	AccessKey   string
	SecretKey   string
	Region      string
	Enabled     bool
}

// NewServerConfig creates a new server configuration with default values
func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:           8080,
		Host:           "0.0.0.0",
		DataDir:        "./data",
		AdminAuth:      &AdminAuthConfig{Enabled: false},
		S3Config:       &S3Config{Region: "us-east-1", Enabled: false},
		LogLevel:       "info",
		LogFile:        "stdout",
		BackupInterval: 60,
	}
}

// ParseFlags parses command line flags and updates the configuration
func (c *ServerConfig) ParseFlags() {
	flag.IntVar(&c.Port, "port", c.Port, "Server port")
	flag.StringVar(&c.Host, "host", c.Host, "Server host")
	flag.StringVar(&c.DataDir, "data-dir", c.DataDir, "Data directory")
	
	// Admin auth flags
	adminUsername := flag.String("admin-username", "", "Admin username")
	adminPassword := flag.String("admin-password", "", "Admin password")
	
	// S3 flags
	s3Endpoint := flag.String("s3-endpoint", "", "S3 endpoint URL")
	s3Bucket := flag.String("s3-bucket", "", "S3 bucket name")
	s3AccessKey := flag.String("s3-access-key", "", "S3 access key")
	s3SecretKey := flag.String("s3-secret-key", "", "S3 secret key")
	s3Region := flag.String("s3-region", c.S3Config.Region, "S3 region")
	backupInterval := flag.Int("backup-interval", c.BackupInterval, "Backup interval in minutes")
	
	// Log flags
	flag.StringVar(&c.LogLevel, "log-level", c.LogLevel, "Log level (debug, info, warn, error)")
	flag.StringVar(&c.LogFile, "log-file", c.LogFile, "Log file path")
	
	flag.Parse()
	
	// Update admin auth config if provided
	if *adminUsername != "" && *adminPassword != "" {
		c.AdminAuth.Username = *adminUsername
		c.AdminAuth.Password = *adminPassword
		c.AdminAuth.Enabled = true
	}
	
	// Update S3 config if provided
	if *s3Endpoint != "" && *s3Bucket != "" && *s3AccessKey != "" && *s3SecretKey != "" {
		c.S3Config.Endpoint = *s3Endpoint
		c.S3Config.Bucket = *s3Bucket
		c.S3Config.AccessKey = *s3AccessKey
		c.S3Config.SecretKey = *s3SecretKey
		c.S3Config.Region = *s3Region
		c.S3Config.Enabled = true
		c.BackupInterval = *backupInterval
	}
}

// ParseEnv parses environment variables and updates the configuration
func (c *ServerConfig) ParseEnv() {
	// Server config
	if port := os.Getenv("FUCKBASE_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Port = p
		}
	}
	
	if host := os.Getenv("FUCKBASE_HOST"); host != "" {
		c.Host = host
	}
	
	if dataDir := os.Getenv("FUCKBASE_DATA_DIR"); dataDir != "" {
		c.DataDir = dataDir
	}
	
	// Admin auth config
	adminUsername := os.Getenv("FUCKBASE_ADMIN_USERNAME")
	adminPassword := os.Getenv("FUCKBASE_ADMIN_PASSWORD")
	if adminUsername != "" && adminPassword != "" {
		c.AdminAuth.Username = adminUsername
		c.AdminAuth.Password = adminPassword
		c.AdminAuth.Enabled = true
	}
	
	// S3 config
	s3Endpoint := os.Getenv("FUCKBASE_S3_ENDPOINT")
	s3Bucket := os.Getenv("FUCKBASE_S3_BUCKET")
	s3AccessKey := os.Getenv("FUCKBASE_S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("FUCKBASE_S3_SECRET_KEY")
	
	if s3Endpoint != "" && s3Bucket != "" && s3AccessKey != "" && s3SecretKey != "" {
		c.S3Config.Endpoint = s3Endpoint
		c.S3Config.Bucket = s3Bucket
		c.S3Config.AccessKey = s3AccessKey
		c.S3Config.SecretKey = s3SecretKey
		c.S3Config.Enabled = true
		
		if s3Region := os.Getenv("FUCKBASE_S3_REGION"); s3Region != "" {
			c.S3Config.Region = s3Region
		}
	}
	
	if backupInterval := os.Getenv("FUCKBASE_BACKUP_INTERVAL"); backupInterval != "" {
		if bi, err := strconv.Atoi(backupInterval); err == nil {
			c.BackupInterval = bi
		}
	}
	
	// Log config
	if logLevel := os.Getenv("FUCKBASE_LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}
	
	if logFile := os.Getenv("FUCKBASE_LOG_FILE"); logFile != "" {
		c.LogFile = logFile
	}
}

// Parse parses both command line flags and environment variables
// Command line flags take precedence over environment variables
func (c *ServerConfig) Parse() {
	c.ParseEnv()
	c.ParseFlags()
}