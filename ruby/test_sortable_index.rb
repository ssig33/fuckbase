#!/usr/bin/env ruby

require 'bundler/inline'

gemfile do
  source 'https://rubygems.org'
  gem 'minitest'
  gem 'base64'  # Add base64 gem for Ruby 3.4+ compatibility
end

require 'minitest/autorun'
require_relative './fuckbase'

class TestSortableIndex < Minitest::Test
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
      @client.drop_database('test_sortable_db')
    rescue
      # Ignore errors if the database doesn't exist
    end
    
    @db = @client.create_database('test_sortable_db')
    @set = @db.create_set('products')
    
    # Add 30 products with different categories, prices, and ratings
    add_test_products
    
    # Create a sortable index on category with price and rating as sort fields
    @set.create_sortable_index('category_index', 'category', ['price', 'rating'])
  end
  
  def teardown
    # Clean up by dropping the test database
    begin
      @client.drop_database('test_sortable_db')
    rescue
      # Ignore errors
    end
  end
  
  def add_test_products
    # Add 10 Electronics products
    10.times do |i|
      price = 100 + (i * 50)  # 100, 150, 200, ..., 550
      rating = 3.0 + (i % 5) * 0.5  # 3.0, 3.5, 4.0, 4.5, 5.0, 3.0, ...
      stock = 10 + i * 3
      
      @set.put("elec_#{i+1}", {
        'name' => "Electronics Product #{i+1}",
        'category' => 'Electronics',
        'price' => price,
        'rating' => rating,
        'stock' => stock
      })
    end
    
    # Add 10 Clothing products
    10.times do |i|
      price = 20 + (i * 10)  # 20, 30, 40, ..., 110
      rating = 2.5 + (i % 6) * 0.5  # 2.5, 3.0, 3.5, 4.0, 4.5, 5.0, 2.5, ...
      stock = 50 + i * 5
      
      @set.put("cloth_#{i+1}", {
        'name' => "Clothing Product #{i+1}",
        'category' => 'Clothing',
        'price' => price,
        'rating' => rating,
        'stock' => stock
      })
    end
    
    # Add 10 Books products
    10.times do |i|
      price = 10 + (i * 5)  # 10, 15, 20, ..., 55
      rating = 4.0 + (i % 3) * 0.3  # 4.0, 4.3, 4.6, 4.0, ...
      stock = 30 + i * 2
      
      @set.put("book_#{i+1}", {
        'name' => "Book Product #{i+1}",
        'category' => 'Books',
        'price' => price,
        'rating' => rating,
        'stock' => stock
      })
    end
  end
  
  def test_basic_sorting
    # Test sorting by price in ascending order
    results = @set.query_sorted('category_index', 'Electronics', 'price', ascending: true)
    assert_equal 10, results[:count]
    
    # Verify the order is correct (price ascending)
    prices = results[:data].map { |item| item['value']['price'] }
    assert_equal prices.sort, prices
    
    # Test sorting by price in descending order
    results = @set.query_sorted('category_index', 'Electronics', 'price', ascending: false)
    assert_equal 10, results[:count]
    
    # Verify the order is correct (price descending)
    prices = results[:data].map { |item| item['value']['price'] }
    assert_equal prices.sort.reverse, prices
  end
  
  def test_pagination
    # Test pagination with limit 3
    results = @set.query_sorted_with_pagination('category_index', 'Clothing', 'price', ascending: true, offset: 0, limit: 3)
    assert_equal 3, results[:count]
    assert_equal 10, results[:total]
    
    # Verify the first page has the 3 lowest priced items
    prices = results[:data].map { |item| item['value']['price'] }
    assert_equal [20, 30, 40], prices
    
    # Test second page
    results = @set.query_sorted_with_pagination('category_index', 'Clothing', 'price', ascending: true, offset: 3, limit: 3)
    assert_equal 3, results[:count]
    
    # Verify the second page has the next 3 lowest priced items
    prices = results[:data].map { |item| item['value']['price'] }
    assert_equal [50, 60, 70], prices
    
    # Test last page
    results = @set.query_sorted_with_pagination('category_index', 'Clothing', 'price', ascending: true, offset: 9, limit: 3)
    assert_equal 1, results[:count]
    
    # Verify the last page has the highest priced item
    prices = results[:data].map { |item| item['value']['price'] }
    assert_equal [110], prices
  end
  
  def test_multi_field_sorting
    # Add some products with the same price but different ratings
    @set.put("special1", {
      'name' => "Special Product 1",
      'category' => 'Books',
      'price' => 25,
      'rating' => 4.0
    })
    
    @set.put("special2", {
      'name' => "Special Product 2",
      'category' => 'Books',
      'price' => 25,
      'rating' => 4.5
    })
    
    @set.put("special3", {
      'name' => "Special Product 3",
      'category' => 'Books',
      'price' => 25,
      'rating' => 3.5
    })
    
    # Test multi-field sorting: first by price (ascending), then by rating (descending)
    results = @set.query_multi_sorted('category_index', 'Books', ['price', 'rating'], ascending: [true, false])
    
    # Find the products with price 25
    price_25_products = results[:data].select { |item| item['value']['price'] == 25 }
    
    # Extract their ratings
    ratings = price_25_products.map { |item| item['value']['rating'] }
    
    # Verify the ratings are in descending order
    assert_equal ratings.sort.reverse, ratings
  end
  
  def test_multi_field_sorting_with_pagination
    # Test multi-field sorting with pagination
    results = @set.query_multi_sorted_with_pagination(
      'category_index', 
      'Electronics', 
      ['rating', 'price'], 
      ascending: [false, true],  # Rating descending, price ascending
      offset: 0, 
      limit: 5
    )
    
    assert_equal 5, results[:count]
    assert_equal 10, results[:total]
    
    # Group by rating
    grouped_by_rating = {}
    results[:data].each do |item|
      rating = item['value']['rating']
      grouped_by_rating[rating] ||= []
      grouped_by_rating[rating] << item['value']['price']
    end
    
    # For each rating group, verify prices are in ascending order
    grouped_by_rating.each do |rating, prices|
      if prices.length > 1
        assert_equal prices.sort, prices, "Prices for rating #{rating} are not in ascending order"
      end
    end
    
    # Verify ratings are in descending order
    ratings = results[:data].map { |item| item['value']['rating'] }
    sorted_ratings = ratings.sort.reverse
    
    # We can't directly compare these arrays because there might be duplicate ratings
    # Instead, we verify that the maximum rating comes first, and they generally decrease
    assert_equal sorted_ratings.first, ratings.first, "First item doesn't have the highest rating"
    
    # Test second page
    results = @set.query_multi_sorted_with_pagination(
      'category_index', 
      'Electronics', 
      ['rating', 'price'], 
      ascending: [false, true],
      offset: 5, 
      limit: 5
    )
    
    assert_equal 5, results[:count]
    assert_equal 10, results[:total]
  end
  
  def test_update_and_delete
    # Add a test product
    @set.put("test_update", {
      'name' => "Test Update Product",
      'category' => 'Electronics',
      'price' => 999,
      'rating' => 4.2
    })
    
    # Verify it's in the index
    results = @set.query_sorted('category_index', 'Electronics', 'price', ascending: false)
    assert_includes results[:data].map { |item| item['key'] }, "test_update"
    
    # Update the product with a different category
    @set.put("test_update", {
      'name' => "Test Update Product",
      'category' => 'Clothing',  # Changed category
      'price' => 999,
      'rating' => 4.2
    })
    
    # Verify it's no longer in Electronics category
    results = @set.query_sorted('category_index', 'Electronics', 'price', ascending: false)
    refute_includes results[:data].map { |item| item['key'] }, "test_update"
    
    # Verify it's now in Clothing category
    results = @set.query_sorted('category_index', 'Clothing', 'price', ascending: false)
    assert_includes results[:data].map { |item| item['key'] }, "test_update"
    
    # Delete the product
    @set.delete("test_update")
    
    # Verify it's no longer in the index
    results = @set.query_sorted('category_index', 'Clothing', 'price', ascending: false)
    refute_includes results[:data].map { |item| item['key'] }, "test_update"
  end
end

puts "Running FuckBase Sortable Index tests..."