package creds

import (
	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
)

type User struct {
	Id       uuid.UUID `json:"id" pg:"type:uuid"`
	Name     string    `json:"name" pg:",notnull"`
	Username string    `json:"username" pg:",notnull,unique"`
	Tokens   []*Token  `json:"tokens" pg:"rel:has-many"`
}

func insertUser(database *pg.DB, name string, username string) (uuid.UUID, error) {
	id := uuid.New()
	user := User{
		Id:       id,
		Name:     name,
		Username: username,
		Tokens:   nil,
	}

	if _, err := database.Model(&user).Insert(); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func getUserById(database *pg.DB, id uuid.UUID) (*User, error) {
	user := &User{Id: id}

	if err := database.Model(user).WherePK().Select(); err != nil {
		return nil, err
	}

	return user, nil
}
