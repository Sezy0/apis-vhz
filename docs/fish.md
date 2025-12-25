# Fish Inventory Endpoint

## Overview

Manage fish caught by players in Fish It game.

---

## Get Fish Inventory

```
GET /api/v1/inventory/{roblox_user_id}/fish
```

### Parameters

| Name | Location | Type | Required | Description |
|------|----------|------|----------|-------------|
| roblox_user_id | path | string | Yes | Player's Roblox User ID |
| page | query | int | No | Page number (default: 1) |
| limit | query | int | No | Items per page (default: 50, max: 500) |

### Example Request

```bash
curl -X GET "https://sanbox.vinzhub.com/api/v1/inventory/123456789/fish?page=1&limit=50"
```

### Example Response

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "key_account_id": 100,
      "roblox_user_id": "123456789",
      "fish_id": 45,
      "uuid": "fish-uuid-abc123",
      "name": "Golden Carp",
      "tier": 5,
      "image_url": null,
      "variant_id": "shiny",
      "variant_name": "Shiny",
      "is_shiny": true,
      "metadata": null,
      "favorited": true,
      "sold": false,
      "sold_at": null,
      "sold_price": null,
      "caught_at": "2025-12-24T07:00:00Z",
      "synced_at": "2025-12-24T07:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 50,
    "total": 1
  }
}
```

---

## Add Fish to Inventory

```
POST /api/v1/inventory/{roblox_user_id}/fish
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| fish_id | int | Yes | Fish type ID from game |
| uuid | string | Yes | Unique item UUID from game |
| name | string | Yes | Fish name |
| tier | int | Yes | Rarity tier (1-7) |
| image_url | string | No | Fish image URL |
| variant_id | string | No | Mutation/variant ID |
| variant_name | string | No | Mutation name |
| is_shiny | bool | No | Is shiny variant |
| metadata | object | No | Additional metadata |
| favorited | bool | No | Is favorited |
| area | string | No | Fishing area where caught |
| rod_id | int | No | Rod used |
| bait_id | int | No | Bait used |

### Example Request

```bash
curl -X POST "https://sanbox.vinzhub.com/api/v1/inventory/123456789/fish" \
  -H "Content-Type: application/json" \
  -d '{
    "fish_id": 45,
    "uuid": "fish-uuid-abc123",
    "name": "Golden Carp",
    "tier": 5,
    "is_shiny": true,
    "favorited": false,
    "area": "Treasure Room",
    "rod_id": 169,
    "bait_id": 15
  }'
```

### Example Response

```json
{
  "success": true,
  "data": {
    "id": 1,
    "key_account_id": 100,
    "roblox_user_id": "123456789",
    "fish_id": 45,
    "uuid": "fish-uuid-abc123",
    "name": "Golden Carp",
    "tier": 5,
    "caught_at": "2025-12-24T07:00:00Z"
  }
}
```

---

## Fish Tier Reference

| Tier | Name | Rarity |
|------|------|--------|
| 1 | Common | Very common |
| 2 | Uncommon | Common |
| 3 | Rare | Uncommon |
| 4 | Epic | Rare |
| 5 | Legendary | Very rare |
| 6 | Mythic | Extremely rare |
| 7 | SECRET | Ultra rare |
