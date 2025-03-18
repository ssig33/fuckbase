require 'net/http'
require 'uri'
require 'json'
require 'base64'

class FuckBase
  class Client
    attr_reader :host, :port, :admin_username, :admin_password

    # Initialize a new FuckBase client
    #
    # @param host [String] The host of the FuckBase server
    # @param port [Integer] The port of the FuckBase server
    # @param admin_username [String] The admin username (optional)
    # @param admin_password [String] The admin password (optional)
    def initialize(host: 'localhost', port: 8080, admin_username: nil, admin_password: nil)
      @host = host
      @port = port
      @admin_username = admin_username
      @admin_password = admin_password
      @base_url = "http://#{host}:#{port}"
    end

    # Create a new database
    #
    # @param name [String] The name of the database
    # @param auth [Hash] Authentication configuration for the database (optional)
    # @return [Database] The created database
    def create_database(name, auth: nil)
      payload = { name: name }
      payload[:auth] = auth if auth

      response = post('/create', payload, admin_auth: true)
      if response['status'] == 'success'
        Database.new(self, name, auth: auth)
      else
        raise "Failed to create database: #{response['message']}"
      end
    end

    # Get a database
    #
    # @param name [String] The name of the database
    # @param auth [Hash] Authentication for the database (optional)
    # @return [Database] The database
    def database(name, auth: nil)
      Database.new(self, name, auth: auth)
    end

    # Drop a database
    #
    # @param name [String] The name of the database
    # @return [Hash] The response from the server
    def drop_database(name)
      post('/drop', { name: name }, admin_auth: true)
    end

    # Create a new set in a database
    #
    # @param database [String] The name of the database
    # @param name [String] The name of the set
    # @param auth [Hash] Authentication for the database (optional)
    # @return [Hash] The response from the server
    def create_set(database, name, auth: nil)
      payload = { database: database, name: name }
      payload[:auth] = auth if auth

      post('/set/create', payload)
    end

    # Get a value from a set
    #
    # @param database [String] The name of the database
    # @param set [String] The name of the set
    # @param key [String] The key to get
    # @param auth [Hash] Authentication for the database (optional)
    # @return [Object] The value
    def get(database, set, key, auth: nil)
      payload = { database: database, set: set, key: key }
      payload[:auth] = auth if auth

      response = post('/set/get', payload)
      response['data'] if response['status'] == 'success'
    end

    # Put a value into a set
    #
    # @param database [String] The name of the database
    # @param set [String] The name of the set
    # @param key [String] The key to put
    # @param value [Object] The value to put
    # @param auth [Hash] Authentication for the database (optional)
    # @return [Hash] The response from the server
    def put(database, set, key, value, auth: nil)
      payload = { 
        database: database, 
        set: set, 
        key: key, 
        value: value 
      }
      payload[:auth] = auth if auth

      post('/set/put', payload)
    end

    # Delete a value from a set
    #
    # @param database [String] The name of the database
    # @param set [String] The name of the set
    # @param key [String] The key to delete
    # @param auth [Hash] Authentication for the database (optional)
    # @return [Hash] The response from the server
    def delete(database, set, key, auth: nil)
      payload = { database: database, set: set, key: key }
      payload[:auth] = auth if auth

      post('/set/delete', payload)
    end

    # List all sets in a database
    #
    # @param database [String] The name of the database
    # @param auth [Hash] Authentication for the database (optional)
    # @return [Array<String>] The list of sets
    def list_sets(database, auth: nil)
      payload = { database: database }
      payload[:auth] = auth if auth

      response = post('/set/list', payload)
      response.dig('data', 'sets') if response['status'] == 'success'
    end

    # Create an index on a field in a set
    #
    # @param database [String] The name of the database
    # @param set [String] The name of the set
    # @param name [String] The name of the index
    # @param field [String] The field to index
    # @param auth [Hash] Authentication for the database (optional)
    # @return [Hash] The response from the server
    def create_index(database, set, name, field, auth: nil)
      payload = { database: database, set: set, name: name, field: field }
      payload[:auth] = auth if auth

      post('/index/create', payload)
    end

    # Drop an index
    #
    # @param database [String] The name of the database
    # @param set [String] The name of the set
    # @param name [String] The name of the index
    # @param auth [Hash] Authentication for the database (optional)
    # @return [Hash] The response from the server
    def drop_index(database, set, name, auth: nil)
      payload = { database: database, set: set, name: name }
      payload[:auth] = auth if auth

      post('/index/drop', payload)
    end

    # Query an index
    #
    # @param database [String] The name of the database
    # @param set [String] The name of the set
    # @param index [String] The name of the index
    # @param value [String] The value to query
    # @param sort [String] The sort order ('asc' or 'desc')
    # @param auth [Hash] Authentication for the database (optional)
    # @return [Hash] The response from the server
    def query_index(database, set, index, value, sort: 'asc', auth: nil)
      payload = {
        database: database,
        set: set,
        index: index,
        value: value,
        sort: sort
      }
      payload[:auth] = auth if auth

      response = post('/index/query', payload)
      if response['status'] == 'success'
        {
          count: response.dig('data', 'count'),
          data: response.dig('data', 'data')
        }
      end
    end

    # Get server information
    #
    # @return [Hash] The server information
    def server_info
      post('/server/info', {}, admin_auth: true)
    end

    # Create a backup of a database
    #
    # @param database [String] The name of the database
    # @return [Hash] The response from the server
    def create_backup(database)
      post('/backup/create', { database: database }, admin_auth: true)
    end

    # List all backups
    #
    # @param database [String] The name of the database (optional)
    # @return [Array<Hash>] The list of backups
    def list_backups(database: nil)
      payload = {}
      payload[:database] = database if database

      response = post('/backup/list', payload, admin_auth: true)
      response['backups'] if response['status'] == 'success'
    end

    # Restore a database from a backup
    #
    # @param backup_name [String] The name of the backup
    # @return [Hash] The response from the server
    def restore_backup(backup_name)
      post('/backup/restore', { backup_name: backup_name }, admin_auth: true)
    end

    private

    # Make a POST request to the server
    #
    # @param path [String] The path to request
    # @param payload [Hash] The payload to send
    # @param admin_auth [Boolean] Whether to use admin authentication
    # @return [Hash] The response from the server
    def post(path, payload, admin_auth: false)
      uri = URI.parse("#{@base_url}#{path}")
      http = Net::HTTP.new(uri.host, uri.port)
      request = Net::HTTP::Post.new(uri.request_uri)
      request.content_type = 'application/json'

      # Add admin authentication if required
      if admin_auth && @admin_username && @admin_password
        auth_string = Base64.strict_encode64("#{@admin_username}:#{@admin_password}")
        request['X-Admin-Authorization'] = "Basic #{auth_string}"
      end

      # Add database authentication if provided in payload
      if payload[:auth] && payload[:auth][:username] && payload[:auth][:password]
        auth_string = Base64.strict_encode64("#{payload[:auth][:username]}:#{payload[:auth][:password]}")
        request['Authorization'] = "Basic #{auth_string}"
      end

      request.body = payload.to_json
      response = http.request(request)

      if response.code.to_i == 200
        JSON.parse(response.body)
      else
        raise "HTTP Error: #{response.code} - #{response.body}"
      end
    end
  end

  class Database
    attr_reader :name, :client, :auth

    # Initialize a new Database instance
    #
    # @param client [FuckBase::Client] The FuckBase client
    # @param name [String] The name of the database
    # @param auth [Hash] Authentication for the database (optional)
    def initialize(client, name, auth: nil)
      @client = client
      @name = name
      @auth = auth
    end

    # Create a new set in the database
    #
    # @param name [String] The name of the set
    # @return [Set] The created set
    def create_set(name)
      response = @client.create_set(@name, name, auth: @auth)
      if response['status'] == 'success'
        Set.new(self, name)
      else
        raise "Failed to create set: #{response['message']}"
      end
    end

    # Get a set from the database
    #
    # @param name [String] The name of the set
    # @return [Set] The set
    def set(name)
      Set.new(self, name)
    end

    # List all sets in the database
    #
    # @return [Array<String>] The list of sets
    def list_sets
      @client.list_sets(@name, auth: @auth)
    end

    # Create an index on a field in a set
    #
    # @param set_name [String] The name of the set
    # @param index_name [String] The name of the index
    # @param field [String] The field to index
    # @return [Hash] The response from the server
    def create_index(set_name, index_name, field)
      @client.create_index(@name, set_name, index_name, field, auth: @auth)
    end

    # Drop an index
    #
    # @param set_name [String] The name of the set
    # @param index_name [String] The name of the index
    # @return [Hash] The response from the server
    def drop_index(set_name, index_name)
      @client.drop_index(@name, set_name, index_name, auth: @auth)
    end

    # Query an index
    #
    # @param set_name [String] The name of the set
    # @param index_name [String] The name of the index
    # @param value [String] The value to query
    # @param sort [String] The sort order ('asc' or 'desc')
    # @return [Hash] The query results
    def query_index(set_name, index_name, value, sort: 'asc')
      @client.query_index(@name, set_name, index_name, value, sort: sort, auth: @auth)
    end

    # Create a backup of the database
    #
    # @return [Hash] The response from the server
    def create_backup
      @client.create_backup(@name)
    end

    # List all backups for the database
    #
    # @return [Array<Hash>] The list of backups
    def list_backups
      @client.list_backups(database: @name)
    end
  end

  class Set
    attr_reader :database, :name

    # Initialize a new Set instance
    #
    # @param database [Database] The database this set belongs to
    # @param name [String] The name of the set
    def initialize(database, name)
      @database = database
      @name = name
    end

    # Get a value from the set
    #
    # @param key [String] The key to get
    # @return [Object] The value
    def get(key)
      @database.client.get(@database.name, @name, key, auth: @database.auth)
    end

    # Put a value into the set
    #
    # @param key [String] The key to put
    # @param value [Object] The value to put
    # @return [Hash] The response from the server
    def put(key, value)
      @database.client.put(@database.name, @name, key, value, auth: @database.auth)
    end

    # Delete a value from the set
    #
    # @param key [String] The key to delete
    # @return [Hash] The response from the server
    def delete(key)
      @database.client.delete(@database.name, @name, key, auth: @database.auth)
    end

    # Create an index on a field in this set
    #
    # @param index_name [String] The name of the index
    # @param field [String] The field to index
    # @return [Hash] The response from the server
    def create_index(index_name, field)
      @database.create_index(@name, index_name, field)
    end

    # Drop an index from this set
    #
    # @param index_name [String] The name of the index
    # @return [Hash] The response from the server
    def drop_index(index_name)
      @database.drop_index(@name, index_name)
    end

    # Query an index on this set
    #
    # @param index_name [String] The name of the index
    # @param value [String] The value to query
    # @param sort [String] The sort order ('asc' or 'desc')
    # @return [Hash] The query results
    def query_index(index_name, value, sort: 'asc')
      @database.query_index(@name, index_name, value, sort: sort)
    end
  end
end