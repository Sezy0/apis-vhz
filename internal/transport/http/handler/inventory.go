package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"vinzhub-rest-api/internal/service"
	"vinzhub-rest-api/internal/transport/http/response"
	"vinzhub-rest-api/pkg/apierror"

	"github.com/go-chi/chi/v5"
)

// InventoryHandler handles inventory-related HTTP requests.
type InventoryHandler struct {
	inventoryService *service.InventoryService
}

// NewInventoryHandler creates a new inventory handler.
func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// SyncRawInventory handles POST /api/v1/inventory/{roblox_user_id}/sync
// Accepts any JSON and stores it raw in the database.
func (h *InventoryHandler) SyncRawInventory(w http.ResponseWriter, r *http.Request) {
	robloxUserID := chi.URLParam(r, "roblox_user_id")
	if robloxUserID == "" {
		response.Error(w, apierror.BadRequest("roblox_user_id is required"))
		return
	}

	// Read raw body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, apierror.BadRequest("failed to read request body"))
		return
	}
	defer r.Body.Close()

	// Validate it's valid JSON
	var jsonData json.RawMessage
	if err := json.Unmarshal(body, &jsonData); err != nil {
		response.Error(w, apierror.BadRequest("invalid JSON"))
		return
	}

	// Store raw JSON
	err = h.inventoryService.SyncRawInventory(r.Context(), robloxUserID, body)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, map[string]interface{}{
		"status":   "synced",
		"user_id":  robloxUserID,
		"size":     len(body),
	})
}

// GetRawInventory handles GET /api/v1/inventory/{roblox_user_id}
// Returns the raw JSON stored for this user.
func (h *InventoryHandler) GetRawInventory(w http.ResponseWriter, r *http.Request) {
	robloxUserID := chi.URLParam(r, "roblox_user_id")
	if robloxUserID == "" {
		response.Error(w, apierror.BadRequest("roblox_user_id is required"))
		return
	}

	data, syncedAt, err := h.inventoryService.GetRawInventory(r.Context(), robloxUserID)
	if err != nil {
		response.Error(w, err)
		return
	}

	// Return raw JSON as-is
	response.OK(w, map[string]interface{}{
		"roblox_user_id": robloxUserID,
		"inventory":      json.RawMessage(data),
		"synced_at":      syncedAt,
	})
}
