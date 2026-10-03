package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// ctx is the background context for Redis operations.
var ctx = context.Background()

func main() {
	// Create a new Redis client.
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password set
		DB:       0,                // Use default DB
	})

	// Ping the server to check the connection.
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Failed to connect to Redis:", err)
		fmt.Println("Please ensure Redis is running on localhost:6379")
		return
	}
	fmt.Println("Successfully connected to Redis!")

	// Run the examples.
	stringExample(rdb)
	hashExample(rdb)
	listExample(rdb)
	expireExample(rdb)
}

// stringExample demonstrates basic SET and GET commands.
func stringExample(rdb *redis.Client) {
	fmt.Println("\n--- String Example ---")

	// Set a key-value pair.
	err := rdb.Set(ctx, "mykey", "Hello, Redis!", 0).Err()
	if err != nil {
		panic(err)
	}

	// Get the value for the key.
	val, err := rdb.Get(ctx, "mykey").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("GET mykey:", val)

	// Get a non-existent key.
	val2, err := rdb.Get(ctx, "nonexistent").Result()
	if err == redis.Nil {
		fmt.Println("GET nonexistent: key does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("nonexistent:", val2)
	}
}

// hashExample demonstrates HSET and HGET commands for storing objects.
func hashExample(rdb *redis.Client) {
	fmt.Println("\n--- Hash Example ---")

	user := map[string]interface{}{"name": "John", "age": 30, "job": "Developer"}

	// Set multiple fields in a hash.
	err := rdb.HSet(ctx, "user:1", user).Err()
	if err != nil {
		panic(err)
	}

	// Get a single field from the hash.
	name, err := rdb.HGet(ctx, "user:1", "name").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("HGET user:1 name:", name)

	// Get all fields from the hash.
	userDetails, err := rdb.HGetAll(ctx, "user:1").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("HGETALL user:1:", userDetails)
}

// listExample demonstrates LPUSH, RPUSH, and LRANGE for managing lists.
func listExample(rdb *redis.Client) {
	fmt.Println("\n--- List Example ---")

	listKey := "tasks"

	// Push items to the left (head) of the list.
	rdb.LPush(ctx, listKey, "task1", "task2")

	// Push an item to the right (tail) of the list.
	rdb.RPush(ctx, listKey, "task3")

	// Get all items from the list.
	tasks, err := rdb.LRange(ctx, listKey, 0, -1).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("LRANGE tasks (0, -1):", tasks)

	// Clean up the list for the next run.
	rdb.Del(ctx, listKey)
}

// expireExample demonstrates how to set an expiration time for a key.
func expireExample(rdb *redis.Client) {
	fmt.Println("\n--- Expire Example ---")

	key := "transient_key"

	// Set a key with a 5-second expiration.
	rdb.Set(ctx, key, "this will disappear soon", 5*time.Second)

	// Check if the key exists.
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		fmt.Println("Error getting key:", err)
	} else {
		fmt.Println("GET transient_key (before expiry):", val)
	}

	// Wait for 6 seconds.
	fmt.Println("Waiting for 6 seconds...")
	time.Sleep(6 * time.Second)

	// Check the key again.
	_, err = rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		fmt.Println("GET transient_key (after expiry): key has expired")
	} else if err != nil {
		panic(err)
	}
}
