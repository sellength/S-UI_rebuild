package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"s-ui/config"
	"strings"
)

const encryptedSecretPrefix = "sui:v1:"

func encryptSecret(plain []byte) (string, error) {
	gcm, err := secretGCM()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate secret nonce: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, plain, nil)
	return encryptedSecretPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

func decryptSecret(ciphertext string) ([]byte, error) {
	ciphertext = strings.TrimSpace(ciphertext)
	if !isEncryptedSecret(ciphertext) {
		return nil, fmt.Errorf("unsupported encrypted secret format")
	}
	gcm, err := secretGCM()
	if err != nil {
		return nil, err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ciphertext, encryptedSecretPrefix))
	if err != nil {
		return nil, fmt.Errorf("decode encrypted secret: %w", err)
	}
	if len(raw) < gcm.NonceSize() {
		return nil, fmt.Errorf("encrypted secret is too short")
	}
	nonce, sealed := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt encrypted secret: %w", err)
	}
	return plain, nil
}

func isEncryptedSecret(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), encryptedSecretPrefix)
}

func secretGCM() (cipher.AEAD, error) {
	secret := config.GetSecretKey()
	if len(secret) < 32 {
		return nil, fmt.Errorf("SUI_SECRET_KEY must be set to at least 32 characters")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("create secret cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create secret gcm: %w", err)
	}
	return gcm, nil
}
