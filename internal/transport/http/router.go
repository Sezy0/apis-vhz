package http

import (
	"net/http"

	"vinzhub-rest-api/internal/transport/http/handler"
	"vinzhub-rest-api/internal/transport/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

// NewRouter creates and configures the HTTP router.
func NewRouter(h *handler.Handler, invHandler *handler.InventoryHandler, adminHandler *handler.AdminHandler) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware stack
	r.Use(middleware.Recovery)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-API-Key"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API Key authentication (skip for health checks)
	r.Use(middleware.APIKeyAuth)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health check endpoints (no auth required)
		r.Get("/health", h.Health)
		r.Get("/ready", h.Ready)

		// Inventory endpoints
		if invHandler != nil {
			r.Route("/inventory/{roblox_user_id}", func(r chi.Router) {
				r.Post("/sync", invHandler.SyncRawInventory)
				r.Get("/", invHandler.GetRawInventory)
			})
		}

		// Admin endpoints
		if adminHandler != nil {
			r.Route("/admin", func(r chi.Router) {
				r.Get("/stats", adminHandler.GetStats)
				r.Get("/health", adminHandler.GetHealth)
			})
		}
	})

	// Static files (admin dashboard)
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Admin dashboard redirect
	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/admin.html", http.StatusMovedPermanently)
	})

	return r
}
