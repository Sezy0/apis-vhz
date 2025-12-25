package domain

import (
	"time"
)

// RawInventory represents raw JSON inventory data.
type RawInventory struct {
	ID            int64      `json:"id"`
	KeyAccountID  int64      `json:"key_account_id"`
	RobloxUserID  string     `json:"roblox_user_id"`
	InventoryJSON []byte     `json:"inventory_json"`
	SyncedAt      time.Time  `json:"synced_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Common errors
var (
	ErrNotFound = &CustomError{Code: "NOT_FOUND", Message: "Resource not found"}
)

// CustomError represents a custom error.
type CustomError struct {
	Code    string
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}
