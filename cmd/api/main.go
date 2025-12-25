package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"vinzhub-rest-api/internal/cache"
	"vinzhub-rest-api/internal/config"
	"vinzhub-rest-api/internal/repository"
	"vinzhub-rest-api/internal/service"
	httpTransport "vinzhub-rest-api/internal/transport/http"
	"vinzhub-rest-api/internal/transport/http/handler"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// OPTIMIZATION FOR SHARED HOSTING (Low Resource Limits)
	// Limit to 1 CPU core to reduce thread usage
	runtime.GOMAXPROCS(1)

	// Load configuration
	cfg := config.MustLoad()

	log.Printf("Starting %s v%s in %s mode",
		cfg.App.Name,
		cfg.App.Version,
		cfg.App.Environment,
	)

	// Initialize infrastructure layer
	memoryCache := cache.NewMemoryCache()
	defer memoryCache.Close()

	// Connect to Main Database (for key_accounts lookup - optional)
	mainDB, err := connectDB(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		"Main DB",
	)
	if err != nil {
		log.Printf("Warning: Failed to connect to Main DB: %v", err)
		mainDB = nil
	} else {
		defer mainDB.Close()
		log.Println("✓ Main DB connected")
	}

	// Create data directory for SQLite
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize SQLite for inventory (LOCAL - no network latency!)
	sqliteRepo, err := repository.NewSQLiteInventoryRepository("./data/inventory.db")
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize SQLite: %v", err)
	}
	defer sqliteRepo.Close()
	log.Println("✓ SQLite database initialized (./data/inventory.db)")

	// KeyAccount repo is optional (uses Main MySQL DB)
	var keyAccountRepo repository.KeyAccountRepository
	if mainDB != nil {
		keyAccountRepo = repository.NewMySQLKeyAccountRepository(mainDB)
	}

	// Initialize Redis buffer (Redis buffers writes, SQLite persists)
	// This buffers sync requests and batch-flushes to SQLite every 30 seconds
	var redisBuffer *cache.RedisInventoryBuffer
	
	flushFunc := func(ctx context.Context, items []*cache.BufferedInventory) error {
		// Convert to repository items
		repoItems := make([]repository.InventoryItem, len(items))
		for i, item := range items {
			repoItems[i] = repository.InventoryItem{
				KeyAccountID: item.KeyAccountID,
				RobloxUserID: item.RobloxUserID,
				RawJSON:      item.RawJSON,
				SyncedAt:     item.UpdatedAt,
			}
		}
		return sqliteRepo.BatchUpsertRawInventory(ctx, repoItems)
	}

	redisCfg := cache.RedisBufferConfig{
		Addr:          "127.0.0.1:6379",
		Password:      "",
		DB:            1,
		FlushInterval: 30 * time.Second,
		KeyPrefix:     "vinzhub:fishit:inventory",
	}

	var redisErr error
	redisBuffer, redisErr = cache.NewRedisInventoryBuffer(redisCfg, flushFunc)
	if redisErr != nil {
		log.Printf("⚠ Redis unavailable: %v (using direct SQLite writes)", redisErr)
		// Redis is optional for development - production should have Redis
	} else {
		defer redisBuffer.Close()
		log.Println("✓ Redis buffer enabled (flush every 30s, DB=1)")
	}

	// Initialize service - with or without Redis buffer
	var inventoryService *service.InventoryService
	if redisBuffer != nil {
		inventoryService = service.NewInventoryServiceWithBuffer(sqliteRepo, keyAccountRepo, redisBuffer)
		log.Println("✓ InventoryService initialized (Redis → SQLite)")
	} else {
		inventoryService = service.NewInventoryService(sqliteRepo, keyAccountRepo)
		log.Println("✓ InventoryService initialized (direct SQLite - no Redis)")
	}
	if inventoryService == nil {
		log.Fatalf("FATAL: Failed to create InventoryService")
	}

	// Initialize transport layer - HTTP
	httpHandler := handler.New(nil)

	var invHandler *handler.InventoryHandler
	if inventoryService != nil {
		invHandler = handler.NewInventoryHandler(inventoryService)
	}

	// Admin handler for stats dashboard
	adminHandler := handler.NewAdminHandler(redisBuffer, sqliteRepo)

	router := httpTransport.NewRouter(httpHandler, invHandler, adminHandler)

	// Configure HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("HTTP server listening on %s", cfg.Server.Address())
		log.Println("Available endpoints:")
		log.Println("  GET  /api/v1/health")
		log.Println("  POST /api/v1/inventory/{roblox_user_id}/sync")
		log.Println("  GET  /api/v1/inventory/{roblox_user_id}")
		log.Println("  GET  /api/v1/admin/stats")
		log.Println("  GET  /admin  (Dashboard UI)")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

// connectDB establishes a connection to a MySQL database.
func connectDB(host string, port int, user, password, dbName, label string) (*sql.DB, error) {
	// DSN with timeout settings to prevent hanging connections
	// timeout: connection timeout, readTimeout/writeTimeout: query timeouts
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&timeout=5s&readTimeout=10s&writeTimeout=10s",
		user, password, host, port, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", label, err)
	}

	// Configure connection pool - balanced for shared hosting
	// Increased from 3 to handle burst traffic while staying within hosting limits
	db.SetMaxOpenConns(10)              // Allow more concurrent connections
	db.SetMaxIdleConns(5)               // Keep some ready for quick reuse
	db.SetConnMaxLifetime(3 * time.Minute) // Recycle connections before they go stale
	db.SetConnMaxIdleTime(1 * time.Minute) // Close idle connections faster

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping %s: %w", label, err)
	}

	return db, nil
}

// init sets up logging format
func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}
