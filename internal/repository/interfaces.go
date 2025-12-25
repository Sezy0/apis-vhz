package repository

import (
	"context"
	"time"
)

// InventoryRepository defines inventory data access methods.
type InventoryRepository interface {
	// Raw JSON storage
	UpsertRawInventory(ctx context.Context, keyAccountID int64, robloxUserID string, rawJSON []byte) error
	GetRawInventory(ctx context.Context, robloxUserID string) ([]byte, *time.Time, error)
}

// KeyAccountRepository defines key account data access methods.
type KeyAccountRepository interface {
	GetKeyAccountByRobloxUser(ctx context.Context, robloxUserID string) (int64, error)
}
