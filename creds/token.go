package creds

import (
	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
)

type Token struct {
	Id     uuid.UUID `json:"id" pg:"type:uuid,pk"`
	Scope  string    `json:"scope" pg:",pk"`
	UserId uuid.UUID `json:"userId" pg:"type:uuid,notnull"`
	User   *User     `json:"user" pg:"rel:has-one"`
}

func insertToken(database *pg.DB, userId uuid.UUID, scope string) (uuid.UUID, error) {
	id := uuid.New()
	token := Token{
		Id:     id,
		Scope:  scope,
		UserId: userId,
		User:   nil,
	}

	if _, err := database.Model(&token).Insert(); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
