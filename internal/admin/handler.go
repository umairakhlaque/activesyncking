package admin

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/umairakhlaque/activesyncking/internal/store"
	"github.com/umairakhlaque/activesyncking/pkg/models"
)

// Handler holds dependencies for all admin HTTP handlers.
type Handler struct {
	adminStore  *Store
	deviceStore *store.DeviceStore
	userStore   *store.UserStore
	apiKey      string
	log         *zap.Logger
}

func NewHandler(
	as *Store,
	ds *store.DeviceStore,
	us *store.UserStore,
	apiKey string,
	log *zap.Logger,
) *Handler {
	return &Handler{
		adminStore:  as,
		deviceStore: ds,
		userStore:   us,
		apiKey:      apiKey,
		log:         log,
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parsePagination(r *http.Request) (page, pageSize int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ = strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	return
}

func randToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// normDeviceStatus maps the DB value "quarantine" to the API value "quarantined".
func normDeviceStatus(s models.DeviceStatus) string {
	if s == models.DeviceStatusQuarantine {
		return "quarantined"
	}
	return string(s)
}

// ── handlers ─────────────────────────────────────────────────────────────────

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if h.adminStore == nil || h.adminStore.pool == nil {
		dbStatus = "unavailable"
	}
	apiKeyLen := len(h.apiKey)
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "ok",
		"service":     "adminsvc",
		"db":          dbStatus,
		"api_key_len": apiKeyLen,
	})
}

// Login accepts {"password":"<api_key>"} and echoes back {"token":"<api_key>"} on success.
// The "token" is used as the Bearer credential for all subsequent requests.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if h.apiKey == "" || body.Password == "" ||
		subtle.ConstantTimeCompare([]byte(body.Password), []byte(h.apiKey)) != 1 {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": h.apiKey})
}

func (h *Handler) DashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminStore.GetDashboardStats(r.Context())
	if err != nil {
		h.log.Error("dashboard stats", zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	users, total, err := h.adminStore.ListUsers(r.Context(), page, pageSize)
	if err != nil {
		h.log.Error("list users", zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	type userJSON struct {
		ID          string            `json:"id"`
		Username    string            `json:"username"`
		Email       string            `json:"email"`
		DisplayName string            `json:"displayName"`
		Status      models.UserStatus `json:"status"`
		Source      models.UserSource `json:"source"`
		MFAEnrolled bool              `json:"mfaEnrolled"`
		DeviceCount int               `json:"deviceCount"`
		LastLogin   *string           `json:"lastLogin"`
		CreatedAt   string            `json:"createdAt"`
		Groups      []string          `json:"groups"`
	}

	items := make([]userJSON, len(users))
	for i, u := range users {
		items[i] = userJSON{
			ID:          u.ID.String(),
			Username:    u.Username,
			Email:       u.Email,
			DisplayName: u.DisplayName,
			Status:      u.Status,
			Source:      u.Source,
			MFAEnrolled: u.MFAEnrolled,
			DeviceCount: u.DeviceCount,
			LastLogin:   nil,
			CreatedAt:   u.CreatedAt.Format(time.RFC3339),
			Groups:      []string{},
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid user id")
		return
	}
	u, err := h.userStore.GetByID(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var body struct {
		Status models.UserStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.adminStore.SetUserStatus(r.Context(), id, body.Status); err != nil {
		h.log.Error("set user status", zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	u, _ := h.userStore.GetByID(r.Context(), id)
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	devices, total, err := h.adminStore.ListDevices(r.Context(), page, pageSize)
	if err != nil {
		h.log.Error("list devices", zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	type deviceJSON struct {
		ID         string `json:"id"`
		DeviceID   string `json:"deviceId"`
		UserID     string `json:"userId"`
		Username   string `json:"username"`
		Type       string `json:"type"`
		OS         string `json:"os"`
		Status     string `json:"status"`
		LastSeen   string `json:"lastSeen"`
		EnrolledAt string `json:"enrolledAt"`
		UserAgent  string `json:"userAgent"`
	}

	items := make([]deviceJSON, len(devices))
	for i, d := range devices {
		enrolledAt := ""
		if d.EnrolledAt != nil {
			enrolledAt = d.EnrolledAt.Format(time.RFC3339)
		}
		items[i] = deviceJSON{
			ID:         d.ID.String(),
			DeviceID:   d.DeviceEASID,
			UserID:     d.UserID.String(),
			Username:   d.Username,
			Type:       d.DeviceType,
			OS:         d.DeviceModel,
			Status:     normDeviceStatus(d.Status),
			LastSeen:   d.UpdatedAt.Format(time.RFC3339),
			EnrolledAt: enrolledAt,
			UserAgent:  d.UserAgent,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *Handler) GetDevice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid device id")
		return
	}
	d, err := h.deviceStore.GetByID(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "device not found")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *Handler) DeviceAction(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid device id")
		return
	}
	var body struct {
		Action string `json:"action"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	switch body.Action {
	case "approve":
		err = h.deviceStore.Approve(r.Context(), id, randToken(), time.Now().Add(720*time.Hour))
	case "block":
		err = h.deviceStore.Block(r.Context(), id, body.Reason)
	case "quarantine":
		err = h.deviceStore.Quarantine(r.Context(), id, body.Reason)
	default:
		writeErr(w, http.StatusBadRequest, fmt.Sprintf("unknown action %q", body.Action))
		return
	}

	if err != nil {
		h.log.Error("device action", zap.String("action", body.Action), zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	d, _ := h.deviceStore.GetByID(r.Context(), id)
	writeJSON(w, http.StatusOK, d)
}

func (h *Handler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid device id")
		return
	}
	if err := h.adminStore.DeleteDevice(r.Context(), id); err != nil {
		h.log.Error("delete device", zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListAudit(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)

	events, total, err := h.adminStore.ListAudit(r.Context(), page, pageSize)
	if err != nil {
		h.log.Error("list audit", zap.Error(err))
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	type auditJSON struct {
		ID        string `json:"id"`
		Timestamp string `json:"timestamp"`
		Action    string `json:"action"`
		Username  string `json:"username"`
		DeviceID  string `json:"deviceId"`
		IPAddress string `json:"ipAddress"`
		Result    string `json:"result"`
		Severity  string `json:"severity"`
		Detail    string `json:"detail"`
		PolicyID  *string `json:"policyId"`
	}

	items := make([]auditJSON, len(events))
	for i, e := range events {
		items[i] = auditJSON{
			ID:        e.ID.String(),
			Timestamp: e.OccurredAt.Format("2006-01-02 15:04:05"),
			Action:    string(e.Action),
			Username:  e.Username,
			DeviceID:  e.DeviceEASID,
			IPAddress: e.ClientIP,
			Result:    e.Result,
			Severity:  e.Severity,
			Detail:    "",
			PolicyID:  nil,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ListPolicies returns an empty list — policy engine is a Phase 2 feature.
func (h *Handler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items": []any{},
		"total": 0,
	})
}
