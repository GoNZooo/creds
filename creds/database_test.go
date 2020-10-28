package creds

import (
	"log"
	"testing"
)

func TestAddAndGetUser(t *testing.T) {
	d := initializeTestDatabase()

	name := "TestUser"
	username := "TestUser"
	id, err := insertUser(d.database, name, username)
	if err != nil {
		log.Panicf("Unable to add user: %s", err.Error())
	}
	if id.ID() == 0 {
		log.Panicf("ID is invalid")
	}

	user, err := getUserById(d.database, id)
	if err != nil {
		log.Panicf("Error getting user: %s", err.Error())
	}
	if user == nil || user.Name != name || user.Username != username {
		log.Panicf("User with id '%s' does not exist or has incorrect data: %+v", id, user)
	}
}

func TestAddAndGetToken(t *testing.T) {
	d := initializeTestDatabase()

	name := "TestUserForToken"
	username := "TestUserForToken"
	id, err := insertUser(d.database, name, username)
	if err != nil {
		log.Panicf("Unable to add user: %s", err.Error())
	}
	if id.ID() == 0 {
		log.Panicf("ID is invalid")
	}

	user, err := getUserById(d.database, id)
	if err != nil {
		log.Panicf("Error getting user: %s", err.Error())
	}
	if user == nil || user.Name != name || user.Username != username {
		log.Panicf("User with id '%s' does not exist or has incorrect data: %+v", id, user)
	}

	tokenId, err := insertToken(d.database, user.Id, "TestingScope")
	if err != nil {
		log.Panicf("Unable to insert token: %s", err.Error())
	}

	token, err := getTokenById(d.database, tokenId)
	if err != nil {
		log.Panicf("Unable to get inserted token: %s", err.Error())
	}
	if token == nil {
		log.Panic("Got nil token")
	}

	if token.Scope != "TestingScope" || token.UserId != user.Id {
		log.Panicf("Returned token is incorrect:\n%+v", token)
	}
}
