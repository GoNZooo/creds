package creds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"
)

type addTokenParameters struct {
	UserId uuid.UUID
	Scope  null.String
}

type addTokenParametersError struct {
	UserId bool
	Scope  bool
}

func (parametersError addTokenParametersError) Error() string {
	errors := make([]string, 0)

	if parametersError.UserId {
		errors = append(errors, "'userId' missing")
	}

	if parametersError.Scope {
		errors = append(errors, "'scope' missing")
	}

	return strings.Join(errors, ", ")
}

func (parameters *addTokenParameters) UnmarshalJSON(bytes []byte) error {
	var toUnmarshal struct {
		UserId uuid.UUID
		Scope  null.String
	}
	if err := json.Unmarshal(bytes, &toUnmarshal); err != nil {
		return err
	}

	parameters.UserId = toUnmarshal.UserId
	parameters.Scope = toUnmarshal.Scope

	if parameters.UserId.ID() == 0 || !parameters.Scope.Valid {
		return addTokenParametersError{
			UserId: parameters.UserId.ID() == 0,
			Scope:  !parameters.Scope.Valid,
		}
	}

	return nil
}

func handleAddToken(database *pg.DB, adminScope string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		adminToken := getAdminTokenId(request)
		hasAdminScope := tokenHasScope(database, adminToken, adminScope)
		if !hasAdminScope {
			response := fmt.Sprintf("Incorrect or no authorization token given for this resource: %s", adminToken)
			http.Error(writer, response, http.StatusUnauthorized)

			return
		}

		var parameters addTokenParameters
		if err := json.NewDecoder(request.Body).Decode(&parameters); err != nil {
			response := fmt.Sprintf("Error decoding parameters for adding token: %s", err.Error())
			http.Error(writer, response, http.StatusBadRequest)

			return
		}

		tokenId, err := insertToken(database, parameters.UserId, parameters.Scope.String)
		if err != nil {
			response := fmt.Sprintf("Unable to create token: %s", err.Error())
			http.Error(writer, response, http.StatusBadRequest)

			return
		}

		if err := json.NewEncoder(writer).Encode(tokenId); err != nil {
			fmt.Printf("Couldn't write token '%s' for request", tokenId)
		}
	}
}

type addUserParameters struct {
	Username null.String
	Name     null.String
}

type addUserParametersError struct {
	Username bool
	Name     bool
}

func (parametersError addUserParametersError) Error() string {
	errors := make([]string, 0)

	if parametersError.Username {
		errors = append(errors, "'username' missing")
	}

	if parametersError.Name {
		errors = append(errors, "'name' missing")
	}

	return strings.Join(errors, ", ")
}

func (parameters *addUserParameters) UnmarshalJSON(bytes []byte) error {
	var toUnmarshal struct {
		Username   null.String
		Name       null.String
		AdminToken uuid.UUID
	}
	if err := json.Unmarshal(bytes, &toUnmarshal); err != nil {
		return err
	}

	parameters.Username = toUnmarshal.Username
	parameters.Name = toUnmarshal.Name

	if !parameters.Username.Valid || !parameters.Name.Valid {
		return addUserParametersError{
			Username: !parameters.Username.Valid,
			Name:     !parameters.Name.Valid,
		}
	}

	return nil
}

func handleAddUser(database *pg.DB, adminScope string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		adminToken := getAdminTokenId(request)
		hasAdminScope := tokenHasScope(database, adminToken, adminScope)
		if !hasAdminScope {
			response := fmt.Sprintf("Incorrect or no authorization token given for this resource: %s", adminToken)
			http.Error(writer, response, http.StatusUnauthorized)

			return
		}

		parameters := addUserParameters{}
		if err := json.NewDecoder(request.Body).Decode(&parameters); err != nil {
			response := fmt.Sprintf("Error decoding parameters for adding user: %s", err.Error())
			http.Error(writer, response, http.StatusBadRequest)

			return
		}

		userId, err := insertUser(database, parameters.Name.String, parameters.Username.String)
		if err != nil {
			response := fmt.Sprintf("Error inserting user: %s", err.Error())
			http.Error(writer, response, http.StatusInternalServerError)

			return
		}

		_ = json.NewEncoder(writer).Encode(userId)
	}
}

func handleDeleteUser(database *pg.DB, adminScope string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		adminToken := getAdminTokenId(request)
		hasAdminScope := tokenHasScope(database, adminToken, adminScope)
		if !hasAdminScope {
			response := fmt.Sprintf("Incorrect or no authorization token given for this resource: %s", adminToken)
			http.Error(writer, response, http.StatusUnauthorized)

			return
		}

		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			response := fmt.Sprintf("Unable to read body: %s", err.Error())
			http.Error(writer, response, http.StatusBadRequest)

			return
		}

		id, err := uuid.ParseBytes(bodyBytes)
		if err != nil {
			response := fmt.Sprintf("Unable to decode parameter as ID: %s", err.Error())
			http.Error(writer, response, http.StatusBadRequest)

			return
		}

		context := database.Context()
		if err := database.RunInTransaction(context, func(transaction *pg.Tx) error {
			tokens := make([]Token, 0)
			if _, err := database.Model(&tokens).Where("user_id = ?", id).Delete(); err != nil {
				return err
			}

			user := User{Id: id}
			if _, err := database.Model(&user).WherePK().Delete(); err != nil {
				return err
			}

			return nil
		}); err != nil {
			response := fmt.Sprintf("Unable to delete user: %s", err.Error())
			http.Error(writer, response, http.StatusInternalServerError)

			return
		}
	}
}

func handleGetUser(database *pg.DB, adminScope string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		adminToken := getAdminTokenId(request)
		hasAdminScope := tokenHasScope(database, adminToken, adminScope)
		if !hasAdminScope {
			response := fmt.Sprintf("Incorrect or no authorization token given for this resource: %s", adminToken)
			http.Error(writer, response, http.StatusUnauthorized)

			return
		}

		id := new(uuid.UUID)
		parameters := getParameters(request)
		if parameters == nil {
			response := "No `UserId` given as path parameter"
			http.Error(writer, response, http.StatusBadRequest)

			return
		}

		if err := id.Scan(parameters.ByName("Id")); err != nil {
			response := fmt.Sprintf("Unable to get `Id` from parameter: %s", err.Error())
			http.Error(writer, response, http.StatusBadRequest)

			return
		}

		users := make([]User, 0)
		if err := database.Model(&users).Where("id = ?", id).Relation("Tokens").Select(); err != nil {
			response := fmt.Sprintf("Error getting user: %s", err.Error())
			http.Error(writer, response, http.StatusInternalServerError)

			return
		}

		if len(users) == 0 {
			response := fmt.Sprintf("User with id '%s' not found", id)
			http.Error(writer, response, http.StatusNotFound)

			return
		}

		_ = json.NewEncoder(writer).Encode(users[0])
	}
}

func handleGetUsers(database *pg.DB, adminScope string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		adminToken := getAdminTokenId(request)
		hasAdminScope := tokenHasScope(database, adminToken, adminScope)
		if !hasAdminScope {
			response := fmt.Sprintf("Incorrect or no authorization token given for this resource: %s", adminToken)
			http.Error(writer, response, http.StatusUnauthorized)

			return
		}

		users := make([]User, 0)
		if err := database.Model(&users).Relation("Tokens").Select(); err != nil {
			response := fmt.Sprintf("Error getting users")
			http.Error(writer, response, http.StatusInternalServerError)

			return
		}

		if err := json.NewEncoder(writer).Encode(users); err != nil {
			fmt.Printf("Unable to write user list to socket: %s", err.Error())
		}
	}
}

func handleGetTokens(database *pg.DB, adminScope string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		adminToken := getAdminTokenId(request)
		hasAdminScope := tokenHasScope(database, adminToken, adminScope)
		if !hasAdminScope {
			response := fmt.Sprintf("Incorrect or no authorization token given for this resource: %s", adminToken)
			http.Error(writer, response, http.StatusUnauthorized)

			return
		}

		tokens := make([]Token, 0)
		if err := database.Model(&tokens).Select(); err != nil {
			response := fmt.Sprintf("Error getting tokens")
			http.Error(writer, response, http.StatusInternalServerError)

			return
		}

		if err := json.NewEncoder(writer).Encode(tokens); err != nil {
			fmt.Printf("Unable to write user list to socket: %s", err.Error())
		}
	}
}

func handleDeleteToken(database *pg.DB, adminScope string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		adminToken := getAdminTokenId(request)
		hasAdminScope := tokenHasScope(database, adminToken, adminScope)
		if !hasAdminScope {
			response := fmt.Sprintf("Incorrect or no authorization token given for this resource: %s", adminToken)
			http.Error(writer, response, http.StatusUnauthorized)

			return
		}

		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			response := fmt.Sprintf("Unable to read body: %s", err.Error())
			http.Error(writer, response, http.StatusBadRequest)

			return
		}

		id, err := uuid.ParseBytes(bodyBytes)
		if err != nil {
			response := fmt.Sprintf("Unable to decode parameter as ID: %s", err.Error())
			http.Error(writer, response, http.StatusBadRequest)

			return
		}

		token := Token{Id: id}
		if _, err := database.Model(&token).WherePK().Delete(); err != nil {
			response := fmt.Sprintf("Unable to delete token: %s", err.Error())
			http.Error(writer, response, http.StatusInternalServerError)

			return
		}
	}
}

func tokenHasScope(database *pg.DB, tokenId uuid.UUID, scope string) bool {
	exists, err := database.Model((*Token)(nil)).Where("id = ? AND scope = ?", tokenId, scope).Exists()
	if err != nil {
		return false
	}

	return exists
}

func getAdminTokenId(request *http.Request) uuid.UUID {
	authorizationHeader := request.Header.Get("Authorization")
	if authorizationHeader == "" || !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return uuid.Nil
	} else {
		id := uuid.UUID{}
		if err := id.Scan(strings.Split(authorizationHeader, " ")[1]); err != nil {
			return uuid.Nil
		}

		return id
	}
}
