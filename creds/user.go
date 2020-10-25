package creds

import "github.com/google/uuid"

type User struct {
	Id       uuid.UUID `json:"id" pg:"type:uuid"`
	Name     string    `json:"name" pg:",notnull"`
	Username string    `json:"username" pg:",notnull,unique"`
	Tokens   []*Token  `json:"tokens" pg:"rel:has-many"`
}
