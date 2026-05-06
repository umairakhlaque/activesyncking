package email

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	gomail "gopkg.in/gomail.v2"
	"golang.org/x/crypto/bcrypt"
	"github.com/umairakhlaque/activesyncking/internal/config"
	"github.com/umairakhlaque/activesyncking/internal/mfa"
	"github.com/umairakhlaque/activesyncking/pkg/models"
)

// Provider implements mfa.Provider for email-delivered one-time passwords.
type Provider struct {
	cfg config.EmailConfig
}

func New(cfg config.EmailConfig) *Provider {
	return &Provider{cfg: cfg}
}

func (p *Provider) Method() models.ChallengeMethod { return models.ChallengeMethodEmail }

// GenerateSecret generates a 6-digit OTP, sends it to the user's email,
// and returns a bcrypt hash of the OTP to be stored in the challenge record.
func (p *Provider) GenerateSecret(_ context.Context, _ string, email string) (secret, provisionURL string, err error) {
	code, err := generateOTP(6)
	if err != nil {
		return "", "", fmt.Errorf("generating otp: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("hashing otp: %w", err)
	}

	if err := p.sendEmail(email, code); err != nil {
		return "", "", fmt.Errorf("sending otp email: %w", err)
	}

	// Return the hash as secret (stored in challenge.OTPHash)
	return string(hash), "", nil
}

// Validate checks the submitted OTP against the bcrypt hash stored in challenge.OTPHash.
func (p *Provider) Validate(_ context.Context, otp string, _ string, challenge *models.Challenge) error {
	if challenge.OTPHash == "" {
		return mfa.ErrInvalidOTP{Msg: "no OTP hash on challenge"}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(challenge.OTPHash), []byte(otp)); err != nil {
		return mfa.ErrInvalidOTP{Msg: "invalid email OTP"}
	}
	return nil
}

func (p *Provider) sendEmail(to, code string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", p.cfg.From, p.cfg.FromName)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "SyncGuard MFA Verification Code")
	m.SetBody("text/plain", buildEmailBody(code))
	m.AddAlternative("text/html", buildEmailHTML(code))

	d := gomail.NewDialer(p.cfg.SMTPHost, p.cfg.SMTPPort, p.cfg.SMTPUser, p.cfg.SMTPPassword)
	return d.DialAndSend(m)
}

func generateOTP(digits int) (string, error) {
	max := new(big.Int)
	max.Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", digits, n), nil
}

func buildEmailBody(code string) string {
	return strings.TrimSpace(fmt.Sprintf(`
Your SyncGuard MFA verification code is: %s

This code expires in 5 minutes. Do not share it with anyone.

If you did not request this code, contact your IT administrator immediately.
`, code))
}

func buildEmailHTML(code string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:480px;margin:40px auto;">
  <h2 style="color:#1a1a2e;">SyncGuard MFA</h2>
  <p>Your verification code is:</p>
  <div style="font-size:36px;font-weight:bold;letter-spacing:8px;color:#0f3460;padding:20px;background:#f0f4ff;border-radius:8px;text-align:center;">%s</div>
  <p style="color:#666;font-size:13px;">Expires in 5 minutes. Do not share this code.</p>
</body>
</html>`, code)
}
