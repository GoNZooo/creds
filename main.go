package main

import (
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

	server := creds.Server{}
	server.Serve(int(port), databaseOptions, adminScope)
}
