package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func NewOpaqueToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func NewUUID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate uuid: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(value[0:4]), hex.EncodeToString(value[4:6]), hex.EncodeToString(value[6:8]), hex.EncodeToString(value[8:10]), hex.EncodeToString(value[10:16])), nil
}
