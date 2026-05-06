package apitypes

// ─── POST /api/v1/authenticate ───────────────────────────────────────────────

type AuthenticateRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	DeviceID  string `json:"device_id"`   // EAS DeviceId header value
	DeviceType string `json:"device_type"` // iPhone, Android, etc.
	UserAgent string `json:"user_agent"`
	IP        string `json:"ip"`
}

type AuthStatus string

const (
	AuthStatusAllow       AuthStatus = "ALLOW"        // trusted device, pass through
	AuthStatusMFARequired AuthStatus = "MFA_REQUIRED" // challenge issued
	AuthStatusDenied      AuthStatus = "DENIED"        // bad credentials
	AuthStatusBlocked     AuthStatus = "BLOCKED"       // device/user blocked
	AuthStatusError       AuthStatus = "ERROR"         // internal error (fail-closed)
)

type AuthenticateResponse struct {
	Status      AuthStatus `json:"status"`
	ChallengeID string     `json:"challenge_id,omitempty"`
	Method      string     `json:"method,omitempty"` // totp, email
	Message     string     `json:"message,omitempty"`
}

// ─── POST /api/v1/mfa/validate ───────────────────────────────────────────────

type MFAValidateRequest struct {
	ChallengeID string `json:"challenge_id"`
	OTP         string `json:"otp"`
}

type MFAStatus string

const (
	MFAStatusSuccess MFAStatus = "SUCCESS"
	MFAStatusFailed  MFAStatus = "FAILED"
	MFAStatusExpired MFAStatus = "EXPIRED"
	MFAStatusLocked  MFAStatus = "LOCKED"
)

type MFAValidateResponse struct {
	Status      MFAStatus `json:"status"`
	AccessToken string    `json:"access_token,omitempty"` // short-lived JWT for gateway
	Message     string    `json:"message,omitempty"`
}

// ─── POST /api/v1/totp/enrol ─────────────────────────────────────────────────

type TOTPEnrolResponse struct {
	Secret      string `json:"secret"`       // base32, shown once
	ProvisionURL string `json:"provision_url"` // otpauth:// URI for QR code
	QRCodeURL   string `json:"qr_code_url"`  // data URI PNG
}

type TOTPVerifyRequest struct {
	OTP string `json:"otp"`
}

// ─── Error ───────────────────────────────────────────────────────────────────

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
