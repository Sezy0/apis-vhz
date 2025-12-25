package handler

// Handler contains all HTTP handlers and their dependencies.
type Handler struct{}

// New creates a new handler.
func New(_ interface{}) *Handler {
	return &Handler{}
}
