package creds

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/guregu/null.v4"
)

type addUserParameters struct {
	Username    null.String
	Name        null.String
	AdminUserId null.String
	AdminToken  uuid.UUID
}

type addUserParametersError struct {
	Username    bool
	Name        bool
	AdminUserId bool
	AdminToken  bool
}

func (aup addUserParametersError) Error() string {
	errors := make([]string, 0)

	if aup.Username {
		errors = append(errors, "'username' missing")
	}

	if aup.Name {
		errors = append(errors, "'name' missing")
	}

	if aup.AdminUserId {
		errors = append(errors, "'adminUserId' missing")
	}

	if aup.AdminToken {
		errors = append(errors, "'adminToken' missing")
	}

	return strings.Join(errors, ", ")
}

func (a *addUserParameters) UnmarshalJSON(bytes []byte) error {
	var s struct {
		Username    null.String
		Name        null.String
		AdminUserId null.String
		AdminToken  uuid.UUID
	}
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	a.Username = s.Username
	a.Name = s.Name
	a.AdminUserId = s.AdminUserId
	a.AdminToken = s.AdminToken

	if !a.Username.Valid || !a.Name.Valid || !a.AdminUserId.Valid || a.AdminToken.ID() == 0 {
		return addUserParametersError{
			Username:    !a.Username.Valid,
			Name:        !a.Name.Valid,
			AdminUserId: !a.AdminUserId.Valid,
			AdminToken:  a.AdminToken.ID() == 0,
		}
	}

	return nil
}

func handleAddUser(db *pg.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parameters := addUserParameters{}
		if err := json.NewDecoder(r.Body).Decode(&parameters); err != nil {
			_, _ = fmt.Fprintf(w, "Error decoding parameters for adding user: %s", err.Error())

			return
		}

		u := User{
			Id:       uuid.New(),
			Name:     parameters.Name.String,
			Username: parameters.Username.String,
			Tokens:   nil,
		}

		context := db.Context()
		if err := db.RunInTransaction(context, func(tx *pg.Tx) error {
			adminTokens := make([]Token, 0)
			adminTokenExists, err := db.Model(&adminTokens).Where(
				"user_id = ? AND scope = 'severnatazvezda.com'", parameters.AdminUserId, parameters.AdminToken.String(),
			).Exists()
			if err != nil {
				return err
			}

			if !adminTokenExists {
				return fmt.Errorf("user does not have privileges for scope 'severnatazvezda.com'")
			}

			_, err = db.Model(&u).Insert()
			if err != nil {
				fmt.Printf("Error inserting user: %s", err.Error())

				return err
			}

			return nil
		}); err != nil {
			_, _ = fmt.Fprintf(w, "Error running transaction: %s", err.Error())
		}

		_ = json.NewEncoder(w).Encode(u)
	}
}

func handleGetUser(db *pg.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := new(uuid.UUID)
		parameters := getParameters(r)
		if parameters == nil {
			_, _ = fmt.Fprint(w, "No `UserId` given as path parameter")

			return
		}

		if err := id.Scan(r.Context().Value(httprouter.ParamsKey).(httprouter.Params).ByName("Id")); err != nil {
			_, _ = fmt.Fprintf(w, "Unable to get `UserId` from parameter: %s", err.Error())

			return
		}

		u := User{}
		if err := db.Model(&u).Where("id = ?", id).Relation("Tokens").Select(); err != nil {
			_, _ = fmt.Fprintf(w, "Error getting user: %s", err.Error())

			return
		}

		_ = json.NewEncoder(w).Encode(u)
	}
}

func handleGetUsers(db *pg.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users := make([]User, 0)
		if err := db.Model(&users).Relation("Tokens").Select(); err != nil {
			_, _ = fmt.Fprintf(w, "Error getting users: %s", err.Error())

			return
		}

		if err := json.NewEncoder(w).Encode(users); err != nil {
			fmt.Printf("Unable to write user list to socket: %s", err.Error())
		}
	}
}

func handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, "hello!")
		if err != nil {
			fmt.Printf("Unable to send hello response: %s", err.Error())
		}
	}
}
