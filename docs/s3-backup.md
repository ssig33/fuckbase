# S3 Backup and Restore Functionality

FuckBase provides a robust backup and restore system using S3-compatible storage. This document explains how to configure and use the S3 backup functionality.

## Configuration

The S3 backup functionality can be configured using environment variables or command-line arguments:

### Environment Variables

```
FUCKBASE_S3_ENDPOINT=minio:9000
FUCKBASE_S3_BUCKET=fuckbase-backups
FUCKBASE_S3_ACCESS_KEY=minioadmin
FUCKBASE_S3_SECRET_KEY=minioadmin
FUCKBASE_S3_REGION=us-east-1
FUCKBASE_BACKUP_INTERVAL=60
```

### Command-Line Arguments

```
--s3-endpoint minio:9000 --s3-bucket fuckbase-backups --s3-access-key minioadmin --s3-secret-key minioadmin --s3-region us-east-1 --backup-interval 60
```

## Backup Types

FuckBase supports two types of backups:

1. **Database-specific backups**: Backs up a single database.
2. **Full backups**: Backs up all databases.

## API Endpoints

### Create a Backup

To create a backup of a specific database:

```
POST /backup/create
{
  "database": "your_database_name"
}
```

To create a full backup of all databases:

```
POST /backup/create
{}
```

### List Backups

To list backups for a specific database:

```
POST /backup/list
{
  "database": "your_database_name"
}
```

To list all backups:

```
POST /backup/list
{}
```

### Restore from Backup

To restore from a backup:
```
POST /backup/restore
{
  "backup_name": "backups/your_database_name/20250318-140947.json"
}
```
```

## Backup Storage Structure

Backups are stored in the S3 bucket with the following structure:

- Database-specific backups: `backups/{database_name}/{timestamp}.json`
- Full backups: `backups/full/{timestamp}.json`

The timestamp format is `YYYYMMDD-HHMMSS`.

## Automatic Backups

FuckBase can perform automatic backups at regular intervals. The interval is specified in minutes using the `FUCKBASE_BACKUP_INTERVAL` environment variable or the `--backup-interval` command-line argument.

## Testing with MinIO

For development and testing, you can use MinIO as an S3-compatible storage service. The docker-compose.yml file includes a MinIO service configured for testing.

## Implementation Details

The S3 backup functionality is implemented in the following files:

- `internal/s3/client.go`: S3 client implementation
- `internal/s3/backup.go`: Backup and restore functionality
- `internal/server/backup_handlers.go`: HTTP handlers for backup and restore endpoints

## Error Handling

The S3 backup functionality includes robust error handling to ensure data integrity. If an error occurs during backup or restore, the operation is aborted and an error message is returned.