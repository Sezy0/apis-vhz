package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"vinzhub-rest-api/internal/service"
	"vinzhub-rest-api/internal/transport/http/response"
	"vinzhub-rest-api/pkg/apierror"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// ContextKeyTokenData is the key for storing token data in request context.
	ContextKeyTokenData ContextKey = "token_data"
)

// tokenServiceInstance is set by SetTokenService for token validation.
var tokenServiceInstance *service.TokenService

// SetTokenService sets the token service for middleware to use.
func SetTokenService(ts *service.TokenService) {
	tokenServiceInstance = ts
}

// APIKeyAuth middleware validates API key or session token.
// Supports both X-API-Key (for server-to-server) and X-Token (for client sessions).
func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check
		if r.URL.Path == "/api/v1/health" || r.URL.Path == "/api/v1/ready" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth for admin dashboard and static files
		if r.URL.Path == "/admin" || strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth for docs
		if strings.HasPrefix(r.URL.Path, "/docs") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth for token generation endpoint
		if r.URL.Path == "/api/v1/auth/token" && r.Method == "POST" {
			next.ServeHTTP(w, r)
			return
		}

		// Try X-Token first (session tokens)
		token := r.Header.Get("X-Token")
		if token != "" && tokenServiceInstance != nil {
			tokenData, err := tokenServiceInstance.ValidateToken(r.Context(), token)
			if err != nil {
				response.Error(w, apierror.Unauthorized("Invalid or expired token"))
				return
			}
			
			// Store token data in context for handlers to use
			ctx := context.WithValue(r.Context(), ContextKeyTokenData, tokenData)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Fall back to X-API-Key (server-to-server or legacy)
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Also check Authorization header (Bearer token style)
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if apiKey == "" {
			response.Error(w, apierror.Unauthorized("Authentication required. Use X-Token or X-API-Key header."))
			return
		}

		// Validate API key
		validKeys := getValidAPIKeys()
		if !isValidKey(apiKey, validKeys) {
			response.Error(w, apierror.Unauthorized("Invalid API key"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getValidAPIKeys returns list of valid API keys from environment.
func getValidAPIKeys() []string {
	// Get from environment variable (comma-separated)
	keysEnv := os.Getenv("API_KEYS")
	if keysEnv == "" {
		// Fallback to single key
		singleKey := os.Getenv("API_KEY")
		if singleKey != "" {
			return []string{singleKey}
		}
		return nil
	}

	keys := strings.Split(keysEnv, ",")
	for i := range keys {
		keys[i] = strings.TrimSpace(keys[i])
	}
	return keys
}

// isValidKey checks if the provided key is in the valid keys list.
func isValidKey(key string, validKeys []string) bool {
	for _, valid := range validKeys {
		if key == valid {
			return true
		}
	}
	return false
}

// GetTokenDataFromContext retrieves token data from request context.
func GetTokenDataFromContext(ctx context.Context) *service.TokenData {
	if data, ok := ctx.Value(ContextKeyTokenData).(*service.TokenData); ok {
		return data
	}
	return nil
}

