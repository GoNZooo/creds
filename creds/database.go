package creds

import (
	"fmt"
	"log"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/google/uuid"
)

func ConnectToDatabase(options DatabaseOptions) *pg.DB {
	address := fmt.Sprintf("%s:%d", options.Host, options.Port)

	return pg.Connect(&pg.Options{
		Addr:     address,
		Database: options.Database,
		User:     options.User,
		Password: options.Password,
	})
}

func CreateSchema(database *pg.DB, options *orm.CreateTableOptions) error {
	models := []interface{}{(*User)(nil), (*Token)(nil)}

	for _, m := range models {
		err := database.Model(m).CreateTable(options)
		if err != nil {
			return err
		}
	}

	return nil
}

type setUpData struct {
	adminId    uuid.UUID
	adminToken uuid.UUID
	database   *pg.DB
	adminScope string
}

func initializeTestData(database *pg.DB) setUpData {
	databaseOptions := DatabaseOptions{
		Host:     GetRequiredEnvironmentVariable("TEST_DATABASE_HOST"),
		Port:     GetRequiredEnvironmentIntegerEnvironmentVariable("TEST_DATABASE_PORT"),
		Database: GetRequiredEnvironmentVariable("TEST_DATABASE_DATABASE"),
		User:     GetRequiredEnvironmentVariable("TEST_DATABASE_USER"),
		Password: GetRequiredEnvironmentVariable("TEST_DATABASE_PASSWORD"),
	}

	adminScope := GetRequiredEnvironmentVariable("TEST_ADMIN_SCOPE")
	if database == nil {
		database = ConnectToDatabase(databaseOptions)
	}
	models := []interface{}{(*User)(nil), (*Token)(nil)}
	for _, model := range models {
		if err := database.Model(model).CreateTable(
			&orm.CreateTableOptions{Temp: true, IfNotExists: true, FKConstraints: true},
		); err != nil {
			log.Panicf("Unable to create test database: %s", err.Error())
		}
	}

	adminId, err := insertUser(database, "Admin", "Admin")
	if err != nil {
		log.Panicf("Unable to create admin user: %s", err.Error())
	}

	adminToken, err := insertToken(database, adminId, adminScope, time.Time{}, time.Time{})
	if err != nil {
		log.Panicf("Unable to create admin token: %s", err.Error())
	}

	return setUpData{adminId: adminId, adminToken: adminToken, database: database, adminScope: adminScope}
}
