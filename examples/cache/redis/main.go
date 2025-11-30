// Package main demonstrates go-redis client usage
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// User represents a user entity
type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func main() {
	ctx := context.Background()

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		PoolSize: 10,
	})
	defer rdb.Close()

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("✓ Connected to Redis")

	// Run examples
	stringExample(ctx, rdb)
	hashExample(ctx, rdb)
	listExample(ctx, rdb)
	setExample(ctx, rdb)
	sortedSetExample(ctx, rdb)
	pipelineExample(ctx, rdb)
	transactionExample(ctx, rdb)
	cacheExample(ctx, rdb)
}

// stringExample demonstrates String operations
func stringExample(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n=== String Operations ===")

	// Set
	err := rdb.Set(ctx, "greeting", "Hello, Redis!", time.Hour).Err()
	if err != nil {
		log.Fatal(err)
	}

	// Get
	val, err := rdb.Get(ctx, "greeting").Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("greeting: %s\n", val)

	// SetNX (only if not exists)
	ok, _ := rdb.SetNX(ctx, "lock:order:123", "1", 10*time.Second).Result()
	fmt.Printf("SetNX lock acquired: %v\n", ok)

	// Increment
	rdb.Set(ctx, "counter", 0, 0)
	n, _ := rdb.Incr(ctx, "counter").Result()
	fmt.Printf("counter after Incr: %d\n", n)

	n, _ = rdb.IncrBy(ctx, "counter", 10).Result()
	fmt.Printf("counter after IncrBy(10): %d\n", n)

	// MSet / MGet
	rdb.MSet(ctx, "name", "Alice", "age", "25", "city", "NYC")
	vals, _ := rdb.MGet(ctx, "name", "age", "city").Result()
	fmt.Printf("MGet: %v\n", vals)

	// Clean up
	rdb.Del(ctx, "greeting", "lock:order:123", "counter", "name", "age", "city")
}

// hashExample demonstrates Hash operations
func hashExample(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n=== Hash Operations ===")

	// HSet - single field
	rdb.HSet(ctx, "user:1", "name", "Alice")

	// HSet - multiple fields
	rdb.HSet(ctx, "user:1", map[string]interface{}{
		"email": "alice@example.com",
		"age":   25,
	})

	// HGet
	name, _ := rdb.HGet(ctx, "user:1", "name").Result()
	fmt.Printf("user:1 name: %s\n", name)

	// HGetAll
	fields, _ := rdb.HGetAll(ctx, "user:1").Result()
	fmt.Printf("user:1 all fields: %v\n", fields)

	// HMGet
	vals, _ := rdb.HMGet(ctx, "user:1", "name", "email").Result()
	fmt.Printf("HMGet name, email: %v\n", vals)

	// HIncrBy
	newAge, _ := rdb.HIncrBy(ctx, "user:1", "age", 1).Result()
	fmt.Printf("age after HIncrBy: %d\n", newAge)

	// HExists
	exists, _ := rdb.HExists(ctx, "user:1", "name").Result()
	fmt.Printf("name field exists: %v\n", exists)

	// Clean up
	rdb.Del(ctx, "user:1")
}

// listExample demonstrates List operations
func listExample(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n=== List Operations ===")

	// LPush
	rdb.LPush(ctx, "tasks", "task1", "task2", "task3")

	// RPush
	rdb.RPush(ctx, "tasks", "task4")

	// LRange
	tasks, _ := rdb.LRange(ctx, "tasks", 0, -1).Result()
	fmt.Printf("All tasks: %v\n", tasks)

	// LLen
	length, _ := rdb.LLen(ctx, "tasks").Result()
	fmt.Printf("Tasks count: %d\n", length)

	// LPop
	task, _ := rdb.LPop(ctx, "tasks").Result()
	fmt.Printf("LPop: %s\n", task)

	// RPop
	task, _ = rdb.RPop(ctx, "tasks").Result()
	fmt.Printf("RPop: %s\n", task)

	// Clean up
	rdb.Del(ctx, "tasks")
}

// setExample demonstrates Set operations
func setExample(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n=== Set Operations ===")

	// SAdd
	rdb.SAdd(ctx, "tags", "go", "redis", "cache", "database")

	// SMembers
	members, _ := rdb.SMembers(ctx, "tags").Result()
	fmt.Printf("All tags: %v\n", members)

	// SIsMember
	isMember, _ := rdb.SIsMember(ctx, "tags", "go").Result()
	fmt.Printf("'go' is member: %v\n", isMember)

	// SCard
	count, _ := rdb.SCard(ctx, "tags").Result()
	fmt.Printf("Tags count: %d\n", count)

	// Set operations
	rdb.SAdd(ctx, "set1", "a", "b", "c")
	rdb.SAdd(ctx, "set2", "b", "c", "d")

	inter, _ := rdb.SInter(ctx, "set1", "set2").Result()
	fmt.Printf("Intersection: %v\n", inter)

	union, _ := rdb.SUnion(ctx, "set1", "set2").Result()
	fmt.Printf("Union: %v\n", union)

	diff, _ := rdb.SDiff(ctx, "set1", "set2").Result()
	fmt.Printf("Difference (set1 - set2): %v\n", diff)

	// Clean up
	rdb.Del(ctx, "tags", "set1", "set2")
}

// sortedSetExample demonstrates Sorted Set operations
func sortedSetExample(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n=== Sorted Set Operations ===")

	// ZAdd
	rdb.ZAdd(ctx, "leaderboard",
		redis.Z{Score: 100, Member: "Alice"},
		redis.Z{Score: 200, Member: "Bob"},
		redis.Z{Score: 150, Member: "Charlie"},
	)

	// ZRange (ascending)
	members, _ := rdb.ZRange(ctx, "leaderboard", 0, -1).Result()
	fmt.Printf("Leaderboard (asc): %v\n", members)

	// ZRevRange (descending)
	members, _ = rdb.ZRevRange(ctx, "leaderboard", 0, -1).Result()
	fmt.Printf("Leaderboard (desc): %v\n", members)

	// ZRevRangeWithScores
	results, _ := rdb.ZRevRangeWithScores(ctx, "leaderboard", 0, -1).Result()
	fmt.Println("Leaderboard with scores:")
	for i, z := range results {
		fmt.Printf("  %d. %s: %.0f\n", i+1, z.Member, z.Score)
	}

	// ZScore
	score, _ := rdb.ZScore(ctx, "leaderboard", "Alice").Result()
	fmt.Printf("Alice's score: %.0f\n", score)

	// ZRank / ZRevRank
	rank, _ := rdb.ZRevRank(ctx, "leaderboard", "Alice").Result()
	fmt.Printf("Alice's rank (0-based): %d\n", rank)

	// ZIncrBy
	newScore, _ := rdb.ZIncrBy(ctx, "leaderboard", 50, "Alice").Result()
	fmt.Printf("Alice's new score after +50: %.0f\n", newScore)

	// Clean up
	rdb.Del(ctx, "leaderboard")
}

// pipelineExample demonstrates Pipeline usage
func pipelineExample(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n=== Pipeline Example ===")

	pipe := rdb.Pipeline()

	// Queue multiple commands
	incr := pipe.Incr(ctx, "pipe_counter")
	pipe.Expire(ctx, "pipe_counter", time.Hour)
	get := pipe.Get(ctx, "pipe_counter")

	// Execute all at once
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		log.Printf("Pipeline error: %v", err)
	}

	fmt.Printf("Incr result: %d\n", incr.Val())
	fmt.Printf("Get result: %s\n", get.Val())

	// Clean up
	rdb.Del(ctx, "pipe_counter")
}

// transactionExample demonstrates Transaction usage
func transactionExample(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n=== Transaction Example ===")

	// TxPipeline (MULTI/EXEC)
	tx := rdb.TxPipeline()

	tx.Set(ctx, "tx_key1", "value1", 0)
	tx.Set(ctx, "tx_key2", "value2", 0)
	tx.Incr(ctx, "tx_counter")

	_, err := tx.Exec(ctx)
	if err != nil {
		log.Printf("Transaction error: %v", err)
	}

	val1, _ := rdb.Get(ctx, "tx_key1").Result()
	val2, _ := rdb.Get(ctx, "tx_key2").Result()
	counter, _ := rdb.Get(ctx, "tx_counter").Result()
	fmt.Printf("tx_key1: %s, tx_key2: %s, tx_counter: %s\n", val1, val2, counter)

	// Watch example (optimistic locking)
	err = rdb.Watch(ctx, func(tx *redis.Tx) error {
		// Get current balance
		balance, err := tx.Get(ctx, "balance").Int()
		if err != nil && err != redis.Nil {
			return err
		}

		// Execute transaction
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, "balance", balance+100, 0)
			return nil
		})
		return err
	}, "balance")

	if err == redis.TxFailedErr {
		fmt.Println("Transaction failed due to concurrent modification")
	} else if err != nil {
		log.Printf("Watch error: %v", err)
	} else {
		balance, _ := rdb.Get(ctx, "balance").Result()
		fmt.Printf("Balance after transaction: %s\n", balance)
	}

	// Clean up
	rdb.Del(ctx, "tx_key1", "tx_key2", "tx_counter", "balance")
}

// cacheExample demonstrates caching patterns
func cacheExample(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n=== Cache Example ===")

	// Cache a user object
	user := User{
		ID:    1,
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   25,
	}

	// Serialize and cache
	key := fmt.Sprintf("user:%d", user.ID)
	data, _ := json.Marshal(user)
	rdb.Set(ctx, key, data, time.Hour)
	fmt.Printf("Cached user: %s\n", key)

	// Retrieve and deserialize
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		fmt.Println("Cache miss")
	} else if err != nil {
		log.Fatal(err)
	} else {
		var cachedUser User
		json.Unmarshal([]byte(val), &cachedUser)
		fmt.Printf("Retrieved user: %+v\n", cachedUser)
	}

	// Check TTL
	ttl, _ := rdb.TTL(ctx, key).Result()
	fmt.Printf("TTL: %v\n", ttl)

	// Clean up
	rdb.Del(ctx, key)
}
