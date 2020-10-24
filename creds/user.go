package creds

import "github.com/google/uuid"

type User struct {
	Id       uuid.UUID
	Name     string
	Username string
	Tokens   []*Token `pg:"rel:has-many"`
}
