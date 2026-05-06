package crypto_test

import (
	"bytes"
	"testing"

	sgcrypto "github.com/umairakhlaque/activesyncking/internal/crypto"
)

const testKey = "0000000000000000000000000000000000000000000000000000000000000001"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	enc, err := sgcrypto.NewEncryptor(testKey)
	if err != nil {
		t.Fatalf("NewEncryptor: %v", err)
	}

	plaintext := []byte("JBSWY3DPEHPK3PXP") // test TOTP secret

	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext should not equal plaintext")
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptProducesUniqueCiphertexts(t *testing.T) {
	enc, err := sgcrypto.NewEncryptor(testKey)
	if err != nil {
		t.Fatalf("NewEncryptor: %v", err)
	}

	plaintext := []byte("same plaintext")

	c1, _ := enc.Encrypt(plaintext)
	c2, _ := enc.Encrypt(plaintext)

	// Each encryption must produce a different ciphertext (random nonce)
	if bytes.Equal(c1, c2) {
		t.Fatal("two encryptions of the same plaintext produced identical ciphertexts")
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	enc, err := sgcrypto.NewEncryptor(testKey)
	if err != nil {
		t.Fatalf("NewEncryptor: %v", err)
	}

	ct, _ := enc.Encrypt([]byte("secret"))
	ct[len(ct)-1] ^= 0xFF // flip last byte

	_, err = enc.Decrypt(ct)
	if err == nil {
		t.Fatal("expected error decrypting tampered ciphertext")
	}
}

func TestInvalidKeyLength(t *testing.T) {
	_, err := sgcrypto.NewEncryptor("tooshort")
	if err == nil {
		t.Fatal("expected error for short key")
	}
}
