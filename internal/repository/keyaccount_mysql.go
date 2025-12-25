package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// MySQLKeyAccountRepository implements KeyAccountRepository using MySQL.
type MySQLKeyAccountRepository struct {
	db *sql.DB
}

// NewMySQLKeyAccountRepository creates a new MySQL key account repository.
func NewMySQLKeyAccountRepository(db *sql.DB) *MySQLKeyAccountRepository {
	return &MySQLKeyAccountRepository{db: db}
}

// GetKeyAccountByRobloxUser finds key_account by roblox_user_id.
func (r *MySQLKeyAccountRepository) GetKeyAccountByRobloxUser(ctx context.Context, robloxUserID string) (int64, error) {
	query := `SELECT id FROM key_accounts WHERE roblox_user_id = ? AND is_active = 1 LIMIT 1`
	
	var id int64
	err := r.db.QueryRowContext(ctx, query, robloxUserID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("key account not found for roblox user: %s", robloxUserID)
		}
		return 0, fmt.Errorf("failed to get key account: %w", err)
	}
	
	return id, nil
}

// ValidateKeyAccount checks if key_account_id exists and is active.
func (r *MySQLKeyAccountRepository) ValidateKeyAccount(ctx context.Context, keyAccountID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM key_accounts WHERE id = ? AND is_active = 1`
	
	var count int
	err := r.db.QueryRowContext(ctx, query, keyAccountID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to validate key account: %w", err)
	}
	
	return count > 0, nil
}

// UpdateLastSync updates last_inventory_sync timestamp and item count.
func (r *MySQLKeyAccountRepository) UpdateLastSync(ctx context.Context, keyAccountID int64, itemCount int) error {
	query := `
		UPDATE key_accounts 
		SET last_inventory_sync = ?, inventory_item_count = ?
		WHERE id = ?`
	
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), itemCount, keyAccountID)
	if err != nil {
		return fmt.Errorf("failed to update last sync: %w", err)
	}
	
	return nil
}

// GetKeyAccountInfo returns key account details including key and user info.
func (r *MySQLKeyAccountRepository) GetKeyAccountInfo(ctx context.Context, keyAccountID int64) (map[string]interface{}, error) {
	query := `
		SELECT 
			ka.id, ka.roblox_user_id, ka.roblox_username, ka.hwid,
			ka.is_active, ka.is_online, ka.last_heartbeat_at,
			k.key as license_key, k.status as key_status
		FROM key_accounts ka
		JOIN keys k ON ka.key_id = k.id
		WHERE ka.id = ?`
	
	var (
		id, robloxUserID, robloxUsername, hwid string
		isActive, isOnline                      bool
		lastHeartbeat                           sql.NullTime
		licenseKey, keyStatus                   string
	)
	
	err := r.db.QueryRowContext(ctx, query, keyAccountID).Scan(
		&id, &robloxUserID, &robloxUsername, &hwid,
		&isActive, &isOnline, &lastHeartbeat,
		&licenseKey, &keyStatus,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("key account not found")
		}
		return nil, err
	}
	
	result := map[string]interface{}{
		"id":              id,
		"roblox_user_id":  robloxUserID,
		"roblox_username": robloxUsername,
		"hwid":            hwid,
		"is_active":       isActive,
		"is_online":       isOnline,
		"license_key":     licenseKey,
		"key_status":      keyStatus,
	}
	
	if lastHeartbeat.Valid {
		result["last_heartbeat_at"] = lastHeartbeat.Time
	}
	
	return result, nil
}
