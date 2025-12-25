package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"vinzhub-rest-api/pkg/apierror"
)

// Recovery is a middleware that recovers from panics.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				log.Printf("PANIC: %v\n%s", err, debug.Stack())

				// Return internal server error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(apierror.InternalError("internal server error").ToJSON())
			}
		}()

		next.ServeHTTP(w, r)
	})
}
