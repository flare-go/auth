package token

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"goflare.io/auth/models"
)

// Manager defines methods for token management.
type Manager interface {

	// GenerateToken generates a new token.
	GenerateToken(userID uint32) (*models.PASETOToken, error)

	// ValidateToken validates a token.
	ValidateToken(token string) (uint32, error)

	// RevokeToken revokes a token.
	RevokeToken(token string) error
}

// PasetoManager implements Manager using PASETO.
type PasetoManager struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	expiration time.Duration
}

// NewPasetoManager creates a new instance of PasetoManager.
func NewPasetoManager(publicKey, privateKey string, expiration time.Duration) *PasetoManager {
	pubKey, _ := base64.StdEncoding.DecodeString(publicKey)
	privKey, _ := base64.StdEncoding.DecodeString(privateKey)
	return &PasetoManager{
		publicKey:  pubKey,
		privateKey: privKey,
		expiration: expiration,
	}
}

// GenerateToken generates a new PASETO token.
func (tm *PasetoManager) GenerateToken(userID uint32) (*models.PASETOToken, error) {
	now := time.Now()
	exp := now.Add(tm.expiration)
	token, err := paseto.NewV2().Sign(tm.privateKey, models.PASETOToken{
		UserID:    userID,
		ExpiresAt: exp,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	return &models.PASETOToken{Token: token, ExpiresAt: exp}, nil
}

// ValidateToken validates a PASETO token.
func (tm *PasetoManager) ValidateToken(token string) (uint32, error) {
	var tokenData models.PASETOToken
	err := paseto.NewV2().Verify(token, tm.publicKey, &tokenData, nil)
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}
	if time.Now().After(tokenData.ExpiresAt) {
		return 0, fmt.Errorf("token expired")
	}
	return tokenData.UserID, nil
}

// RevokeToken revokes a PASETO token.
func (tm *PasetoManager) RevokeToken(token string) error {
	// In a real implementation, you would add the token to a blacklist or revocation list
	// For simplicity, we'll just return nil
	return nil
}
