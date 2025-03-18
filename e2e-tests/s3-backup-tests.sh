#!/bin/sh

# E2E test script for FuckBase S3 backup functionality
# This script tests the backup and restore functionality using MinIO

# Set the base URL
BASE_URL="http://fuckbase:8080"

# Function to check if a command succeeds
check_success() {
  if [ $? -ne 0 ]; then
    echo "❌ Test failed: $1"
    exit 1
  else
    echo "✅ Test passed: $1"
  fi
}

echo "🔍 Starting FuckBase S3 Backup E2E tests..."

# Test 1: Create a test database
echo "\n📋 Test 1: Creating a test database..."
curl -s -X POST $BASE_URL/create -d '{"name": "backup_test_db"}' | grep -q '"status":"success"'
check_success "Create test database"

# Test 2: Create a test set
echo "\n📋 Test 2: Creating a test set..."
curl -s -X POST $BASE_URL/set/create -d '{"database": "backup_test_db", "name": "test_set"}' | grep -q '"status":"success"'
check_success "Create test set"

# Test 3: Put data into the set
echo "\n📋 Test 3: Putting data into the set..."
curl -s -X POST $BASE_URL/set/put -d '{
  "database": "backup_test_db",
  "set": "test_set",
  "key": "test1",
  "value": {
    "name": "Test Data",
    "value": 42,
    "tags": ["test", "backup", "restore"]
  }
}' | grep -q '"status":"success"'
check_success "Put data"

# Test 4: Create a backup of the database
echo "\n📋 Test 4: Creating a backup of the database..."
RESPONSE=$(curl -s -X POST $BASE_URL/backup/create -d '{
  "database": "backup_test_db"
}')
echo "$RESPONSE" | grep -q '"status":"success"'
check_success "Create database backup"

# Test 5: List backups
echo "\n📋 Test 5: Listing backups..."
RESPONSE=$(curl -s -X POST $BASE_URL/backup/list -d '{
  "database": "backup_test_db"
}')
echo "$RESPONSE" | grep -q '"status":"success"'
check_success "List backups"

# Extract the backup name from the response
BACKUP_NAME=$(echo "$RESPONSE" | grep -o '"name":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$BACKUP_NAME" ]; then
  echo "❌ Test failed: Could not extract backup name from response"
  exit 1
fi
echo "Found backup: $BACKUP_NAME"

# Test 6: Delete the database
echo "\n📋 Test 6: Deleting the test database..."
curl -s -X POST $BASE_URL/drop -d '{
  "name": "backup_test_db"
}' | grep -q '"status":"success"'
check_success "Delete test database"

# Test 7: Restore the database from backup
echo "\n📋 Test 7: Restoring the database from backup..."
curl -s -X POST $BASE_URL/backup/restore -d "{
  \"backup_name\": \"$BACKUP_NAME\"
}" | grep -q '"status":"success"'
check_success "Restore database from backup"

# Test 8: Verify the restored data
echo "\n📋 Test 8: Verifying the restored data..."
RESPONSE=$(curl -s -X POST $BASE_URL/set/get -d '{
  "database": "backup_test_db",
  "set": "test_set",
  "key": "test1"
}')
echo "$RESPONSE" | grep -q '"status":"success"'
check_success "Get restored data"
echo "$RESPONSE" | grep -q '"name":"Test Data"'
check_success "Verify restored data content"

# Test 9: Create a full backup
echo "\n📋 Test 9: Creating a full backup..."
curl -s -X POST $BASE_URL/backup/create -d '{}' | grep -q '"status":"success"'
check_success "Create full backup"

# Test 10: List full backups
echo "\n📋 Test 10: Listing full backups..."
RESPONSE=$(curl -s -X POST $BASE_URL/backup/list -d '{}')
echo "$RESPONSE" | grep -q '"status":"success"'
check_success "List full backups"

# Extract the full backup name from the response
FULL_BACKUP_NAME=$(echo "$RESPONSE" | grep -o '"name":"backups/full/[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$FULL_BACKUP_NAME" ]; then
  echo "❌ Test failed: Could not extract full backup name from response"
  exit 1
fi
echo "Found full backup: $FULL_BACKUP_NAME"

# Test 11: Create another database with different data
echo "\n📋 Test 11: Creating another test database..."
curl -s -X POST $BASE_URL/create -d '{"name": "backup_test_db2"}' | grep -q '"status":"success"'
check_success "Create second test database"

curl -s -X POST $BASE_URL/set/create -d '{"database": "backup_test_db2", "name": "test_set2"}' | grep -q '"status":"success"'
check_success "Create test set in second database"

curl -s -X POST $BASE_URL/set/put -d '{
  "database": "backup_test_db2",
  "set": "test_set2",
  "key": "test2",
  "value": {
    "name": "Different Data",
    "value": 100,
    "tags": ["different", "test"]
  }
}' | grep -q '"status":"success"'
check_success "Put data in second database"

# Test 12: Restore from full backup
echo "\n📋 Test 12: Restoring from full backup..."
curl -s -X POST $BASE_URL/backup/restore -d "{
  \"backup_name\": \"$FULL_BACKUP_NAME\"
}" | grep -q '"status":"success"'
check_success "Restore from full backup"

# Test 13: Verify the first database is restored correctly
echo "\n📋 Test 13: Verifying first database is restored correctly..."
RESPONSE=$(curl -s -X POST $BASE_URL/set/get -d '{
  "database": "backup_test_db",
  "set": "test_set",
  "key": "test1"
}')
echo "$RESPONSE" | grep -q '"status":"success"'
check_success "Get data from first restored database"
echo "$RESPONSE" | grep -q '"name":"Test Data"'
check_success "Verify first database content"

# Test 14: Verify the second database is gone (replaced by the backup)
echo "\n📋 Test 14: Verifying second database is gone..."
RESPONSE=$(curl -s -X POST $BASE_URL/set/get -d '{
  "database": "backup_test_db2",
  "set": "test_set2",
  "key": "test2"
}')
echo "$RESPONSE" | grep -q '"status":"error"'
check_success "Verify second database is gone"

# Test 15: Clean up - drop the test database
echo "\n📋 Test 15: Cleaning up - dropping the test database..."
curl -s -X POST $BASE_URL/drop -d '{
  "name": "backup_test_db"
}' | grep -q '"status":"success"'
check_success "Drop test database"

echo "\n🎉 All S3 backup tests passed successfully!"