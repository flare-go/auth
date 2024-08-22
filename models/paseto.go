package models

import "time"

type PASETOToken struct {
	Token     string
	ExpiresAt time.Time
}

type Claims struct {
	UserID    uint32
	Roles     []*Role
	Token     string
	Purpose   string
	ExpiresAt time.Time
}
