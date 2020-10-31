package creds

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
)

type Token struct {
	Id     uuid.UUID `json:"id" pg:"type:uuid,pk"`
	Scope  string    `json:"scope" pg:",notnull"`
	UserId uuid.UUID `json:"userId" pg:"type:uuid,notnull"`
	User   *User     `json:"user" pg:"rel:has-one"`
	Start  time.Time `json:"start" pg:",notnull"`
	End    time.Time `json:"end" pg:",notnull"`
}

type NoSuchUserError struct {
	UserId uuid.UUID
}

func (noSuchUserError NoSuchUserError) Error() string {
	return fmt.Sprintf("User with Id '%s' does not exist", noSuchUserError.UserId)
}

func insertToken(database *pg.DB, id uuid.UUID, scope string, start time.Time, end time.Time) (uuid.UUID, error) {
	tokenId := uuid.New()
	if start.IsZero() {
		start = time.Now()
	}
	if end.IsZero() {
		end = time.Now().AddDate(1, 0, 0)
	}
	token := Token{
		Id:     tokenId,
		Scope:  scope,
		UserId: id,
		User:   nil,
		Start:  start,
		End:    end,
	}

	if _, err := database.Model(&token).Insert(); err != nil {
		if strings.Contains(err.Error(), "tokens_user_id_fkey") {
			return uuid.Nil, NoSuchUserError{UserId: id}
		}

		return uuid.Nil, err
	}

	return tokenId, nil
}

func getTokenById(database *pg.DB, id uuid.UUID) (*Token, error) {
	token := &Token{Id: id}

	if err := database.Model(token).WherePK().Select(); err != nil {
		return nil, err
	}

	return token, nil
}
