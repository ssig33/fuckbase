#!/bin/sh

# E2E test script for FuckBase
# This script tests the basic functionality of FuckBase using curl commands

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

echo "🔍 Starting FuckBase E2E tests..."

# Test 1: Create a database
echo "\n📋 Test 1: Creating a database..."
curl -s -X POST $BASE_URL/create -d '{"name": "testdb"}' | grep -q '"status":"success"'
check_success "Create database"

# Test 2: Create a set
echo "\n📋 Test 2: Creating a set..."
curl -s -X POST $BASE_URL/set/create -d '{"database": "testdb", "name": "users"}' | grep -q '"status":"success"'
check_success "Create set"

# Test 3: Put data into the set
echo "\n📋 Test 3: Putting data into the set..."
curl -s -X POST $BASE_URL/set/put -d '{
  "database": "testdb",
  "set": "users",
  "key": "user1",
  "value": {
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }
}' | grep -q '"status":"success"'
check_success "Put data"

# Test 4: Get data from the set
echo "\n📋 Test 4: Getting data from the set..."
RESPONSE=$(curl -s -X POST $BASE_URL/set/get -d '{
  "database": "testdb",
  "set": "users",
  "key": "user1"
}')
echo "$RESPONSE" | grep -q '"status":"success"'
check_success "Get data"
echo "$RESPONSE" | grep -q '"name":"John Doe"'
check_success "Data content verification"

# Test 5: Create an index
echo "\n📋 Test 5: Creating an index..."
curl -s -X POST $BASE_URL/index/create -d '{
  "database": "testdb",
  "set": "users",
  "name": "email_index",
  "field": "email"
}' | grep -q '"status":"success"'
check_success "Create index"

# Test 6: Query the index
echo "\n📋 Test 6: Querying the index..."
RESPONSE=$(curl -s -X POST $BASE_URL/index/query -d '{
  "database": "testdb",
  "set": "users",
  "index": "email_index",
  "value": "john@example.com"
}')
echo "$RESPONSE" | grep -q '"status":"success"'
check_success "Query index"
echo "$RESPONSE" | grep -q '"count":1'
check_success "Index query result count"

# Test 7: Delete data
echo "\n📋 Test 7: Deleting data..."
curl -s -X POST $BASE_URL/set/delete -d '{
  "database": "testdb",
  "set": "users",
  "key": "user1"
}' | grep -q '"status":"success"'
check_success "Delete data"

# Test 8: Verify data is deleted
echo "\n📋 Test 8: Verifying data is deleted..."
curl -s -X POST $BASE_URL/set/get -d '{
  "database": "testdb",
  "set": "users",
  "key": "user1"
}' | grep -q '"status":"error"'
check_success "Data deletion verification"

# Test 9: Drop the index
echo "\n📋 Test 9: Dropping the index..."
curl -s -X POST $BASE_URL/index/drop -d '{
  "database": "testdb",
  "set": "users",
  "name": "email_index"
}' | grep -q '"status":"success"'
check_success "Drop index"

# Test 10: Drop the database
echo "\n📋 Test 10: Dropping the database..."
curl -s -X POST $BASE_URL/drop -d '{
  "name": "testdb"
}' | grep -q '"status":"success"'
check_success "Drop database"

echo "\n🎉 All tests passed successfully!"