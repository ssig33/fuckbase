#!/bin/bash

set -e

echo "Starting FuckBase server and running client tests..."

# Start FuckBase server using Docker Compose
docker-compose up -d

# Wait for the server to start
echo "Waiting for FuckBase server to start..."
sleep 5

# Run the client tests
echo "Running client tests..."
ruby test_fuckbase.rb

# Stop the server
echo "Tests completed, stopping FuckBase server..."
docker-compose down

echo "Done!"