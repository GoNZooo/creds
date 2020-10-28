package creds

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func TestGetUsers(t *testing.T) {
	url := "/users"
	setup := initializeTestData(nil)
	router := new(httprouter.Router)
	setupRoutes(router, setup.database, setup.adminScope)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", url, nil)
	if request == nil {
		log.Panic("Created request is `nil`")
	}
	if err != nil {
		log.Panicf("Unable to create request: %s", err.Error())
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", setup.adminToken.String()))
	router.ServeHTTP(recorder, request)

	users := make([]User, 0)
	if err := json.NewDecoder(recorder.Body).Decode(&users); err != nil {
		log.Panicf("Unable to decode response into `[]User`: %s", err.Error())
	}

	if len(users) != 1 {
		log.Panicf("Unexpected user list length: %d", len(users))
	}

	runBadTokenTests(router, url)
}

func TestGetUser(t *testing.T) {
	setup := initializeTestData(nil)

	existingUserUrl := fmt.Sprintf("/user/%s", setup.adminId)
	badUserUrl := fmt.Sprintf("/user/%s", uuid.New())
	router := new(httprouter.Router)
	setupRoutes(router, setup.database, setup.adminScope)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", existingUserUrl, nil)
	if request == nil {
		log.Panic("Created request is `nil`")
	}
	if err != nil {
		log.Panicf("Unable to create request: %s", err.Error())
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", setup.adminToken.String()))
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		log.Panicf("Bad code for existing user: %d", recorder.Code)
	}

	user := User{}
	if err := json.NewDecoder(recorder.Body).Decode(&user); err != nil {
		log.Panicf("Unable to decode response into `User`: %s", err.Error())
	}

	// Bad user (non-existent)
	recorder = httptest.NewRecorder()
	request, err = http.NewRequest("GET", badUserUrl, nil)
	if request == nil {
		log.Panic("Created request is `nil`")
	}
	if err != nil {
		log.Panicf("Unable to create request: %s", err.Error())
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", setup.adminToken.String()))
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		log.Panicf("Bad code for not found user: %d", recorder.Code)
	}

	runBadTokenTests(router, existingUserUrl)
}

func runBadTokenTests(router *httprouter.Router, url string) {
	// Bad token test
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", url, nil)

	if request == nil {
		log.Panic("Created request is `nil`")
	}
	if err != nil {
		log.Panicf("Unable to create request: %s", err.Error())
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", uuid.New()))
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		log.Panicf("Bad token does not return unauthorized status code: %d", recorder.Code)
	}

	// No token test
	recorder = httptest.NewRecorder()
	request, err = http.NewRequest("GET", url, nil)
	if request == nil {
		log.Panic("Created request is `nil`")
	}
	if err != nil {
		log.Panicf("Unable to create request: %s", err.Error())
	}
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		log.Panicf("Bad token does not return unauthorized status code: %d", recorder.Code)
	}
}
