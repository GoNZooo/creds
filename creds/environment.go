package creds

import (
	"log"
	"os"
	"strconv"
)

func GetRequiredEnvironmentIntegerEnvironmentVariable(key string) int {
	portString := GetRequiredEnvironmentVariable(key)
	portInteger64, err := strconv.ParseInt(portString, 10, 32)
	if err != nil {
		log.Panicf("Environment variable '%s' is not parsable as integer: %s", key, err.Error())
	}

	return int(portInteger64)
}

func GetRequiredEnvironmentVariable(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Panicf("Environment variable '%s' not set, aborting.", key)
	}

	return value
}
