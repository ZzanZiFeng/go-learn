// Package main demonstrates local cache usage with go-cache
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/patrickmn/go-cache"
)

// User represents a user entity
type User struct {
	ID    uint
	Name  string
	Email string
}

// CacheStats tracks cache performance
type CacheStats struct {
	Hits   int64
	Misses int64
}

// StatsCache wraps go-cache with statistics
type StatsCache struct {
	cache *cache.Cache
	stats CacheStats
}

func main() {
	basicExample()
	advancedExample()
	statsExample()
	concurrencyExample()
	genericCacheExample()
}

// basicExample demonstrates basic go-cache operations
func basicExample() {
	fmt.Println("=== Basic go-cache Example ===")

	// Create cache with default expiration of 5 minutes and cleanup interval of 10 minutes
	c := cache.New(5*time.Minute, 10*time.Minute)

	// Set with default expiration
	c.Set("name", "Alice", cache.DefaultExpiration)

	// Set with custom expiration
	c.Set("session", "abc123", 30*time.Minute)

	// Set with no expiration
	c.Set("config", map[string]string{"env": "prod"}, cache.NoExpiration)

	// Get
	if val, found := c.Get("name"); found {
		fmt.Printf("name: %v\n", val)
	}

	// Get with type assertion
	if val, found := c.Get("config"); found {
		config := val.(map[string]string)
		fmt.Printf("config: %v\n", config)
	}

	// Delete
	c.Delete("name")

	// Check if exists
	if _, found := c.Get("name"); !found {
		fmt.Println("name was deleted")
	}

	// Get all items
	items := c.Items()
	fmt.Printf("Total items: %d\n", len(items))

	// ItemCount
	fmt.Printf("Item count: %d\n", c.ItemCount())

	fmt.Println()
}

// advancedExample demonstrates advanced operations
func advancedExample() {
	fmt.Println("=== Advanced Operations ===")

	c := cache.New(5*time.Minute, 10*time.Minute)

	// Add - only if not exists
	err := c.Add("key1", "value1", cache.DefaultExpiration)
	if err == nil {
		fmt.Println("key1 added successfully")
	}

	err = c.Add("key1", "value2", cache.DefaultExpiration)
	if err != nil {
		fmt.Printf("Add failed: %v\n", err)
	}

	// Replace - only if exists
	err = c.Replace("key1", "new_value", cache.DefaultExpiration)
	if err == nil {
		val, _ := c.Get("key1")
		fmt.Printf("key1 replaced: %v\n", val)
	}

	err = c.Replace("nonexistent", "value", cache.DefaultExpiration)
	if err != nil {
		fmt.Printf("Replace failed: %v\n", err)
	}

	// Increment/Decrement
	c.Set("counter", 0, cache.NoExpiration)

	newVal, err := c.Increment("counter", 1)
	if err == nil {
		fmt.Printf("counter after Increment: %d\n", newVal)
	}

	newVal, _ = c.IncrementInt("counter", 10)
	fmt.Printf("counter after IncrementInt(10): %d\n", newVal)

	newVal, _ = c.Decrement("counter", 5)
	fmt.Printf("counter after Decrement(5): %d\n", newVal)

	// Float increment
	c.Set("balance", 100.0, cache.NoExpiration)
	newFloat, _ := c.IncrementFloat("balance", 10.5)
	fmt.Printf("balance after IncrementFloat: %.2f\n", newFloat)

	// SetDefault (uses default expiration)
	c.SetDefault("key2", "value2")

	// Flush all
	c.Flush()
	fmt.Printf("After Flush, item count: %d\n", c.ItemCount())

	fmt.Println()
}

// statsExample demonstrates cache with statistics
func statsExample() {
	fmt.Println("=== Cache with Statistics ===")

	sc := &StatsCache{
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}

	// Set some users
	users := []User{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
		{ID: 3, Name: "Charlie", Email: "charlie@example.com"},
	}

	for _, user := range users {
		key := fmt.Sprintf("user:%d", user.ID)
		sc.cache.Set(key, user, cache.DefaultExpiration)
	}

	// Simulate reads
	keys := []string{"user:1", "user:2", "user:1", "user:3", "user:99", "user:1"}
	for _, key := range keys {
		if val, found := sc.cache.Get(key); found {
			atomic.AddInt64(&sc.stats.Hits, 1)
			user := val.(User)
			fmt.Printf("Hit: %s -> %s\n", key, user.Name)
		} else {
			atomic.AddInt64(&sc.stats.Misses, 1)
			fmt.Printf("Miss: %s\n", key)
		}
	}

	// Print stats
	hits := atomic.LoadInt64(&sc.stats.Hits)
	misses := atomic.LoadInt64(&sc.stats.Misses)
	total := hits + misses
	hitRate := float64(hits) / float64(total) * 100

	fmt.Printf("\nCache Statistics:\n")
	fmt.Printf("  Hits: %d\n", hits)
	fmt.Printf("  Misses: %d\n", misses)
	fmt.Printf("  Hit Rate: %.2f%%\n", hitRate)

	fmt.Println()
}

// concurrencyExample demonstrates concurrent cache access
func concurrencyExample() {
	fmt.Println("=== Concurrent Access ===")

	c := cache.New(5*time.Minute, 10*time.Minute)
	c.Set("counter", 0, cache.NoExpiration)

	var wg sync.WaitGroup
	numGoroutines := 100
	incrementsPerGoroutine := 100

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				c.Increment("counter", 1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	val, _ := c.Get("counter")
	expected := numGoroutines * incrementsPerGoroutine
	fmt.Printf("Final counter value: %v (expected: %d)\n", val, expected)
	fmt.Printf("Time elapsed: %v\n", elapsed)

	fmt.Println()
}

// GenericCache provides type-safe cache operations
type GenericCache[T any] struct {
	cache *cache.Cache
}

// NewGenericCache creates a new generic cache
func NewGenericCache[T any](defaultExpiration, cleanupInterval time.Duration) *GenericCache[T] {
	return &GenericCache[T]{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Set stores a value
func (c *GenericCache[T]) Set(key string, value T, expiration time.Duration) {
	c.cache.Set(key, value, expiration)
}

// Get retrieves a value
func (c *GenericCache[T]) Get(key string) (T, bool) {
	var zero T
	val, found := c.cache.Get(key)
	if !found {
		return zero, false
	}
	return val.(T), true
}

// Delete removes a value
func (c *GenericCache[T]) Delete(key string) {
	c.cache.Delete(key)
}

// GetOrSet retrieves a value or loads it using the provided function
func (c *GenericCache[T]) GetOrSet(key string, loader func() (T, error), expiration time.Duration) (T, error) {
	if val, found := c.Get(key); found {
		return val, nil
	}

	value, err := loader()
	if err != nil {
		var zero T
		return zero, err
	}

	c.Set(key, value, expiration)
	return value, nil
}

// genericCacheExample demonstrates type-safe generic cache
func genericCacheExample() {
	fmt.Println("=== Generic Cache Example ===")

	// Create a cache for User type
	userCache := NewGenericCache[*User](5*time.Minute, 10*time.Minute)

	// Set users
	userCache.Set("user:1", &User{ID: 1, Name: "Alice", Email: "alice@example.com"}, cache.DefaultExpiration)
	userCache.Set("user:2", &User{ID: 2, Name: "Bob", Email: "bob@example.com"}, cache.DefaultExpiration)

	// Get user (no type assertion needed!)
	if user, found := userCache.Get("user:1"); found {
		fmt.Printf("Found user: %s (%s)\n", user.Name, user.Email)
	}

	// GetOrSet with loader
	user, err := userCache.GetOrSet("user:3", func() (*User, error) {
		// Simulate database load
		fmt.Println("Loading user:3 from 'database'...")
		return &User{ID: 3, Name: "Charlie", Email: "charlie@example.com"}, nil
	}, cache.DefaultExpiration)

	if err == nil {
		fmt.Printf("Got user: %s\n", user.Name)
	}

	// Second call uses cache
	user, _ = userCache.GetOrSet("user:3", func() (*User, error) {
		fmt.Println("This should not print - using cache")
		return nil, nil
	}, cache.DefaultExpiration)
	fmt.Printf("Got cached user: %s\n", user.Name)

	// Create a cache for simple strings
	stringCache := NewGenericCache[string](time.Minute, 5*time.Minute)
	stringCache.Set("greeting", "Hello, World!", cache.DefaultExpiration)

	if greeting, found := stringCache.Get("greeting"); found {
		fmt.Printf("Greeting: %s\n", greeting)
	}

	fmt.Println()
}
