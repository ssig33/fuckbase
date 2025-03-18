#!/bin/bash

# Start the containers if they're not already running
echo "Starting containers..."
docker compose up -d

# Wait for the server to be ready
echo "Waiting for server to be ready..."
sleep 5

# Run the tests
echo "Running E2E tests..."
docker compose exec test-client sh /tests/run-tests.sh

# Get the exit code
EXIT_CODE=$?

# Output the result
if [ $EXIT_CODE -eq 0 ]; then
  echo "✅ E2E tests passed!"
else
  echo "❌ E2E tests failed!"
fi

# Return the exit code
exit $EXIT_CODE