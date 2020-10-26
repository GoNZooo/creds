package creds

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"
)

type addTokenParameters struct {
	UserId     uuid.UUID
	Scope      null.String
	AdminToken uuid.UUID
}

type addTokenParametersError struct {
	UserId     bool
	Scope      bool
	AdminToken bool
}

func (parametersError addTokenParametersError) Error() string {
	errors := make([]string, 0)

	if parametersError.UserId {
		errors = append(errors, "'userId' missing")
	}

	if parametersError.Scope {
		errors = append(errors, "'scope' missing")
	}

	if parametersError.AdminToken {
		errors = append(errors, "'adminToken' missing")
	}

	return strings.Join(errors, ", ")
}

func (parameters *addTokenParameters) UnmarshalJSON(bytes []byte) error {
	var s struct {
		UserId      uuid.UUID
		Scope       null.String
		AdminUserId uuid.UUID
		AdminToken  uuid.UUID
	}
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	parameters.UserId = s.UserId
	parameters.AdminToken = s.AdminToken
	parameters.Scope = s.Scope

	if parameters.UserId.ID() == 0 || !parameters.Scope.Valid || parameters.AdminToken.ID() == 0 {
		return addTokenParametersError{
			UserId:     parameters.UserId.ID() == 0,
			Scope:      !parameters.Scope.Valid,
			AdminToken: parameters.AdminToken.ID() == 0,
		}
	}

	return nil
}

func handleAddToken(database *pg.DB, adminScope string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var parameters addTokenParameters
		if err := json.NewDecoder(request.Body).Decode(&parameters); err != nil {
			_, _ = fmt.Fprintf(writer, "Error decoding parameters for adding token: %s", err.Error())

			return
		}

		adminToken := Token{}
		if err := database.Model(&adminToken).Where("id = ?", parameters.AdminToken).Select(); err != nil {
			_, _ = fmt.Fprintf(writer, "Unable to get admin token: %s", err.Error())

			return
		}

		if adminToken.Scope != adminScope {
			_, _ = fmt.Fprint(writer, "Insufficient privileges for adding tokens")

			return
		}

		token := Token{
			Id:     uuid.New(),
			Scope:  parameters.Scope.String,
			UserId: parameters.UserId,
			User:   nil,
		}

		if _, err := database.Model(&token).Insert(); err != nil {
			_, _ = fmt.Fprintf(writer, "Unable to create token: %s", err.Error())

			return
		}
	}
}

type addUserParameters struct {
	Username   null.String
	Name       null.String
	AdminToken uuid.UUID
}

type addUserParametersError struct {
	Username   bool
	Name       bool
	AdminToken bool
}

func (parametersError addUserParametersError) Error() string {
	errors := make([]string, 0)

	if parametersError.Username {
		errors = append(errors, "'username' missing")
	}

	if parametersError.Name {
		errors = append(errors, "'name' missing")
	}

	if parametersError.AdminToken {
		errors = append(errors, "'adminToken' missing")
	}

	return strings.Join(errors, ", ")
}

func (parameters *addUserParameters) UnmarshalJSON(bytes []byte) error {
	var s struct {
		Username   null.String
		Name       null.String
		AdminToken uuid.UUID
	}
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	parameters.Username = s.Username
	parameters.Name = s.Name
	parameters.AdminToken = s.AdminToken

	if !parameters.Username.Valid || !parameters.Name.Valid || parameters.AdminToken.ID() == 0 {
		return addUserParametersError{
			Username:   !parameters.Username.Valid,
			Name:       !parameters.Name.Valid,
			AdminToken: parameters.AdminToken.ID() == 0,
		}
	}

	return nil
}

func handleAddUser(database *pg.DB, adminScope string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		parameters := addUserParameters{}
		if err := json.NewDecoder(request.Body).Decode(&parameters); err != nil {
			_, _ = fmt.Fprintf(writer, "Error decoding parameters for adding user: %s", err.Error())

			return
		}

		user := User{
			Id:       uuid.New(),
			Name:     parameters.Name.String,
			Username: parameters.Username.String,
			Tokens:   nil,
		}

		context := database.Context()
		if err := database.RunInTransaction(context, func(_ *pg.Tx) error {
			adminTokens := make([]Token, 0)
			adminTokenExists, err := database.Model(&adminTokens).Where(
				"id = ? AND scope = '?'", parameters.AdminToken, adminScope,
			).Exists()
			if err != nil {
				return err
			}

			if !adminTokenExists {
				return fmt.Errorf("user does not have privileges for scope '%s'", adminScope)
			}

			_, err = database.Model(&user).Insert()
			if err != nil {
				fmt.Printf("Error inserting user: %s", err.Error())

				return err
			}

			return nil
		}); err != nil {
			_, _ = fmt.Fprintf(writer, "Error running transaction: %s", err.Error())
		}

		_ = json.NewEncoder(writer).Encode(user)
	}
}

func handleGetUser(database *pg.DB, adminScope string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		adminToken := getAdminToken(request)
		hasAdminScope := tokenHasScope(database, adminToken, adminScope)
		if !hasAdminScope {
			_, _ = fmt.Fprintf(writer, "Incorrect or no authorization token given for this resource: %s", adminToken)

			return
		}

		id := new(uuid.UUID)
		parameters := getParameters(request)
		if parameters == nil {
			_, _ = fmt.Fprint(writer, "No `UserId` given as path parameter")

			return
		}

		if err := id.Scan(parameters.ByName("Id")); err != nil {
			_, _ = fmt.Fprintf(writer, "Unable to get `Id` from parameter: %s", err.Error())

			return
		}

		user := User{}
		if err := database.Model(&user).Where("id = ?", id).Relation("Tokens").Select(); err != nil {
			_, _ = fmt.Fprintf(writer, "Error getting user: %s", err.Error())

			return
		}

		_ = json.NewEncoder(writer).Encode(user)
	}
}

func handleGetUsers(database *pg.DB, adminScope string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		adminToken := getAdminToken(request)
		hasAdminScope := tokenHasScope(database, adminToken, adminScope)
		if !hasAdminScope {
			_, _ = fmt.Fprintf(writer, "Incorrect or no authorization token given for this resource: %s", adminToken)

			return
		}

		users := make([]User, 0)
		if err := database.Model(&users).Relation("Tokens").Select(); err != nil {
			_, _ = fmt.Fprintf(writer, "Error getting users: %s", err.Error())

			return
		}

		if err := json.NewEncoder(writer).Encode(users); err != nil {
			fmt.Printf("Unable to write user list to socket: %s", err.Error())
		}
	}
}

func tokenHasScope(database *pg.DB, token string, scope string) bool {
	tokenAsUuid := uuid.UUID{}
	if err := tokenAsUuid.Scan(token); err != nil {
		return false
	}

	exists, err := database.Model((*Token)(nil)).Where("id = ? AND scope = ?", tokenAsUuid, scope).Exists()
	if err != nil {
		return false
	}

	return exists
}

func getAdminToken(request *http.Request) string {
	authorizationHeader := request.Header.Get("Authorization")
	if authorizationHeader == "" || !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return ""
	} else {
		return strings.Split(authorizationHeader, " ")[1]
	}
}
