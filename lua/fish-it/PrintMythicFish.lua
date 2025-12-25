--[[
    Print All Mythic+ Fish Data (Tier 6 & 7)
    Displays: Name, ID, Tier, Sell Price, Image URL
    
    Usage: Run this script in your executor while in Fish It game
    Output: Check F9 Console / Dev Console
]]

local HttpService = game:GetService("HttpService")
local ReplicatedStorage = game:GetService("ReplicatedStorage")

--------------------------------------------------------------------------------
-- CONFIGURATION
--------------------------------------------------------------------------------

local TIER_NAMES = {
    [1] = "Common",
    [2] = "Uncommon", 
    [3] = "Rare",
    [4] = "Epic",
    [5] = "Legendary",
    [6] = "Mythic",
    [7] = "SECRET"
}

local TIER_EMOJI = {
    [1] = "⚪",
    [2] = "🟢", 
    [3] = "🔵",
    [4] = "🟣",
    [5] = "🟠",
    [6] = "🔴",
    [7] = "✨"
}

--------------------------------------------------------------------------------
-- ICON UTILITIES (from InventorySync)
--------------------------------------------------------------------------------

local IconCache = {}

local function GetThumbnailURL(assetString)
    if not assetString then return nil end
    
    local assetId = tostring(assetString):match("rbxassetid://(%d+)")
    if not assetId then return nil end
    
    -- Check cache first
    if IconCache[assetId] then
        return IconCache[assetId]
    end
    
    local api = string.format(
        "https://thumbnails.roblox.com/v1/assets?assetIds=%s&returnPolicy=PlaceHolder&size=420x420&format=Png",
        assetId
    )
    
    local success, response = pcall(function()
        local raw = game:HttpGet(api)
        return HttpService:JSONDecode(raw)
    end)
    
    if success and response and response.data and response.data[1] then
        local imageUrl = response.data[1].imageUrl
        IconCache[assetId] = imageUrl
        return imageUrl
    end
    
    return nil
end

--------------------------------------------------------------------------------
-- FORMAT HELPERS
--------------------------------------------------------------------------------

local function shortenNumber(n)
    if type(n) ~= "number" then return "N/A" end
    local scales = {{1e18, "Qi"}, {1e15, "Qa"}, {1e12, "T"}, {1e9, "B"}, {1e6, "M"}, {1e3, "K"}}
    local negative = n < 0
    n = math.abs(n)
    if n < 1000 then return (negative and "-" or "") .. tostring(math.floor(n)) end
    for i = 1, #scales do
        local scale, label = scales[i][1], scales[i][2]
        if n >= scale then
            local value = n / scale
            if value % 1 == 0 then return (negative and "-" or "") .. string.format("%.0f%s", value, label)
            else return (negative and "-" or "") .. string.format("%.2f%s", value, label) end
        end
    end
    return (negative and "-" or "") .. tostring(n)
end

--------------------------------------------------------------------------------
-- MAIN FUNCTION
--------------------------------------------------------------------------------

local function printMythicPlusFish()
    print("\n")
    print("========================================")
    print("🐟 LOADING MYTHIC+ FISH DATABASE...")
    print("========================================")
    
    -- Load Items module
    local success, itemsData = pcall(require, ReplicatedStorage:WaitForChild("Items"))
    if not success then 
        warn("❌ Failed to load Items module")
        return 
    end
    
    local mythicFish = {}
    local secretFish = {}
    local totalProcessed = 0
    
    print("📦 Processing fish data...")
    
    for _, item in pairs(itemsData) do
        if type(item) == "table" and item.Data and item.Data.Type == "Fish" then
            local tier = item.Data.Tier
            if tier and tier >= 6 then -- Mythic (6) and SECRET (7)
                totalProcessed = totalProcessed + 1
                
                -- Get image URL from CDN
                local iconUrl = GetThumbnailURL(item.Data.Icon)
                
                local fishInfo = {
                    Id = item.Data.Id,
                    Name = item.Data.Name,
                    Tier = tier,
                    TierName = TIER_NAMES[tier] or "Unknown",
                    Emoji = TIER_EMOJI[tier] or "❓",
                    SellPrice = item.SellPrice or 0,
                    SellPriceFormatted = shortenNumber(item.SellPrice or 0),
                    Icon = item.Data.Icon, -- rbxassetid://
                    IconURL = iconUrl, -- CDN URL
                }
                
                if tier == 6 then
                    table.insert(mythicFish, fishInfo)
                elseif tier == 7 then
                    table.insert(secretFish, fishInfo)
                end
                
                -- Show progress
                if totalProcessed % 5 == 0 then
                    print(string.format("   Processed: %d fish...", totalProcessed))
                end
            end
        end
    end
    
    -- Sort by sell price (highest first)
    table.sort(mythicFish, function(a, b) return a.SellPrice > b.SellPrice end)
    table.sort(secretFish, function(a, b) return a.SellPrice > b.SellPrice end)
    
    -- Print header
    print("\n")
    print("========================================")
    print("🐟 MYTHIC & SECRET FISH DATABASE")
    print("========================================")
    print(string.format("Total Found: %d fish (Mythic: %d, SECRET: %d)", 
        #mythicFish + #secretFish, #mythicFish, #secretFish))
    print("========================================")
    
    -- Print SECRET fish (T7)
    print("\n")
    print("╔══════════════════════════════════════════════════════════════╗")
    print("║          ✨ SECRET FISH (Tier 7) - Total: " .. string.format("%-3d", #secretFish) .. "              ║")
    print("╚══════════════════════════════════════════════════════════════╝")
    print("")
    
    for i, fish in ipairs(secretFish) do
        print(string.format("┌─────────────────────────────────────────────────────────────┐"))
        print(string.format("│ #%-3d %s %s", i, fish.Emoji, fish.Name))
        print(string.format("│      ID: %-10d | Tier: %s", fish.Id, fish.TierName))
        print(string.format("│      Sell Price: %s coins", fish.SellPriceFormatted))
        if fish.IconURL then
            print(string.format("│      Image: %s", fish.IconURL))
        else
            print(string.format("│      Asset: %s", fish.Icon or "N/A"))
        end
        print(string.format("└─────────────────────────────────────────────────────────────┘"))
        print("")
    end
    
    -- Print Mythic fish (T6)
    print("\n")
    print("╔══════════════════════════════════════════════════════════════╗")
    print("║          🔴 MYTHIC FISH (Tier 6) - Total: " .. string.format("%-3d", #mythicFish) .. "              ║")
    print("╚══════════════════════════════════════════════════════════════╝")
    print("")
    
    for i, fish in ipairs(mythicFish) do
        print(string.format("┌─────────────────────────────────────────────────────────────┐"))
        print(string.format("│ #%-3d %s %s", i, fish.Emoji, fish.Name))
        print(string.format("│      ID: %-10d | Tier: %s", fish.Id, fish.TierName))
        print(string.format("│      Sell Price: %s coins", fish.SellPriceFormatted))
        if fish.IconURL then
            print(string.format("│      Image: %s", fish.IconURL))
        else
            print(string.format("│      Asset: %s", fish.Icon or "N/A"))
        end
        print(string.format("└─────────────────────────────────────────────────────────────┘"))
        print("")
    end
    
    -- Summary
    print("========================================")
    print("📊 SUMMARY")
    print("========================================")
    print(string.format("✨ SECRET Fish: %d", #secretFish))
    print(string.format("🔴 Mythic Fish: %d", #mythicFish))
    print(string.format("📦 Total: %d fish", #mythicFish + #secretFish))
    print("========================================")
    print("✅ Done! Check F9 Console for full output")
    print("========================================")
    
    return mythicFish, secretFish
end

-- Run it
printMythicPlusFish()
