package models

import "time"

// PasetoSecret is the secret for the PASETO token.
type PasetoSecret struct {
	PasetoPrivateKey string
	PasetoPublicKey  string
}

// PASETOToken is the PASETO token.
type PASETOToken struct {
	UserID     uint32
	Token      string
	Expiration int64
	ExpiresAt  time.Time
}

// Claims are the claims for the PASETO token.
type Claims struct {
	UserID     uint32
	Roles      []*Role
	Token      string
	Purpose    string
	Expiration int64
	ExpiresAt  time.Time
}
