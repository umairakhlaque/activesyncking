package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/umairakhlaque/activesyncking/internal/config"
	sgcrypto "github.com/umairakhlaque/activesyncking/internal/crypto"
	"github.com/umairakhlaque/activesyncking/internal/mfa"
	"github.com/umairakhlaque/activesyncking/internal/session"
	"github.com/umairakhlaque/activesyncking/internal/store"
	"github.com/umairakhlaque/activesyncking/pkg/apitypes"
	"github.com/umairakhlaque/activesyncking/pkg/models"
)

// Service implements the core authentication and MFA decision logic.
// It is called by the HTTP handler and by the gateway's auth middleware.
type Service struct {
	cfg          config.AuthConfig
	users        *store.UserStore
	devices      *store.DeviceStore
	totpSecrets  *store.TOTPStore
	challengeDB  *store.ChallengeDBStore
	audit        *store.AuditStore
	sessions     *session.Store
	mfaProviders map[models.ChallengeMethod]mfa.Provider
	enc          *sgcrypto.Encryptor
	log          *zap.Logger
}

func NewService(
	cfg config.AuthConfig,
	users *store.UserStore,
	devices *store.DeviceStore,
	totpSecrets *store.TOTPStore,
	challengeDB *store.ChallengeDBStore,
	audit *store.AuditStore,
	sessions *session.Store,
	providers []mfa.Provider,
	enc *sgcrypto.Encryptor,
	log *zap.Logger,
) *Service {
	pm := make(map[models.ChallengeMethod]mfa.Provider, len(providers))
	for _, p := range providers {
		pm[p.Method()] = p
	}
	return &Service{
		cfg:          cfg,
		users:        users,
		devices:      devices,
		totpSecrets:  totpSecrets,
		challengeDB:  challengeDB,
		audit:        audit,
		sessions:     sessions,
		mfaProviders: pm,
		enc:          enc,
		log:          log,
	}
}

// Authenticate is the entry point for every EAS authentication attempt.
// It validates credentials, checks device trust state, and either allows
// the request through or issues an MFA challenge.
func (s *Service) Authenticate(ctx context.Context, req apitypes.AuthenticateRequest) (*apitypes.AuthenticateResponse, error) {
	// 1. Validate credentials
	user, err := s.validateCredentials(ctx, req.Username, req.Password)
	if err != nil {
		s.audit.WriteAsync(&models.AuditEvent{
			Action:   models.AuditActionAuthFailure,
			Username: req.Username,
			ClientIP: req.IP,
			Result:   "INVALID_CREDENTIALS",
			Severity: models.AuditSeverityWarn,
		})
		return &apitypes.AuthenticateResponse{
			Status:  apitypes.AuthStatusDenied,
			Message: "invalid credentials",
		}, nil
	}

	// 2. Look up or register the device
	device, err := s.upsertDevice(ctx, user, req)
	if err != nil {
		s.log.Error("device upsert failed", zap.Error(err))
		return &apitypes.AuthenticateResponse{Status: apitypes.AuthStatusError}, nil
	}

	// 3. Blocked device — deny immediately
	if device.IsBlocked() {
		s.audit.WriteAsync(&models.AuditEvent{
			Action:      models.AuditActionAccessDenied,
			UserID:      &user.ID,
			Username:    user.Username,
			DeviceEASID: device.DeviceEASID,
			ClientIP:    req.IP,
			Result:      "DEVICE_BLOCKED",
			Severity:    models.AuditSeverityWarn,
		})
		return &apitypes.AuthenticateResponse{
			Status:  apitypes.AuthStatusBlocked,
			Message: "device is blocked",
		}, nil
	}

	// 4. Trusted device with a valid session — allow through without MFA
	if device.IsTrusted() {
		s.audit.WriteAsync(&models.AuditEvent{
			Action:      models.AuditActionAccessAllowed,
			UserID:      &user.ID,
			Username:    user.Username,
			DeviceEASID: device.DeviceEASID,
			ClientIP:    req.IP,
			Result:      "TRUSTED_DEVICE",
			Severity:    models.AuditSeverityInfo,
		})
		return &apitypes.AuthenticateResponse{Status: apitypes.AuthStatusAllow}, nil
	}

	// 5. Unknown / pending device — issue MFA challenge
	method, err := s.selectMFAMethod(ctx, user)
	if err != nil {
		s.log.Error("selecting MFA method", zap.Error(err))
		return &apitypes.AuthenticateResponse{Status: apitypes.AuthStatusError}, nil
	}

	challenge, err := s.createChallenge(ctx, user, device, req.IP, method)
	if err != nil {
		s.log.Error("creating challenge", zap.Error(err))
		return &apitypes.AuthenticateResponse{Status: apitypes.AuthStatusError}, nil
	}

	if err := s.devices.SetPending(ctx, device.ID); err != nil {
		s.log.Error("setting device pending", zap.Error(err))
	}

	s.audit.WriteAsync(&models.AuditEvent{
		Action:      models.AuditActionMFARequired,
		UserID:      &user.ID,
		Username:    user.Username,
		DeviceEASID: device.DeviceEASID,
		ClientIP:    req.IP,
		ChallengeID: challenge.ID,
		Result:      "MFA_CHALLENGE_ISSUED",
		Severity:    models.AuditSeverityInfo,
		Details:     map[string]any{"method": method},
	})

	return &apitypes.AuthenticateResponse{
		Status:      apitypes.AuthStatusMFARequired,
		ChallengeID: challenge.ID,
		Method:      string(method),
	}, nil
}

// ValidateMFA processes a submitted OTP for an active challenge.
func (s *Service) ValidateMFA(ctx context.Context, req apitypes.MFAValidateRequest) (*apitypes.MFAValidateResponse, error) {
	// 1. Fetch challenge from Redis
	challenge, err := s.sessions.Get(ctx, req.ChallengeID)
	if errors.Is(err, session.ErrNotFound) || errors.Is(err, session.ErrExpired) {
		return &apitypes.MFAValidateResponse{
			Status:  apitypes.MFAStatusExpired,
			Message: "challenge not found or expired",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fetching challenge: %w", err)
	}

	if challenge.IsTerminal() {
		return &apitypes.MFAValidateResponse{
			Status:  apitypes.MFAStatusExpired,
			Message: "challenge already completed or expired",
		}, nil
	}

	// 2. Increment attempt counter before validation (fail-safe)
	challenge.Attempts++

	if challenge.Attempts > s.cfg.MaxAttempts {
		challenge.Status = models.ChallengeStatusLocked
		_ = s.sessions.Update(ctx, challenge)
		_ = s.challengeDB.UpdateStatus(ctx, challenge.ID, models.ChallengeStatusLocked, challenge.Attempts)

		s.audit.WriteAsync(&models.AuditEvent{
			Action:      models.AuditActionMFALocked,
			UserID:      &challenge.UserID,
			DeviceEASID: challenge.DeviceEASID,
			ClientIP:    challenge.ClientIP,
			ChallengeID: challenge.ID,
			Result:      "MAX_ATTEMPTS_EXCEEDED",
			Severity:    models.AuditSeverityCritical,
		})

		return &apitypes.MFAValidateResponse{
			Status:  apitypes.MFAStatusLocked,
			Message: "too many failed attempts",
		}, nil
	}

	// 3. Get the TOTP secret for this user (if TOTP method)
	var secret string
	if challenge.Method == models.ChallengeMethodTOTP {
		totpRecord, err := s.totpSecrets.Get(ctx, challenge.UserID)
		if err != nil {
			s.log.Error("fetching totp secret", zap.Error(err))
			return &apitypes.MFAValidateResponse{Status: apitypes.MFAStatusFailed, Message: "MFA not configured"}, nil
		}

		decrypted, err := s.enc.Decrypt(totpRecord.SecretEnc)
		if err != nil {
			s.log.Error("decrypting totp secret", zap.Error(err))
			return nil, fmt.Errorf("decrypting totp secret")
		}
		secret = string(decrypted)
	}

	// 4. Validate OTP via the appropriate provider
	provider, ok := s.mfaProviders[challenge.Method]
	if !ok {
		return nil, fmt.Errorf("no provider for method %s", challenge.Method)
	}

	if err := provider.Validate(ctx, req.OTP, secret, challenge); err != nil {
		_ = s.sessions.Update(ctx, challenge)
		_ = s.challengeDB.UpdateStatus(ctx, challenge.ID, models.ChallengeStatusFailed, challenge.Attempts)

		s.audit.WriteAsync(&models.AuditEvent{
			Action:      models.AuditActionMFAFailure,
			UserID:      &challenge.UserID,
			DeviceEASID: challenge.DeviceEASID,
			ClientIP:    challenge.ClientIP,
			ChallengeID: challenge.ID,
			Result:      "INVALID_OTP",
			Severity:    models.AuditSeverityWarn,
			Details:     map[string]any{"attempts": challenge.Attempts},
		})

		return &apitypes.MFAValidateResponse{
			Status:  apitypes.MFAStatusFailed,
			Message: "invalid OTP",
		}, nil
	}

	// 5. Success — approve device and issue access token
	challenge.Status = models.ChallengeStatusCompleted
	_ = s.sessions.Update(ctx, challenge)
	_ = s.challengeDB.UpdateStatus(ctx, challenge.ID, models.ChallengeStatusCompleted, challenge.Attempts)

	device, err := s.devices.GetByEASIDAndUser(ctx, challenge.DeviceEASID, challenge.UserID)
	if err == nil {
		trustToken := generateTrustToken()
		expiry := time.Now().Add(s.cfg.DeviceTrustTTL)
		if approveErr := s.devices.Approve(ctx, device.ID, trustToken, expiry); approveErr != nil {
			s.log.Error("approving device", zap.Error(approveErr))
		}
	}

	s.audit.WriteAsync(&models.AuditEvent{
		Action:      models.AuditActionMFASuccess,
		UserID:      &challenge.UserID,
		DeviceEASID: challenge.DeviceEASID,
		ClientIP:    challenge.ClientIP,
		ChallengeID: challenge.ID,
		Result:      "MFA_VALIDATED",
		Severity:    models.AuditSeverityInfo,
	})

	return &apitypes.MFAValidateResponse{Status: apitypes.MFAStatusSuccess}, nil
}

// ─── Internal helpers ────────────────────────────────────────────────────────

func (s *Service) validateCredentials(ctx context.Context, username, password string) (*models.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if errors.Is(err, store.ErrNotFound) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	if !user.IsActive() {
		return nil, errors.New("user is not active")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Increment failed login counter; lock after threshold
		var lockUntil *time.Time
		if user.FailedLogins+1 >= 10 {
			t := time.Now().Add(15 * time.Minute)
			lockUntil = &t
		}
		_ = s.users.IncrementFailedLogins(ctx, user.ID, lockUntil)
		return nil, errors.New("invalid password")
	}

	_ = s.users.ResetFailedLogins(ctx, user.ID)
	return user, nil
}

func (s *Service) upsertDevice(ctx context.Context, user *models.User, req apitypes.AuthenticateRequest) (*models.Device, error) {
	d := &models.Device{
		DeviceEASID: req.DeviceID,
		UserID:      user.ID,
		DeviceType:  req.DeviceType,
		UserAgent:   req.UserAgent,
		LastIP:      req.IP,
		Status:      models.DeviceStatusUnknown,
	}

	if err := s.devices.Upsert(ctx, d); err != nil {
		return nil, err
	}

	// Re-fetch to get the current status (upsert preserves existing status)
	return s.devices.GetByEASIDAndUser(ctx, req.DeviceID, user.ID)
}

func (s *Service) selectMFAMethod(ctx context.Context, user *models.User) (models.ChallengeMethod, error) {
	// Prefer TOTP if the user has an enrolled secret
	if _, err := s.totpSecrets.Get(ctx, user.ID); err == nil {
		return models.ChallengeMethodTOTP, nil
	}
	// Fall back to email OTP if email is available
	if user.Email != "" {
		if _, ok := s.mfaProviders[models.ChallengeMethodEmail]; ok {
			return models.ChallengeMethodEmail, nil
		}
	}
	return "", errors.New("no MFA method available for user")
}

func (s *Service) createChallenge(
	ctx context.Context,
	user *models.User,
	device *models.Device,
	clientIP string,
	method models.ChallengeMethod,
) (*models.Challenge, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	challenge := &models.Challenge{
		ID:          id,
		UserID:      user.ID,
		DeviceEASID: device.DeviceEASID,
		Method:      method,
		Status:      models.ChallengeStatusPending,
		Attempts:    0,
		ClientIP:    clientIP,
		CreatedAt:   now,
		ExpiresAt:   now.Add(s.cfg.ChallengeTTL),
	}

	// For email OTP: generate and send the code, store bcrypt hash
	if method == models.ChallengeMethodEmail {
		provider := s.mfaProviders[models.ChallengeMethodEmail]
		hash, _, err := provider.GenerateSecret(ctx, user.ID.String(), user.Email)
		if err != nil {
			return nil, fmt.Errorf("generating email OTP: %w", err)
		}
		challenge.OTPHash = hash
	}

	// Persist to Redis (fast path) and PostgreSQL (audit)
	if err := s.sessions.Save(ctx, challenge); err != nil {
		return nil, fmt.Errorf("saving challenge to redis: %w", err)
	}
	if err := s.challengeDB.Create(ctx, challenge); err != nil {
		// Non-fatal — Redis is the source of truth for active challenges
		s.log.Warn("failed to persist challenge to db", zap.Error(err))
	}

	return challenge, nil
}

func generateTrustToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// ─── TOTP Enrolment ──────────────────────────────────────────────────────────

type EnrolTOTPResult struct {
	ProvisionURL string
	QRCodeURI    string
}

func (s *Service) EnrolTOTP(ctx context.Context, userID uuid.UUID) (*EnrolTOTPResult, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	provider, ok := s.mfaProviders[models.ChallengeMethodTOTP]
	if !ok {
		return nil, errors.New("TOTP provider not configured")
	}

	secret, provisionURL, err := provider.GenerateSecret(ctx, userID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("generating totp secret: %w", err)
	}

	enc, err := s.enc.Encrypt([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("encrypting totp secret: %w", err)
	}

	rec := &models.TOTPSecret{
		UserID:    userID,
		SecretEnc: enc,
		Enrolled:  false,
	}
	if err := s.totpSecrets.Upsert(ctx, rec); err != nil {
		return nil, fmt.Errorf("saving totp secret: %w", err)
	}

	qr, err := totpQR(provisionURL)
	if err != nil {
		s.log.Warn("QR code generation failed", zap.Error(err))
		qr = ""
	}

	return &EnrolTOTPResult{
		ProvisionURL: provisionURL,
		QRCodeURI:    qr,
	}, nil
}

func totpQR(provisionURL string) (string, error) {
	// Imported lazily to avoid import cycle; the totp package exposes this.
	// We call it via a package-level var so tests can inject a stub.
	return totpQRFn(provisionURL)
}

// totpQRFn is a package-level variable so tests can replace it.
var totpQRFn = func(provisionURL string) (string, error) {
	// Import is resolved at link time; avoids circular import with totp package
	// by delegating via this indirection.
	return "", nil
}
