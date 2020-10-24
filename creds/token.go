package creds

import "github.com/google/uuid"

type Token struct {
	UserId uuid.UUID
	User   *User `pg:"rel:belong"`
}
