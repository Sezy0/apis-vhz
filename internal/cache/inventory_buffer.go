package cache

import (
	"context"
	"log"
	"sync"
	"time"
)

// InventoryBuffer holds pending inventory updates to be flushed to DB.
// This implements write-behind caching to reduce database connections.
type InventoryBuffer struct {
	mu          sync.RWMutex
	pending     map[string]*BufferedInventory // key: roblox_user_id
	flushFunc   FlushFunc
	flushTicker *time.Ticker
	stopFlush   chan struct{}
}

// BufferedInventory represents a pending inventory update.
type BufferedInventory struct {
	KeyAccountID int64
	RobloxUserID string
	RawJSON      []byte
	UpdatedAt    time.Time
}

// FlushFunc is called to persist buffered data to database.
type FlushFunc func(ctx context.Context, items []*BufferedInventory) error

// NewInventoryBuffer creates a new write-behind buffer.
// flushInterval: how often to flush to database (e.g., 30s)
// flushFunc: function to call when flushing to database
func NewInventoryBuffer(flushInterval time.Duration, flushFunc FlushFunc) *InventoryBuffer {
	b := &InventoryBuffer{
		pending:     make(map[string]*BufferedInventory),
		flushFunc:   flushFunc,
		flushTicker: time.NewTicker(flushInterval),
		stopFlush:   make(chan struct{}),
	}

	// Start background flush goroutine
	go b.backgroundFlush()

	log.Printf("[InventoryBuffer] Started with %v flush interval", flushInterval)
	return b
}

// Add adds or updates an inventory entry in the buffer.
// This is very fast - no database hit!
func (b *InventoryBuffer) Add(keyAccountID int64, robloxUserID string, rawJSON []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Make a copy of the JSON data
	jsonCopy := make([]byte, len(rawJSON))
	copy(jsonCopy, rawJSON)

	b.pending[robloxUserID] = &BufferedInventory{
		KeyAccountID: keyAccountID,
		RobloxUserID: robloxUserID,
		RawJSON:      jsonCopy,
		UpdatedAt:    time.Now(),
	}
}

// Get retrieves a buffered inventory (for read-through).
func (b *InventoryBuffer) Get(robloxUserID string) (*BufferedInventory, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	inv, exists := b.pending[robloxUserID]
	return inv, exists
}

// Count returns the number of pending items.
func (b *InventoryBuffer) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.pending)
}

// Flush immediately flushes all pending items to the database.
func (b *InventoryBuffer) Flush(ctx context.Context) error {
	b.mu.Lock()

	if len(b.pending) == 0 {
		b.mu.Unlock()
		return nil
	}

	// Collect all pending items
	items := make([]*BufferedInventory, 0, len(b.pending))
	for _, inv := range b.pending {
		items = append(items, inv)
	}

	// Clear the pending map
	b.pending = make(map[string]*BufferedInventory)
	b.mu.Unlock()

	log.Printf("[InventoryBuffer] Flushing %d items to database", len(items))

	// Flush to database
	if err := b.flushFunc(ctx, items); err != nil {
		log.Printf("[InventoryBuffer] Flush error: %v", err)
		// Re-add failed items back to buffer
		b.mu.Lock()
		for _, inv := range items {
			// Only re-add if not already updated
			if _, exists := b.pending[inv.RobloxUserID]; !exists {
				b.pending[inv.RobloxUserID] = inv
			}
		}
		b.mu.Unlock()
		return err
	}

	log.Printf("[InventoryBuffer] Successfully flushed %d items", len(items))
	return nil
}

// backgroundFlush runs the periodic flush to database.
func (b *InventoryBuffer) backgroundFlush() {
	for {
		select {
		case <-b.flushTicker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			b.Flush(ctx)
			cancel()
		case <-b.stopFlush:
			// Final flush on shutdown
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			b.Flush(ctx)
			cancel()
			return
		}
	}
}

// Close stops the background flush and performs a final flush.
func (b *InventoryBuffer) Close() error {
	b.flushTicker.Stop()
	close(b.stopFlush)
	return nil
}
