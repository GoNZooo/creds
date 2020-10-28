package main

import (
	"log"

	"github.com/go-pg/pg/v10/orm"

	"creds/creds"
)

func main() {
	port := creds.GetRequiredEnvironmentIntegerEnvironmentVariable("PORT")

	databaseOptions := creds.DatabaseOptions{
		Host:     creds.GetRequiredEnvironmentVariable("DATABASE_HOST"),
		Port:     creds.GetRequiredEnvironmentIntegerEnvironmentVariable("DATABASE_PORT"),
		Database: creds.GetRequiredEnvironmentVariable("DATABASE_DATABASE"),
		User:     creds.GetRequiredEnvironmentVariable("DATABASE_USER"),
		Password: creds.GetRequiredEnvironmentVariable("DATABASE_PASSWORD"),
	}

	adminScope := creds.GetRequiredEnvironmentVariable("ADMIN_SCOPE")
	database := creds.ConnectToDatabase(databaseOptions)
	err := creds.CreateSchema(database, &orm.CreateTableOptions{Temp: false, IfNotExists: true})
	if err != nil {
		log.Panicf("`CreateSchema` error: %server", err.Error())
	}

	server := creds.Server{}
	server.Serve(port, database, adminScope)
}
