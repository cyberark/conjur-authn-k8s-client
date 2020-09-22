package utils

import (
	"os"
	"strconv"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// GetEnvFunc is a function that returns a value for an environment variable
type GetEnvFunc func(envVarName string) string

// defaultGetEnv is the default GetEnvFunc. It uses the standard OS Getenv.
var defaultGetEnvFunc = os.Getenv

// IntFromEnvOrDefault extracts the value of the given environment variable,
// or the default value if the environment variable is not defined, and returns
// it as an int object
func IntFromEnvOrDefault(
	envVarName string,
	defaultValue string,
	getEnv GetEnvFunc,
) (int, error) {
	value := envVarValueOrDefault(envVarName, defaultValue, getEnv)
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return 0, log.RecordedError(log.CAKC010, envVarName, err)
	}

	return valueInt, nil
}

// DurationFromEnvOrDefault extracts the value of the given environment variable,
// or the default value if the environment variable is not defined, and returns
// it as a time.Duration object
func DurationFromEnvOrDefault(
	envVarName string,
	defaultValue string,
	getEnv GetEnvFunc,
) (time.Duration, error) {
	value := envVarValueOrDefault(envVarName, defaultValue, getEnv)
	valueDuration, err := time.ParseDuration(value)
	if err != nil {
		return 0, log.RecordedError(log.CAKC010, envVarName, err)
	}

	return valueDuration, nil
}

func envVarValueOrDefault(envVarName string, defaultValue string, getEnv GetEnvFunc) string {
	if getEnv == nil {
		getEnv = defaultGetEnvFunc
	}
	value := getEnv(envVarName)
	if len(value) > 0 {
		return value
	}

	return defaultValue
}
