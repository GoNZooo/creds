package creds

import "github.com/google/uuid"

type Token struct {
	Id     uuid.UUID `json:"id" pg:"type:uuid,pk"`
	Scope  string    `json:"scope" pg:",pk"`
	UserId uuid.UUID `json:"userId" pg:"type:uuid,notnull"`
	User   *User     `json:"user" pg:"rel:has-one"`
}
