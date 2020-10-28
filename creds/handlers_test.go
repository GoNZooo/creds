package creds

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUsers(t *testing.T) {
	setup := initializeTestData(nil)

	handler := handleGetUsers(setup.database, setup.adminScope)
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/users", nil)
	if request == nil {
		log.Panic("Created request is `nil`")
	}
	if err != nil {
		log.Panicf("Unable to create request: %s", err.Error())
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", setup.adminToken.String()))
	handler.ServeHTTP(recorder, request)

	users := make([]User, 0)
	if err := json.NewDecoder(recorder.Body).Decode(&users); err != nil {
		log.Panicf("Unable to decode response into `[]User`: %s", err.Error())
	}

	if len(users) != 1 {
		log.Panicf("Unexpected user list length: %d", len(users))
	}
}
