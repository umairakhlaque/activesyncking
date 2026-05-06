package totp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/umairakhlaque/activesyncking/internal/mfa"
	"github.com/umairakhlaque/activesyncking/pkg/models"
)

const (
	issuer  = "SyncGuard MFA"
	digits  = otp.DigitsSix
	algo    = otp.AlgorithmSHA1 // RFC 6238 default; widest authenticator app support
	skew    = 1                  // ±1 window (±30s) drift tolerance
)

// Provider implements mfa.Provider for RFC 6238 TOTP.
type Provider struct{}

func New() *Provider { return &Provider{} }

func (p *Provider) Method() models.ChallengeMethod { return models.ChallengeMethodTOTP }

// GenerateSecret creates a new TOTP key for a user and returns:
//   - the base32-encoded secret (store encrypted)
//   - the otpauth:// provisioning URL for QR code generation
//   - a data-URI PNG of the QR code (convenience for admin portal enrolment flow)
func (p *Provider) GenerateSecret(_ context.Context, _ string, email string) (secret, provisionURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
		Digits:      digits,
		Algorithm:   algo,
	})
	if err != nil {
		return "", "", fmt.Errorf("generating totp key: %w", err)
	}
	return key.Secret(), key.URL(), nil
}

// QRCodeDataURI converts a provisioning URL into a base64-encoded PNG data URI.
func QRCodeDataURI(provisionURL string) (string, error) {
	key, err := otp.NewKeyFromURL(provisionURL)
	if err != nil {
		return "", fmt.Errorf("parsing provision url: %w", err)
	}

	img, err := key.Image(256, 256)
	if err != nil {
		return "", fmt.Errorf("generating qr image: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("encoding qr png: %w", err)
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// Validate verifies a TOTP code against the base32 secret.
// secret is the plaintext (decrypted) base32 TOTP secret.
func (p *Provider) Validate(_ context.Context, otp string, secret string, _ *models.Challenge) error {
	valid, err := totp.ValidateCustom(otp, secret, time.Now().UTC(), totp.ValidateOpts{
		Digits:    digits,
		Algorithm: algo,
		Skew:      skew,
	})
	if err != nil {
		return fmt.Errorf("totp validation error: %w", err)
	}
	if !valid {
		return mfa.ErrInvalidOTP{Msg: "invalid TOTP code"}
	}
	return nil
}
