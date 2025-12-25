package service

import (
	"context"
	"time"

	"vinzhub-rest-api/internal/cache"
	"vinzhub-rest-api/internal/repository"
)

// InventoryService handles inventory business logic.
type InventoryService struct {
	inventoryRepo  repository.InventoryRepository
	keyAccountRepo repository.KeyAccountRepository
	buffer         *cache.RedisInventoryBuffer
}

// NewInventoryService creates a new inventory service.
// Returns nil if inventoryRepo is nil (required dependency).
// DEPRECATED: Use NewInventoryServiceWithBuffer for Redis-first mode.
func NewInventoryService(
	inventoryRepo repository.InventoryRepository,
	keyAccountRepo repository.KeyAccountRepository,
) *InventoryService {
	if inventoryRepo == nil {
		return nil // Cannot function without inventory repository
	}
	return &InventoryService{
		inventoryRepo:  inventoryRepo,
		keyAccountRepo: keyAccountRepo, // Optional, can be nil
	}
}

// NewInventoryServiceWithBuffer creates a new inventory service with Redis buffer.
// Redis buffer is REQUIRED. inventoryRepo can be nil (Redis-only mode).
func NewInventoryServiceWithBuffer(
	inventoryRepo repository.InventoryRepository,
	keyAccountRepo repository.KeyAccountRepository,
	buffer *cache.RedisInventoryBuffer,
) *InventoryService {
	if buffer == nil {
		return nil // Redis buffer is required for high-traffic
	}
	return &InventoryService{
		inventoryRepo:  inventoryRepo, // Can be nil - flush will skip
		keyAccountRepo: keyAccountRepo,
		buffer:         buffer,
	}
}

// SetBuffer sets the Redis buffer for write-behind caching.
func (s *InventoryService) SetBuffer(buffer *cache.RedisInventoryBuffer) {
	s.buffer = buffer
}

// SyncRawInventory stores raw JSON inventory data.
// If buffer is set, writes to Redis first (fast), otherwise direct to DB.
// Safe to call even if keyAccountRepo is nil.
func (s *InventoryService) SyncRawInventory(ctx context.Context, robloxUserID string, rawJSON []byte) error {
	// Get key account ID (optional - can be 0 if not linked or repo unavailable)
	var keyAccountID int64
	if s.keyAccountRepo != nil {
		keyAccountID, _ = s.keyAccountRepo.GetKeyAccountByRobloxUser(ctx, robloxUserID)
	}
	
	// If buffer is available, use write-behind caching
	if s.buffer != nil {
		return s.buffer.Add(ctx, keyAccountID, robloxUserID, rawJSON)
	}
	
	// Fallback to direct DB write
	return s.inventoryRepo.UpsertRawInventory(ctx, keyAccountID, robloxUserID, rawJSON)
}

// GetRawInventory retrieves raw JSON inventory data.
// Checks Redis buffer first, then falls back to database.
func (s *InventoryService) GetRawInventory(ctx context.Context, robloxUserID string) ([]byte, *time.Time, error) {
	// Check buffer first
	if s.buffer != nil {
		if inv, err := s.buffer.Get(ctx, robloxUserID); err == nil && inv != nil {
			return inv.RawJSON, &inv.UpdatedAt, nil
		}
	}
	
	// Fall back to database
	return s.inventoryRepo.GetRawInventory(ctx, robloxUserID)
}
