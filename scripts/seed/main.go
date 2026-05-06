// seed creates a development admin user with a TOTP secret for local testing.
// Usage: go run ./scripts/seed/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/umairakhlaque/activesyncking/internal/config"
	sgcrypto "github.com/umairakhlaque/activesyncking/internal/crypto"
	mfatotp "github.com/umairakhlaque/activesyncking/internal/mfa/totp"
	"github.com/umairakhlaque/activesyncking/internal/store"
	"github.com/umairakhlaque/activesyncking/pkg/models"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	ctx := context.Background()
	pool, err := store.Open(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}
	defer pool.Close()

	enc, err := sgcrypto.NewEncryptor(cfg.Encryption.Key)
	if err != nil {
		log.Fatalf("initialising encryptor: %v", err)
	}

	users := store.NewUserStore(pool)
	totpStore := store.NewTOTPStore(pool)

	// Create dev admin user
	password := envOrDefault("SEED_PASSWORD", "admin123!")
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hashing password: %v", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     envOrDefault("SEED_USERNAME", "admin"),
		Email:        envOrDefault("SEED_EMAIL", "admin@syncguard.local"),
		DisplayName:  "SyncGuard Admin",
		Status:       models.UserStatusActive,
		Source:       models.UserSourceLocal,
		PasswordHash: string(hash),
	}

	if err := users.Create(ctx, user); err != nil {
		log.Fatalf("creating user: %v", err)
	}
	fmt.Printf("Created user: %s (id=%s)\n", user.Username, user.ID)

	// Generate TOTP secret
	provider := mfatotp.New()
	secret, provisionURL, err := provider.GenerateSecret(ctx, user.ID.String(), user.Email)
	if err != nil {
		log.Fatalf("generating totp secret: %v", err)
	}

	encSecret, err := enc.Encrypt([]byte(secret))
	if err != nil {
		log.Fatalf("encrypting totp secret: %v", err)
	}

	rec := &models.TOTPSecret{
		UserID:    user.ID,
		SecretEnc: encSecret,
		Enrolled:  true,
	}
	if err := totpStore.Upsert(ctx, rec); err != nil {
		log.Fatalf("saving totp secret: %v", err)
	}

	fmt.Printf("TOTP secret: %s\n", secret)
	fmt.Printf("Provisioning URL: %s\n", provisionURL)
	fmt.Println("\nScan the provisioning URL into your authenticator app.")
	fmt.Printf("Then test: POST http://localhost:8080/api/v1/authenticate\n")
	fmt.Printf(`  {"username":%q,"password":%q,"device_id":"DEV001"}`+"\n", user.Username, password)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
