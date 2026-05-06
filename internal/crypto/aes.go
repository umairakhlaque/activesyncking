package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// Encryptor provides AES-256-GCM encryption/decryption.
// Used to protect TOTP secrets at rest in the database.
type Encryptor struct {
	key []byte
}

// NewEncryptor parses a 64-character hex-encoded 32-byte key.
func NewEncryptor(hexKey string) (*Encryptor, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("decoding encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, errors.New("encryption key must be 32 bytes (64 hex chars)")
	}
	return &Encryptor{key: key}, nil
}

// Encrypt returns ciphertext as nonce||ciphertext bytes.
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt unpacks nonce||ciphertext and returns plaintext.
func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting: %w", err)
	}

	return plaintext, nil
}
