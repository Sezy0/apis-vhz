# Player Progress Endpoint

## Overview

Track and update player progress including coins, quests, and unlocks.

---

## Get Player Progress

```
GET /api/v1/inventory/{roblox_user_id}/progress
```

### Example Request

```bash
curl -X GET "https://sanbox.vinzhub.com/api/v1/inventory/123456789/progress"
```

### Example Response

```json
{
  "success": true,
  "data": {
    "id": 1,
    "key_account_id": 100,
    "roblox_user_id": "123456789",
    "coins": 1500000,
    "total_fish_caught": 5000,
    "total_fish_sold": 4500,
    "total_coins_earned": 10000000,
    "ghostfin_quest": {
      "rare_epic": 300,
      "mythic": 3,
      "secret": 1,
      "coins": 1000000
    },
    "element_quest": {
      "ghostfin_rod": 1,
      "secret_jungle": 1,
      "secret_temple": 1,
      "transcended_stones": 3
    },
    "unlocked_ruins": true,
    "unlocked_iron_cavern": false,
    "temple_levers": {
      "Arrow Artifact": true,
      "Crescent Artifact": true,
      "Diamond Artifact": true,
      "Hourglass Diamond Artifact": false
    },
    "equipped_rod_id": 169,
    "equipped_bait_id": 15,
    "last_area": "Treasure Room",
    "last_played_at": "2025-12-24T07:00:00Z",
    "synced_at": "2025-12-24T07:00:00Z"
  }
}
```

---

## Update Player Progress

```
PUT /api/v1/inventory/{roblox_user_id}/progress
```

### Request Body

| Field | Type | Description |
|-------|------|-------------|
| coins | int | Current coins |
| total_fish_caught | int | Total fish caught |
| total_fish_sold | int | Total fish sold |
| total_coins_earned | int | Lifetime coins earned |
| ghostfin_quest | object | Ghostfin quest progress |
| element_quest | object | Element rod quest progress |
| deep_sea_quest | object | Deep sea quest progress |
| unlocked_ruins | bool | Ruins unlocked status |
| unlocked_iron_cavern | bool | Iron Cavern unlocked status |
| temple_levers | object | Temple lever status |
| equipped_rod_id | int | Currently equipped rod |
| equipped_bait_id | int | Currently equipped bait |
| last_area | string | Last fishing area |
| last_position | object | Last CFrame position |

### Example Request

```bash
curl -X PUT "https://sanbox.vinzhub.com/api/v1/inventory/123456789/progress" \
  -H "Content-Type: application/json" \
  -d '{
    "coins": 2000000,
    "total_fish_caught": 5500,
    "unlocked_ruins": true,
    "temple_levers": {
      "Arrow Artifact": true,
      "Crescent Artifact": true,
      "Diamond Artifact": true,
      "Hourglass Diamond Artifact": true
    },
    "equipped_rod_id": 257,
    "last_area": "Ancient Ruin"
  }'
```

### Example Response

```json
{
  "success": true,
  "data": {
    "status": "updated"
  }
}
```

---

## Quest Progress Structure

### Ghostfin Quest

```json
{
  "rare_epic": 300,      // Catch 300 Rare/Epic fish
  "mythic": 3,           // Catch 3 Mythic fish
  "secret": 1,           // Catch 1 SECRET fish
  "coins": 1000000       // Earn 1M coins
}
```

### Element Rod Quest

```json
{
  "ghostfin_rod": 1,         // Own Ghostfin Rod
  "secret_jungle": 1,        // Catch SECRET in Ancient Jungle
  "secret_temple": 1,        // Catch SECRET in Sacred Temple
  "transcended_stones": 3    // Create 3 Transcended Stones
}
```

### Temple Levers

```json
{
  "Arrow Artifact": true,
  "Crescent Artifact": true,
  "Diamond Artifact": true,
  "Hourglass Diamond Artifact": true
}
```
