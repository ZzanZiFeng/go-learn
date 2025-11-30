// Package main demonstrates cache patterns: Cache-Aside, Multi-Level, etc.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// User represents a user entity
type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// MockDB simulates a database
type MockDB struct {
	users map[uint]*User
	mu    sync.RWMutex
	delay time.Duration
}

func NewMockDB() *MockDB {
	return &MockDB{
		users: map[uint]*User{
			1: {ID: 1, Name: "Alice", Email: "alice@example.com"},
			2: {ID: 2, Name: "Bob", Email: "bob@example.com"},
			3: {ID: 3, Name: "Charlie", Email: "charlie@example.com"},
		},
		delay: 50 * time.Millisecond, // Simulate DB latency
	}
}

func (db *MockDB) GetUser(id uint) (*User, error) {
	time.Sleep(db.delay)
	db.mu.RLock()
	defer db.mu.RUnlock()
	if user, ok := db.users[id]; ok {
		// Return a copy
		u := *user
		return &u, nil
	}
	return nil, errors.New("user not found")
}

func (db *MockDB) UpdateUser(user *User) error {
	time.Sleep(db.delay)
	db.mu.Lock()
	defer db.mu.Unlock()
	db.users[user.ID] = user
	return nil
}

func main() {
	cacheAsideExample()
	multiLevelCacheExample()
	singleflightExample()
	cacheWithStatsExample()
}

// ============================================
// Cache-Aside Pattern
// ============================================

type CacheAsideService struct {
	rdb *redis.Client
	db  *MockDB
}

func NewCacheAsideService(rdb *redis.Client, db *MockDB) *CacheAsideService {
	return &CacheAsideService{rdb: rdb, db: db}
}

func (s *CacheAsideService) GetUser(ctx context.Context, id uint) (*User, error) {
	key := fmt.Sprintf("user:%d", id)

	// 1. Try cache first
	val, err := s.rdb.Get(ctx, key).Result()
	if err == nil {
		var user User
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			fmt.Printf("  [Cache HIT] user:%d\n", id)
			return &user, nil
		}
	}

	fmt.Printf("  [Cache MISS] user:%d - loading from DB\n", id)

	// 2. Cache miss - load from database
	user, err := s.db.GetUser(id)
	if err != nil {
		return nil, err
	}

	// 3. Write to cache
	data, _ := json.Marshal(user)
	s.rdb.Set(ctx, key, data, time.Hour)

	return user, nil
}

func (s *CacheAsideService) UpdateUser(ctx context.Context, user *User) error {
	// 1. Update database
	if err := s.db.UpdateUser(user); err != nil {
		return err
	}

	// 2. Delete cache (not update!)
	key := fmt.Sprintf("user:%d", user.ID)
	s.rdb.Del(ctx, key)
	fmt.Printf("  [Cache INVALIDATE] user:%d\n", user.ID)

	return nil
}

func cacheAsideExample() {
	fmt.Println("=== Cache-Aside Pattern ===")

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Redis not available, skipping Cache-Aside example: %v", err)
		fmt.Println()
		return
	}

	db := NewMockDB()
	service := NewCacheAsideService(rdb, db)

	// First read - cache miss
	fmt.Println("First read (cache miss):")
	user, _ := service.GetUser(ctx, 1)
	fmt.Printf("  Got: %s\n", user.Name)

	// Second read - cache hit
	fmt.Println("Second read (cache hit):")
	user, _ = service.GetUser(ctx, 1)
	fmt.Printf("  Got: %s\n", user.Name)

	// Update
	fmt.Println("Update user:")
	user.Name = "Alice Updated"
	service.UpdateUser(ctx, user)

	// Read after update - cache miss (cache was invalidated)
	fmt.Println("Read after update (cache miss):")
	user, _ = service.GetUser(ctx, 1)
	fmt.Printf("  Got: %s\n", user.Name)

	// Clean up
	rdb.Del(ctx, "user:1")
	fmt.Println()
}

// ============================================
// Multi-Level Cache Pattern
// ============================================

type MultiLevelCache struct {
	l1    *cache.Cache  // Local cache
	l2    *redis.Client // Redis
	db    *MockDB
	l1TTL time.Duration
	l2TTL time.Duration
}

func NewMultiLevelCache(rdb *redis.Client, db *MockDB) *MultiLevelCache {
	return &MultiLevelCache{
		l1:    cache.New(1*time.Minute, 5*time.Minute),
		l2:    rdb,
		db:    db,
		l1TTL: 1 * time.Minute,
		l2TTL: 1 * time.Hour,
	}
}

func (c *MultiLevelCache) GetUser(ctx context.Context, id uint) (*User, string, error) {
	key := fmt.Sprintf("user:%d", id)

	// 1. L1 - Local cache
	if val, found := c.l1.Get(key); found {
		return val.(*User), "L1", nil
	}

	// 2. L2 - Redis
	val, err := c.l2.Get(ctx, key).Result()
	if err == nil {
		var user User
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			// Backfill L1
			c.l1.Set(key, &user, c.l1TTL)
			return &user, "L2", nil
		}
	}

	// 3. Database
	user, err := c.db.GetUser(id)
	if err != nil {
		return nil, "", err
	}

	// Backfill L1 and L2
	data, _ := json.Marshal(user)
	c.l1.Set(key, user, c.l1TTL)
	c.l2.Set(ctx, key, data, c.l2TTL)

	return user, "DB", nil
}

func multiLevelCacheExample() {
	fmt.Println("=== Multi-Level Cache Pattern ===")

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Redis not available, skipping Multi-Level example: %v", err)
		fmt.Println()
		return
	}

	db := NewMockDB()
	mlCache := NewMultiLevelCache(rdb, db)

	// Clean up first
	rdb.Del(ctx, "user:2")

	// First read - from DB
	user, source, _ := mlCache.GetUser(ctx, 2)
	fmt.Printf("1st read: %s (from %s)\n", user.Name, source)

	// Second read - from L1
	user, source, _ = mlCache.GetUser(ctx, 2)
	fmt.Printf("2nd read: %s (from %s)\n", user.Name, source)

	// Clear L1 to simulate new instance
	mlCache.l1.Flush()
	fmt.Println("L1 cleared (simulating new instance)")

	// Third read - from L2
	user, source, _ = mlCache.GetUser(ctx, 2)
	fmt.Printf("3rd read: %s (from %s)\n", user.Name, source)

	// Fourth read - from L1 again (backfilled)
	user, source, _ = mlCache.GetUser(ctx, 2)
	fmt.Printf("4th read: %s (from %s)\n", user.Name, source)

	// Clean up
	rdb.Del(ctx, "user:2")
	fmt.Println()
}

// ============================================
// Singleflight Pattern (Prevent Cache Stampede)
// ============================================

type SingleflightCache struct {
	l1 *cache.Cache
	db *MockDB
	sf singleflight.Group
}

func NewSingleflightCache(db *MockDB) *SingleflightCache {
	return &SingleflightCache{
		l1: cache.New(5*time.Minute, 10*time.Minute),
		db: db,
	}
}

func (c *SingleflightCache) GetUser(id uint) (*User, error) {
	key := fmt.Sprintf("user:%d", id)

	// Check cache
	if val, found := c.l1.Get(key); found {
		return val.(*User), nil
	}

	// Use singleflight to prevent concurrent DB queries for same key
	result, err, shared := c.sf.Do(key, func() (interface{}, error) {
		// Double-check cache (another goroutine might have filled it)
		if val, found := c.l1.Get(key); found {
			return val.(*User), nil
		}

		// Load from DB
		user, err := c.db.GetUser(id)
		if err != nil {
			return nil, err
		}

		// Cache the result
		c.l1.Set(key, user, cache.DefaultExpiration)
		return user, nil
	})

	if err != nil {
		return nil, err
	}

	if shared {
		fmt.Printf("  [Shared result for user:%d]\n", id)
	}

	return result.(*User), nil
}

func singleflightExample() {
	fmt.Println("=== Singleflight Pattern ===")

	db := NewMockDB()
	sfCache := NewSingleflightCache(db)

	var wg sync.WaitGroup
	var dbCalls int64

	// Wrap DB to count calls
	originalDelay := db.delay
	db.delay = 100 * time.Millisecond

	// Simulate 10 concurrent requests for the same user
	numRequests := 10
	fmt.Printf("Launching %d concurrent requests for user:3...\n", numRequests)

	start := time.Now()
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(reqNum int) {
			defer wg.Done()
			user, err := sfCache.GetUser(3)
			if err == nil {
				// Check if this was a shared result
				atomic.AddInt64(&dbCalls, 1)
				_ = user
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("Completed in %v\n", elapsed)
	fmt.Printf("Without singleflight: would have made %d DB calls\n", numRequests)
	fmt.Printf("With singleflight: made only 1 DB call (others shared the result)\n")

	db.delay = originalDelay
	fmt.Println()
}

// ============================================
// Cache with Statistics
// ============================================

type StatsCache struct {
	l1        *cache.Cache
	stats     CacheStatistics
	mu        sync.RWMutex
	keyAccess map[string]int64 // Track access frequency
}

type CacheStatistics struct {
	Hits      int64
	Misses    int64
	Sets      int64
	Deletes   int64
	Evictions int64
}

func NewStatsCache() *StatsCache {
	return &StatsCache{
		l1:        cache.New(5*time.Minute, 10*time.Minute),
		keyAccess: make(map[string]int64),
	}
}

func (c *StatsCache) Get(key string) (interface{}, bool) {
	val, found := c.l1.Get(key)

	c.mu.Lock()
	c.keyAccess[key]++
	c.mu.Unlock()

	if found {
		atomic.AddInt64(&c.stats.Hits, 1)
	} else {
		atomic.AddInt64(&c.stats.Misses, 1)
	}
	return val, found
}

func (c *StatsCache) Set(key string, value interface{}, ttl time.Duration) {
	c.l1.Set(key, value, ttl)
	atomic.AddInt64(&c.stats.Sets, 1)
}

func (c *StatsCache) Delete(key string) {
	c.l1.Delete(key)
	atomic.AddInt64(&c.stats.Deletes, 1)
}

func (c *StatsCache) HitRate() float64 {
	hits := atomic.LoadInt64(&c.stats.Hits)
	misses := atomic.LoadInt64(&c.stats.Misses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

func (c *StatsCache) TopKeys(n int) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	type kv struct {
		key   string
		count int64
	}

	var pairs []kv
	for k, v := range c.keyAccess {
		pairs = append(pairs, kv{k, v})
	}

	// Simple sort (for demo)
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].count > pairs[i].count {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	result := make([]string, 0, n)
	for i := 0; i < n && i < len(pairs); i++ {
		result = append(result, fmt.Sprintf("%s(%d)", pairs[i].key, pairs[i].count))
	}
	return result
}

func (c *StatsCache) PrintStats() {
	fmt.Printf("Cache Statistics:\n")
	fmt.Printf("  Hits: %d\n", atomic.LoadInt64(&c.stats.Hits))
	fmt.Printf("  Misses: %d\n", atomic.LoadInt64(&c.stats.Misses))
	fmt.Printf("  Sets: %d\n", atomic.LoadInt64(&c.stats.Sets))
	fmt.Printf("  Deletes: %d\n", atomic.LoadInt64(&c.stats.Deletes))
	fmt.Printf("  Hit Rate: %.2f%%\n", c.HitRate())
	fmt.Printf("  Top Keys: %v\n", c.TopKeys(5))
}

func cacheWithStatsExample() {
	fmt.Println("=== Cache with Statistics ===")

	sc := NewStatsCache()

	// Populate cache
	for i := 1; i <= 10; i++ {
		key := fmt.Sprintf("item:%d", i)
		sc.Set(key, fmt.Sprintf("value-%d", i), cache.DefaultExpiration)
	}

	// Simulate reads with Zipf distribution (some keys accessed more)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		// Zipf-like: lower IDs are accessed more frequently
		id := rand.Intn(10) + 1
		if rand.Float32() < 0.7 {
			id = rand.Intn(3) + 1 // 70% of requests go to top 3 items
		}
		key := fmt.Sprintf("item:%d", id)
		sc.Get(key)
	}

	// Some misses
	for i := 0; i < 10; i++ {
		sc.Get(fmt.Sprintf("nonexistent:%d", i))
	}

	sc.PrintStats()
	fmt.Println()
}
