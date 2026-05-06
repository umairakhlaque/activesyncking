package auth

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter wires all auth service routes with standard middleware.
func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Cache-Control", "no-store")
			next.ServeHTTP(w, req)
		})
	})

	r.Get("/healthz", h.Healthz)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/authenticate", h.Authenticate)

		r.Route("/mfa", func(r chi.Router) {
			r.Post("/validate", h.ValidateMFA)
		})
	})

	return r
}
