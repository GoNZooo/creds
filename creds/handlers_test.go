package creds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/guregu/null.v4"
)

func TestGetUsers(t *testing.T) {
	url := "/users"
	setup := initializeTestData(nil)
	router := new(httprouter.Router)
	setupRoutes(router, setup.database, setup.adminScope)

	withRecorder("GET",
		url,
		nil,
		[]headerEntry{bearerToken(setup.adminToken)},
		router,
		func(recorder *httptest.ResponseRecorder, request *http.Request) {
			users := make([]User, 0)
			if err := json.NewDecoder(recorder.Body).Decode(&users); err != nil {
				log.Panicf("Unable to decode response into `[]User`: %s", err.Error())
			}

			if len(users) != 1 {
				log.Panicf("Unexpected user list length: %d", len(users))
			}

			if users[0].Id != setup.adminId || users[0].Tokens[0].Id != setup.adminToken {
				log.Panicf("Retrieved data doesn't match setup data:\n\tSetup: %+v\n\tRetrieved User: %+v\n", setup, users[0])
			}
		})

	runBadTokenTests(router, url)
}

func TestGetTokens(t *testing.T) {
	url := "/tokens"
	setup := initializeTestData(nil)
	router := new(httprouter.Router)
	setupRoutes(router, setup.database, setup.adminScope)

	withRecorder("GET",
		url,
		nil,
		[]headerEntry{bearerToken(setup.adminToken)},
		router,
		func(recorder *httptest.ResponseRecorder, request *http.Request) {
			tokens := make([]Token, 0)
			if err := json.NewDecoder(recorder.Body).Decode(&tokens); err != nil {
				log.Panicf("Unable to decode response into `[]User`: %s", err.Error())
			}

			if len(tokens) != 1 {
				log.Panicf("Unexpected token list length: %d", len(tokens))
			}

			if tokens[0].Id != setup.adminToken {
				log.Panicf("Retrieved data doesn't match setup data:\n\tSetup: %+v\n\tRetrieved Token: %+v\n", setup, tokens[0])
			}
		})

	runBadTokenTests(router, url)
}

func TestGetUser(t *testing.T) {
	setup := initializeTestData(nil)

	existingUserUrl := fmt.Sprintf("/user/%s", setup.adminId)
	badUserUrl := fmt.Sprintf("/user/%s", uuid.New())
	router := new(httprouter.Router)
	setupRoutes(router, setup.database, setup.adminScope)

	headers := []headerEntry{bearerToken(setup.adminToken)}
	withRecorder("GET",
		existingUserUrl,
		nil,
		headers,
		router,
		func(recorder *httptest.ResponseRecorder, request *http.Request) {
			if recorder.Code != http.StatusOK {
				log.Panicf("Bad code for existing user: %d", recorder.Code)
			}

			user := User{}
			if err := json.NewDecoder(recorder.Body).Decode(&user); err != nil {
				log.Panicf("Unable to decode response into `User`: %s", err.Error())
			}
		})

	// Bad user (non-existent)
	withRecorder("GET",
		badUserUrl,
		nil,
		headers,
		router,
		func(recorder *httptest.ResponseRecorder, request *http.Request) {
			if recorder.Code != http.StatusNotFound {
				log.Panicf("Bad code for not found user: %d", recorder.Code)
			}
		})

	runBadTokenTests(router, existingUserUrl)
}

func TestAddUser(t *testing.T) {
	setup := initializeTestData(nil)

	url := "/users"
	router := new(httprouter.Router)
	setupRoutes(router, setup.database, setup.adminScope)

	headers := []headerEntry{bearerToken(setup.adminToken)}
	parameterBytes, err := json.Marshal(addUserParameters{
		Username: null.StringFrom("DJ Testo"),
		Name:     null.StringFrom("Test Testersson"),
	})
	if err != nil {
		log.Panicf("Unable to serialize `addUserParameters`: %s", err.Error())
	}
	parameterReader := bytes.NewReader(parameterBytes)

	withRecorder("POST",
		url,
		parameterReader,
		headers,
		router,
		func(recorder *httptest.ResponseRecorder, request *http.Request) {
			if recorder.Code != http.StatusOK {
				log.Panicf("Bad status code for adding user: %d", recorder.Code)
			}

			id := uuid.UUID{}
			if err := json.NewDecoder(recorder.Body).Decode(&id); err != nil {
				log.Panicf("Unable to read body into UUID: %s", err.Error())
			}
		})

	runBadTokenTests(router, url)
}

func TestAddToken(t *testing.T) {
	setup := initializeTestData(nil)

	url := "/tokens"
	router := new(httprouter.Router)
	setupRoutes(router, setup.database, setup.adminScope)

	headers := []headerEntry{bearerToken(setup.adminToken)}
	parameterBytes, err := json.Marshal(addTokenParameters{
		UserId: setup.adminId,
		Scope:  null.StringFrom("testing-scope"),
	})
	if err != nil {
		log.Panicf("Unable to serialize `addUserParameters`: %s", err.Error())
	}
	parameterReader := bytes.NewReader(parameterBytes)

	withRecorder("POST",
		url,
		parameterReader,
		headers,
		router,
		func(recorder *httptest.ResponseRecorder, request *http.Request) {
			if recorder.Code != http.StatusOK {
				log.Panicf("Bad status code for adding token: %d\n\tBody: %s\n\tSetup.AdminToken: %s\tSetup.AdminID: %s", recorder.Code, recorder.Body, setup.adminToken, setup.adminId)
			}

			id := uuid.UUID{}
			if err := json.NewDecoder(recorder.Body).Decode(&id); err != nil {
				log.Panicf("Unable to read body into UUID: %s", err.Error())
			}
		})

	noSuchUserBytes, err := json.Marshal(addTokenParameters{
		UserId: uuid.New(),
		Scope:  null.StringFrom(setup.adminScope),
	})
	if err != nil {
		log.Panicf("Unable to serialize `addUserParameters`: %s", err.Error())
	}
	noSuchUserParameterReader := bytes.NewReader(noSuchUserBytes)

	withRecorder("POST",
		url,
		noSuchUserParameterReader,
		headers,
		router,
		func(recorder *httptest.ResponseRecorder, request *http.Request) {
			if recorder.Code != http.StatusNotFound {
				log.Panicf("Bad status code for adding token for non-existant user: %d", recorder.Code)
			}
		})

	runBadTokenTests(router, url)
}

type headerEntry struct {
	key   string
	value string
}

func withRecorder(method string,
	url string,
	body io.Reader,
	headers []headerEntry,
	router *httprouter.Router,
	f func(recorder *httptest.ResponseRecorder, request *http.Request),
) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Panicf("Error creating request: %s", err.Error())
	}
	if request == nil {
		log.Panicln("Nil request on create")
	}

	for _, h := range headers {
		request.Header.Set(h.key, h.value)
	}

	router.ServeHTTP(recorder, request)
	f(recorder, request)
}

func runBadTokenTests(router *httprouter.Router, url string) {
	// Bad token test
	withRecorder("GET",
		url,
		nil,
		[]headerEntry{bearerToken(uuid.New())},
		router,
		func(recorder *httptest.ResponseRecorder, request *http.Request) {
			if recorder.Code != http.StatusUnauthorized {
				log.Panicf("Bad token does not return unauthorized status code: %d", recorder.Code)
			}
		})

	// No token test
	withRecorder("GET",
		url,
		nil,
		[]headerEntry{},
		router,
		func(recorder *httptest.ResponseRecorder, request *http.Request) {
			if recorder.Code != http.StatusUnauthorized {
				log.Panicf("No token does not return unauthorized status code: %d", recorder.Code)
			}
		})
}

func bearerToken(token fmt.Stringer) headerEntry {
	return headerEntry{"Authorization", fmt.Sprintf("Bearer %s", token)}
}
