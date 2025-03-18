# FuckBase Ruby Client

A simple Ruby client library for interacting with FuckBase, a HTTP-based key-value database with indexing capabilities.

## Features

- Simple, single-file implementation
- Support for all FuckBase API operations
- Object-oriented interface for databases and sets
- Authentication support
- Index operations
- Backup and restore operations

## Installation

Simply copy the `fuckbase.rb` file to your project and require it:

```ruby
require_relative 'path/to/fuckbase'
```

## Dependencies

- Ruby 2.6 or higher
- `msgpack` gem for MessagePack encoding/decoding

## Usage

### Basic Usage

```ruby
# Create a client
client = FuckBase::Client.new(
  host: 'localhost',
  port: 8080,
  # Optional admin credentials
  admin_username: 'admin',
  admin_password: 'secure_password'
)

# Create a database
db = client.create_database('my_database')

# Create a set
users_set = db.create_set('users')

# Store data
users_set.put('user1', {
  'name' => 'John Doe',
  'email' => 'john@example.com',
  'age' => 30
})

# Retrieve data
user = users_set.get('user1')
puts "User: #{user['name']}, Email: #{user['email']}"

# Create an index
users_set.create_index('email_index', 'email')

# Query the index
results = users_set.query_index('email_index', 'john@example.com')
puts "Found #{results[:count]} users"

# Delete data
users_set.delete('user1')

# Drop the database when done
client.drop_database('my_database')
```

### Authentication

```ruby
# Create a database with authentication
db = client.create_database('secure_db', auth: {
  username: 'db_user',
  password: 'db_password'
})

# Connect to an existing database with authentication
db = client.database('secure_db', auth: {
  username: 'db_user',
  password: 'db_password'
})
```

### Backup and Restore

```ruby
# Create a backup
db.create_backup

# List backups
backups = db.list_backups
puts "Available backups: #{backups.map { |b| b['name'] }.join(', ')}"

# Restore from a backup
client.restore_backup('backups/my_database/20250318-123456.json')
```

## Testing

Run the tests with:

```bash
./run-client-tests.sh
```

This script will:
1. Start a FuckBase server using Docker Compose
2. Run the tests
3. Stop the server when done

## CI Integration

This repository includes a GitHub Actions workflow that automatically runs the tests on push and pull requests.

## License

This client library is released under the same license as FuckBase (WTFPL).