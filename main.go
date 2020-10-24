package main

import (
	"log"
	"os"
	"strconv"

	"creds/creds"
)

func main() {
	port := getRequiredEnvironmentIntegerEnvironmentVariable("PORT")

	databaseOptions := creds.DatabaseOptions{
		Host:     getRequiredEnvironmentVariable("DATABASE_HOST"),
		Port:     getRequiredEnvironmentIntegerEnvironmentVariable("DATABASE_PORT"),
		Database: getRequiredEnvironmentVariable("DATABASE_DATABASE"),
		User:     getRequiredEnvironmentVariable("DATABASE_USER"),
		Password: getRequiredEnvironmentVariable("DATABASE_PASSWORD"),
	}

	server := creds.Server{}
	server.Serve(int(port), databaseOptions)
}

func getRequiredEnvironmentIntegerEnvironmentVariable(key string) int {
	portString := getRequiredEnvironmentVariable(key)
	portInteger64, err := strconv.ParseInt(portString, 10, 32)
	if err != nil {
		log.Panicf("Environment variable '%s' is not parsable as integer: %s", key, err.Error())
	}

	return int(portInteger64)
}

func getRequiredEnvironmentVariable(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Panicf("Environment variable '%s' not set, aborting.", key)
	}

	return value
}
