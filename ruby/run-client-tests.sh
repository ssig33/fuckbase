#!/bin/bash

set -e

echo "Starting FuckBase server and running client tests..."

# Rebuild and start FuckBase server using Docker Compose
docker-compose up -d --build

# Wait for the server to start
echo "Waiting for FuckBase server to start..."
sleep 5

# Run the client tests
echo "Running client tests..."
echo "Running basic client tests..."
ruby $(dirname "$0")/test_fuckbase.rb

echo "Running sortable index tests..."
ruby $(dirname "$0")/test_sortable_index.rb

# Stop the server
echo "Tests completed, stopping FuckBase server..."
docker-compose down

echo "Done!"