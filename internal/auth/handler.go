package auth

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/umairakhlaque/activesyncking/pkg/apitypes"
)

// Handler exposes the auth service over HTTP.
type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// POST /api/v1/authenticate
//
// Request:  AuthenticateRequest  (JSON body)
// Response: AuthenticateResponse (JSON body)
//
// Called by the EAS gateway's auth middleware before every proxied request.
// The gateway submits credentials + device context; this handler decides
// ALLOW, MFA_REQUIRED, DENIED, or BLOCKED.
func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req apitypes.AuthenticateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "username and password are required")
		return
	}
	if req.DeviceID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_DEVICE_ID", "device_id is required")
		return
	}

	// Populate IP from X-Forwarded-For if not set by caller
	if req.IP == "" {
		req.IP = realIP(r)
	}

	resp, err := h.svc.Authenticate(r.Context(), req)
	if err != nil {
		h.log.Error("authenticate error", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "authentication service error")
		return
	}

	// Fail-closed: treat error status as deny
	if resp.Status == apitypes.AuthStatusError {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_ERROR", "authentication service unavailable")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// POST /api/v1/mfa/validate
//
// Request:  MFAValidateRequest  (JSON body)
// Response: MFAValidateResponse (JSON body)
//
// Submits a TOTP code or email OTP against an active challenge.
func (h *Handler) ValidateMFA(w http.ResponseWriter, r *http.Request) {
	var req apitypes.MFAValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "malformed request body")
		return
	}

	if req.ChallengeID == "" || req.OTP == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "challenge_id and otp are required")
		return
	}

	resp, err := h.svc.ValidateMFA(r.Context(), req)
	if err != nil {
		h.log.Error("validate mfa error", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "MFA validation error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GET /healthz
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, apitypes.ErrorResponse{Code: code, Message: msg})
}

func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
