package middleware

import (
	"net/http"
	"os"
	"strings"

	"vinzhub-rest-api/internal/transport/http/response"
	"vinzhub-rest-api/pkg/apierror"
)

// APIKeyAuth middleware validates API key from X-API-Key header.
func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check
		if r.URL.Path == "/api/v1/health" || r.URL.Path == "/api/v1/ready" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth for docs
		if strings.HasPrefix(r.URL.Path, "/docs") {
			next.ServeHTTP(w, r)
			return
		}

		// Get API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Also check Authorization header (Bearer token style)
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if apiKey == "" {
			response.Error(w, apierror.Unauthorized("API key required. Use X-API-Key header."))
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
