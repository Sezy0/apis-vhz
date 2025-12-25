# VinzHub Fish It Inventory API

> REST API Documentation for Fish It Game Inventory System

**Base URL:** `https://sandbox.vinzhub.com/api/v1`  
**Version:** 1.0.0

---

## Authentication

All inventory endpoints require a valid `roblox_user_id` linked to an active `key_account`.

---

## Endpoints

### Health Check

#### `GET /health`

Check API health status.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2025-12-24T07:00:00Z",
    "version": "1.0.0"
  }
}
```

---

### Fish Inventory

#### `GET /inventory/{roblox_user_id}/fish`

Get fish inventory for a player.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| roblox_user_id | string | Yes | Roblox User ID |
| page | int | No | Page number (default: 1) |
| limit | int | No | Items per page (default: 50, max: 500) |

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "fish_id": 123,
      "uuid": "abc-123-xyz",
      "name": "Golden Carp",
      "tier": 5,
      "variant_name": "Shiny",
      "is_shiny": true,
      "favorited": true,
      "caught_at": "2025-12-24T07:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 50,
    "total": 100
  }
}
```

---

#### `POST /inventory/{roblox_user_id}/fish`

Add fish to inventory.

**Request Body:**
```json
{
  "fish_id": 123,
  "uuid": "abc-123-xyz",
  "name": "Golden Carp",
  "tier": 5,
  "variant_id": "shiny",
  "variant_name": "Shiny",
  "is_shiny": true,
  "favorited": false,
  "area": "Treasure Room",
  "rod_id": 169,
  "bait_id": 15
}
```

---

### Rod Inventory

#### `GET /inventory/{roblox_user_id}/rods`

Get owned rods.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "rod_id": 169,
      "uuid": "rod-uuid-123",
      "name": "Ghostfin Rod",
      "equipped": true,
      "obtained_at": "2025-12-24T07:00:00Z"
    }
  ]
}
```

---

### Bait Inventory

#### `GET /inventory/{roblox_user_id}/baits`

Get owned baits.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "bait_id": 15,
      "name": "Corrupt Bait",
      "quantity": 50,
      "equipped": true
    }
  ]
}
```

---

### Player Progress

#### `GET /inventory/{roblox_user_id}/progress`

Get player progress data.

**Response:**
```json
{
  "success": true,
  "data": {
    "coins": 1500000,
    "total_fish_caught": 5000,
    "total_fish_sold": 4500,
    "unlocked_ruins": true,
    "unlocked_iron_cavern": false,
    "equipped_rod_id": 169,
    "equipped_bait_id": 15,
    "last_area": "Treasure Room"
  }
}
```

---

#### `PUT /inventory/{roblox_user_id}/progress`

Update player progress.

**Request Body:**
```json
{
  "coins": 1500000,
  "total_fish_caught": 5000,
  "unlocked_ruins": true,
  "temple_levers": {
    "Arrow Artifact": true,
    "Crescent Artifact": true
  }
}
```

---

### Full Sync

#### `POST /inventory/{roblox_user_id}/sync`

Sync complete inventory from game client.

**Request Body:**
```json
{
  "fish": [...],
  "rods": [...],
  "baits": [...],
  "progress": {...}
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "synced",
    "fish_count": 50,
    "rods_count": 5,
    "baits_count": 10
  }
}
```

---

### Summary

#### `GET /inventory/{roblox_user_id}/summary`

Get inventory summary.

**Response:**
```json
{
  "success": true,
  "data": {
    "total_fish": 150,
    "total_rods": 8,
    "total_baits": 12,
    "coins": 1500000,
    "fish_by_tier": {
      "1": 50,
      "2": 40,
      "3": 30,
      "4": 20,
      "5": 8,
      "6": 2
    }
  }
}
```

---

### Catch Logs

#### `GET /inventory/{roblox_user_id}/logs`

Get catch history logs.

| Parameter | Type | Description |
|-----------|------|-------------|
| page | int | Page number |
| limit | int | Items per page (max: 500) |

---

## Tier Reference

| Tier | Name | Color |
|------|------|-------|
| 1 | Common | Gray |
| 2 | Uncommon | Green |
| 3 | Rare | Blue |
| 4 | Epic | Purple |
| 5 | Legendary | Orange |
| 6 | Mythic | Red |
| 7 | SECRET | Pink |

---

## Error Responses

```json
{
  "success": false,
  "error": {
    "code": 404,
    "message": "key account not found for this roblox user"
  }
}
```

| Code | Description |
|------|-------------|
| 400 | Bad Request - Invalid input |
| 404 | Not Found - Resource not found |
| 500 | Internal Server Error |

---

## WebSocket

Connect to `/api/v1/ws` for real-time updates.

```javascript
const ws = new WebSocket('wss://sandbox.vinzhub.com/api/v1/ws');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Update:', data);
};
```
