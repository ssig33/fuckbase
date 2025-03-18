#!/usr/bin/env ruby

require 'bundler/inline'

gemfile do
  source 'https://rubygems.org'
  gem 'minitest'
  gem 'base64'  # Add base64 gem for Ruby 3.4+ compatibility
end

require 'minitest/autorun'
require_relative './fuckbase'

class TestFuckBase < Minitest::Test
  def setup
    # Configure the client to connect to your FuckBase server
    @client = FuckBase::Client.new(
      host: 'localhost',
      port: 8080,
      # Uncomment and set these if your server requires admin authentication
      # admin_username: 'admin',
      # admin_password: 'secure_password'
    )
    
    # Create a test database
    begin
      @client.drop_database('test_db')
    rescue
      # Ignore errors if the database doesn't exist
    end
    
    @db = @client.create_database('test_db')
    @set = @db.create_set('test_set')
  end
  
  def teardown
    # Clean up by dropping the test database
    begin
      @client.drop_database('test_db')
    rescue
      # Ignore errors
    end
  end
  
  def test_server_info
    # Test getting server info
    info = @client.server_info
    assert_equal 'success', info['status']
    assert info['version']
    assert info['uptime']
    assert info['databases_count'] >= 1
  end
  
  def test_database_operations
    # Test listing sets
    sets = @db.list_sets
    assert_includes sets, 'test_set'
    
    # Test creating another set
    another_set = @db.create_set('another_set')
    sets = @db.list_sets
    assert_includes sets, 'another_set'
  end
  
  def test_set_operations
    # Test putting and getting data
    user_data = { 'name' => 'John Doe', 'email' => 'john@example.com', 'age' => 30 }
    response = @set.put('user1', user_data)
    assert_equal 'success', response['status']
    
    # Get the data we just put
    retrieved_data = @set.get('user1')
    assert_equal user_data, retrieved_data
    
    # Test deleting data
    response = @set.delete('user1')
    assert_equal 'success', response['status']
    
    # Verify the data is deleted - this should return nil or raise an error
    begin
      deleted_data = @set.get('user1')
      assert_nil deleted_data, "Expected nil for deleted key, got: #{deleted_data.inspect}"
    rescue RuntimeError => e
      # It's also acceptable if the server returns a KEY_NOT_FOUND error
      assert_match(/KEY_NOT_FOUND/, e.message)
    end
  end
  
  def test_index_operations
    # Add some test data
    @set.put('user1', { 'name' => 'John Doe', 'email' => 'john@example.com', 'age' => 30 })
    @set.put('user2', { 'name' => 'Jane Smith', 'email' => 'jane@example.com', 'age' => 25 })
    @set.put('user3', { 'name' => 'Bob Johnson', 'email' => 'bob@example.com', 'age' => 40 })
    
    # Create an index on the email field
    response = @set.create_index('email_index', 'email')
    assert_equal 'success', response['status']
    
    # Query the index
    results = @set.query_index('email_index', 'jane@example.com')
    assert_equal 1, results[:count]
    assert_equal 'user2', results[:data][0]['key']
    assert_equal 'Jane Smith', results[:data][0]['value']['name']
    
    # Drop the index
    response = @db.drop_index('test_set', 'email_index')
    assert_equal 'success', response['status']
  end
  
  def test_backup_operations
    # Skip this test if S3 is not configured
    begin
      # Create a backup
      response = @db.create_backup
      assert_equal 'success', response['status']
      
      # List backups
      backups = @db.list_backups
      assert backups.length >= 1
    rescue => e
      skip "S3 backup not available: #{e.message}"
    end
  end
end

puts "Running FuckBase client tests..."