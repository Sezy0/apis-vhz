package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

func init() {
	// Load .env file if it exists (silent fail if not)
	_ = godotenv.Load()
}

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Server   ServerConfig
	App      AppConfig
	Cache    CacheConfig
	Database DatabaseConfig
	// Note: GameDB removed - now using SQLite for inventory storage
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host            string        `envconfig:"SERVER_HOST" default:"0.0.0.0"`
	Port            int           `envconfig:"SERVER_PORT" default:"8080"`
	ReadTimeout     time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"15s"`
	WriteTimeout    time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"15s"`
	ShutdownTimeout time.Duration `envconfig:"SERVER_SHUTDOWN_TIMEOUT" default:"30s"`
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Name        string `envconfig:"APP_NAME" default:"vinzhub-api"`
	Environment string `envconfig:"APP_ENV" default:"development"`
	Debug       bool   `envconfig:"APP_DEBUG" default:"false"`
	Version     string `envconfig:"APP_VERSION" default:"1.0.0"`
}

// CacheConfig holds cache settings.
type CacheConfig struct {
	Type string        `envconfig:"CACHE_TYPE" default:"memory"`
	TTL  time.Duration `envconfig:"CACHE_TTL" default:"5m"`

	RedisHost     string `envconfig:"REDIS_HOST" default:"localhost"`
	RedisPort     int    `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" default:""`
	RedisDB       int    `envconfig:"REDIS_DB" default:"0"`
}

// DatabaseConfig holds main database connection settings (Users/Auth - for KeyAccount lookup).
type DatabaseConfig struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     int    `envconfig:"DB_PORT" default:"3306"`
	Name     string `envconfig:"DB_NAME" default:"vinzhub"`
	User     string `envconfig:"DB_USER" default:"root"`
	Password string `envconfig:"DB_PASS" default:""`
}

// Address returns the server address in host:port format.
func (s *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// IsDevelopment returns true if running in development mode.
func (a *AppConfig) IsDevelopment() bool {
	return a.Environment == "development"
}

// IsProduction returns true if running in production mode.
func (a *AppConfig) IsProduction() bool {
	return a.Environment == "production"
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}

// MustLoad loads configuration or panics on error.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
