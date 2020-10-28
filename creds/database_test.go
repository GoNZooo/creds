package creds

import (
	"log"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/google/uuid"
)

type setUpData struct {
	adminId    uuid.UUID
	adminToken uuid.UUID
	database   *pg.DB
	adminScope string
}

func setUp() setUpData {
	databaseOptions := DatabaseOptions{
		Host:     GetRequiredEnvironmentVariable("TEST_DATABASE_HOST"),
		Port:     GetRequiredEnvironmentIntegerEnvironmentVariable("TEST_DATABASE_PORT"),
		Database: GetRequiredEnvironmentVariable("TEST_DATABASE_DATABASE"),
		User:     GetRequiredEnvironmentVariable("TEST_DATABASE_USER"),
		Password: GetRequiredEnvironmentVariable("TEST_DATABASE_PASSWORD"),
	}

	adminScope := GetRequiredEnvironmentVariable("TEST_ADMIN_SCOPE")
	database := connectToDatabase(databaseOptions)
	models := []interface{}{(*User)(nil), (*Token)(nil)}
	for _, model := range models {
		if err := database.Model(model).CreateTable(
			&orm.CreateTableOptions{Temp: true, IfNotExists: true},
		); err != nil {
			log.Panicf("Unable to create test database: %s", err.Error())
		}
	}

	adminId, err := insertUser(database, "Admin", "Admin")
	if err != nil {
		log.Panicf("Unable to create admin user: %s", err.Error())
	}

	adminToken, err := insertToken(database, adminId, adminScope)
	if err != nil {
		log.Panicf("Unable to create admin token: %s", err.Error())
	}

	return setUpData{adminId: adminId, adminToken: adminToken, database: database, adminScope: adminScope}
}

func TestAddAndGetUser(t *testing.T) {
	d := setUp()

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
	d := setUp()

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
