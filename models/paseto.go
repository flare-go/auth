package models

import "time"

type PasetoSecret struct {
	PasetoPrivateKey string
	PasetoPublicKey  string
}

type PASETOToken struct {
	UserID     uint32
	Token      string
	Expiration int64
	ExpiresAt  time.Time
}

type Claims struct {
	UserID     uint32
	Roles      []*Role
	Token      string
	Purpose    string
	Expiration int64
	ExpiresAt  time.Time
}
