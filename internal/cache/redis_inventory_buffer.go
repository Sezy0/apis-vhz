package cache

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var deleteIfUnchangedScript = redis.NewScript(`
	if redis.call("HGET", KEYS[1], ARGV[1]) == ARGV[2] then
		redis.call("HDEL", KEYS[1], ARGV[1])
		redis.call("SREM", KEYS[2], ARGV[1])
		return 1
	else
		return 0
	end
`)

// RedisInventoryBuffer uses Redis for write-behind caching.
// Sync requests are buffered in Redis, then batch-flushed to MySQL.
type RedisInventoryBuffer struct {
	client      *redis.Client
	flushFunc   FlushFunc
	flushTicker *time.Ticker
	stopFlush   chan struct{}
	stopOnce    sync.Once
	keyPrefix   string
}

// RedisBufferConfig holds configuration for Redis buffer.
type RedisBufferConfig struct {
	Addr          string        // Redis address (e.g., "127.0.0.1:6379")
	Password      string        // Redis password (empty if none)
	DB            int           // Redis database number (use different DB per app)
	FlushInterval time.Duration // How often to flush to MySQL
	KeyPrefix     string        // Optional custom key prefix
}

// NewRedisInventoryBuffer creates a Redis-backed inventory buffer.
func NewRedisInventoryBuffer(cfg RedisBufferConfig, flushFunc FlushFunc) (*RedisInventoryBuffer, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB, // Use dedicated DB to avoid conflicts with other apps
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	keyPrefix := cfg.KeyPrefix
	if keyPrefix == "" {
		keyPrefix = "vinzhub:fishit:inventory"
	}

	b := &RedisInventoryBuffer{
		client:      client,
		flushFunc:   flushFunc,
		flushTicker: time.NewTicker(cfg.FlushInterval),
		stopFlush:   make(chan struct{}),
		keyPrefix:   keyPrefix,
	}

	// Start background flush goroutine
	go b.backgroundFlush()

	log.Printf("[RedisInventoryBuffer] Connected to Redis DB:%d, prefix:%s, flush interval: %v", 
		cfg.DB, keyPrefix, cfg.FlushInterval)
	return b, nil
}

// bufferKey returns the namespaced buffer key
func (b *RedisInventoryBuffer) bufferKey() string {
	return b.keyPrefix + ":buffer"
}

// pendingKey returns the namespaced pending set key
func (b *RedisInventoryBuffer) pendingKey() string {
	return b.keyPrefix + ":pending"
}

// Add buffers an inventory update in Redis.
// This is very fast - no MySQL hit!
func (b *RedisInventoryBuffer) Add(ctx context.Context, keyAccountID int64, robloxUserID string, rawJSON []byte) error {
	// Store the inventory data
	data := &BufferedInventory{
		KeyAccountID: keyAccountID,
		RobloxUserID: robloxUserID,
		RawJSON:      rawJSON,
		UpdatedAt:    time.Now(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	pipe := b.client.Pipeline()

	// Store the data with user ID as field
	pipe.HSet(ctx, b.bufferKey(), robloxUserID, jsonData)

	// Add to pending set for tracking
	pipe.SAdd(ctx, b.pendingKey(), robloxUserID)

	_, err = pipe.Exec(ctx)
	return err
}

// Get retrieves a buffered inventory from Redis.
func (b *RedisInventoryBuffer) Get(ctx context.Context, robloxUserID string) (*BufferedInventory, error) {
	data, err := b.client.HGet(ctx, b.bufferKey(), robloxUserID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var inv BufferedInventory
	if err := json.Unmarshal(data, &inv); err != nil {
		return nil, err
	}

	return &inv, nil
}

// Count returns the number of pending items.
func (b *RedisInventoryBuffer) Count(ctx context.Context) (int64, error) {
	return b.client.SCard(ctx, b.pendingKey()).Result()
}

// Flush writes all buffered items to the database.
func (b *RedisInventoryBuffer) Flush(ctx context.Context) error {
	// Get all pending user IDs
	userIDs, err := b.client.SMembers(ctx, b.pendingKey()).Result()
	if err != nil {
		return err
	}

	if len(userIDs) == 0 {
		return nil
	}

	log.Printf("[RedisInventoryBuffer] Flushing %d items to database", len(userIDs))

	// Get all buffered data
	items := make([]*BufferedInventory, 0, len(userIDs))
	// Map to store original JSON strings for safe deletion (Optimistic Locking)
	originalData := make(map[string]string)

	for _, userID := range userIDs {
		data, err := b.client.HGet(ctx, b.bufferKey(), userID).Bytes()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Printf("[RedisInventoryBuffer] Error getting data for %s: %v", userID, err)
			continue
		}

		// Store original data for comparison later
		originalData[userID] = string(data)

		var inv BufferedInventory
		if err := json.Unmarshal(data, &inv); err != nil {
			log.Printf("[RedisInventoryBuffer] Error unmarshaling data for %s: %v", userID, err)
			continue
		}
		items = append(items, &inv)
	}

	if len(items) == 0 {
		return nil
	}

	// Flush to database
	if err := b.flushFunc(ctx, items); err != nil {
		log.Printf("[RedisInventoryBuffer] Flush error: %v", err)
		return err
	}

	// Clear flushed items from Redis using safe atomic script
	pipe := b.client.Pipeline()
	for userID, rawJSON := range originalData {
		// Only delete if the data in Redis hasn't changed since we read it
		deleteIfUnchangedScript.Run(ctx, pipe, []string{b.bufferKey(), b.pendingKey()}, userID, rawJSON)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("[RedisInventoryBuffer] Error clearing Redis: %v", err)
	}

	log.Printf("[RedisInventoryBuffer] Successfully flushed %d items", len(items))
	return nil
}

// backgroundFlush runs the periodic flush to database.
func (b *RedisInventoryBuffer) backgroundFlush() {
	for {
		select {
		case <-b.flushTicker.C:
			// Short timeout to fail fast rather than block for 60s
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			if err := b.Flush(ctx); err != nil {
				log.Printf("[RedisInventoryBuffer] Background flush error: %v", err)
			}
			cancel()
		case <-b.stopFlush:
			// Final flush on shutdown with longer timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			b.Flush(ctx)
			cancel()
			return
		}
	}
}

// Close stops the buffer and performs a final flush.
func (b *RedisInventoryBuffer) Close() error {
	b.stopOnce.Do(func() {
		b.flushTicker.Stop()
		close(b.stopFlush)
	})
	return b.client.Close()
}
