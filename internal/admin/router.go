package admin

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter wires all admin API routes with CORS and API-key auth middleware.
func NewRouter(h *Handler, apiKey string, corsOrigins []string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware(corsOrigins))

	r.Get("/healthz", h.Healthz)
	r.Post("/v1/auth/login", h.Login)

	r.Group(func(r chi.Router) {
		r.Use(apiKeyAuth(apiKey))
		r.Get("/v1/dashboard/stats", h.DashboardStats)
		r.Get("/v1/users", h.ListUsers)
		r.Get("/v1/users/{id}", h.GetUser)
		r.Patch("/v1/users/{id}", h.UpdateUser)
		r.Get("/v1/devices", h.ListDevices)
		r.Get("/v1/devices/{id}", h.GetDevice)
		r.Post("/v1/devices/{id}/action", h.DeviceAction)
		r.Delete("/v1/devices/{id}", h.DeleteDevice)
		r.Get("/v1/audit", h.ListAudit)
		r.Get("/v1/policies", h.ListPolicies)
	})

	return r
}

func apiKeyAuth(apiKey string) func(http.Handler) http.Handler {
	key := []byte(apiKey)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if len(key) == 0 || len(token) == 0 || subtle.ConstantTimeCompare([]byte(token), key) != 1 {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func corsMiddleware(origins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(origins))
	for _, o := range origins {
		allowed[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (len(allowed) == 0 || allowed[origin] || allowed["*"]) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
