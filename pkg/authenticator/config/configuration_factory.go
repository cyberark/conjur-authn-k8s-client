package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"

	jwtAuthenticator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/jwt"
	k8sAuthenticator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/k8s"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

const authnURLVarName string = "CONJUR_AUTHN_URL"
const defaultLogLevel string = "info"

// AuthnSettings represents a group of authenticator client configuration settings.
type AuthnSettings map[string]string

// NewConfigFromEnv returns a config ConfigFromEnv using the standard file reader for reading certs
func NewConfigFromEnv() (Configuration, error) {
	return ConfigFromEnv(os.ReadFile)
}

// ConfigFromEnv returns a new authenticator configuration object
func ConfigFromEnv(readFileFunc common.ReadFileFunc) (Configuration, error) {
	return NewConfigFromCustomEnv(readFileFunc, os.Getenv)
}

func NewConfigFromCustomEnv(readFileFunc common.ReadFileFunc, customEnv func(key string) string) (Configuration, error) {
	log.Debug(log.CAKC068)
	logLevel := getConfiguredLogLevel(customEnv)
	log.SetLogLevel(logLevel)
	authnUrl := customEnv(authnURLVarName)
	conf, err := getConfiguration(authnUrl)
	if err != nil {
		return nil, err
	}
	envSettings := GatherSettings(conf, customEnv)

	errLogs := envSettings.validate(conf, readFileFunc)
	if len(errLogs) > 0 {
		logErrors(errLogs)
		return nil, errors.New(log.CAKC061)
	}
	log.Debug(log.CAKC074)

	conf.LoadConfig(envSettings)
	return conf, nil
}

// GatherSettings retrieves authenticator client configuration settings from a slice
// of arbitrary `func(key string) string` functions. Values received from 'Getter' functions
// are prioritized in the order that the functions are provided.
func GatherSettings(conf Configuration, getters ...func(key string) string) AuthnSettings {
	log.Debug(log.CAKC071)
	defaultVariables := conf.GetDefaultValues()

	getDefault := func(key string) string {
		return defaultVariables[key]
	}

	getters = append(getters, getDefault)
	settings := make(AuthnSettings)
	getEnv := getConfigVariable(getters...)

	for _, key := range conf.GetEnvVariables() {
		value := getEnv(key)
		settings[key] = value
	}

	log.Debug(log.CAKC072)
	return settings
}

func getConfiguration(url string) (Configuration, error) {
	switch {
	case strings.Contains(url, k8sAuthenticator.AuthnType):
		log.Info(log.CAKC070, k8sAuthenticator.AuthnType)
		return &k8sAuthenticator.Config{}, nil
	case strings.Contains(url, jwtAuthenticator.AuthnType):
		log.Info(log.CAKC070, jwtAuthenticator.AuthnType)
		return &jwtAuthenticator.Config{}, nil
	default:
		return nil, fmt.Errorf(log.CAKC063, url)
	}
}

// Validate confirms that the given AuthnSettings yield a valid authenticator
// client configuration. Returns a list of Error logs.
func (settings AuthnSettings) validate(conf Configuration, readFileFunc common.ReadFileFunc) []error {
	log.Debug(log.CAKC073)
	errorLogs := []error{}

	// ensure required values exist
	for _, key := range conf.GetRequiredVariables() {
		if settings[key] == "" {
			errorLogs = append(errorLogs, fmt.Errorf(log.CAKC062, key))
		}
	}

	// ensure provided values are of the correct type
	for _, key := range conf.GetEnvVariables() {
		err := common.ValidateSetting(key, settings[key])
		if err != nil {
			errorLogs = append(errorLogs, err)
		}
	}

	// ensure that the certificate settings are valid
	cert, err := common.ReadSSLCert(settings, readFileFunc)
	if err != nil {
		errorLogs = append(errorLogs, err)
	} else {
		if settings["CONJUR_SSL_CERTIFICATE"] == "" {
			settings["CONJUR_SSL_CERTIFICATE"] = string(cert)
		}
	}

	return errorLogs
}

func logErrors(errLogs []error) {
	for _, err := range errLogs {
		log.Error(err.Error())
	}
}

func getConfigVariable(getters ...func(key string) string) func(string) string {
	return func(key string) string {
		var val string
		for _, getter := range getters {
			val = getter(key)
			if len(val) > 0 {
				return val
			}
		}
		return ""
	}
}

func getConfiguredLogLevel(getConfigFunc func(key string) string) string {
	validLogLevels := []string{"debug", "info", "warn", "error"}
	logLevel := getConfigFunc("LOG_LEVEL")

	if logLevel != "" {
		// If log level is configured, check if it's valid
		for _, validLevel := range validLogLevels {
			if logLevel == validLevel {
				return logLevel
			}
		}

		// If log level is configured but it's invalid, log a warning and return default
		log.Warn(log.CAKC034, logLevel, validLogLevels)
		return defaultLogLevel
	}

	// If log level is not configured, check if debug is configured.
	// This is for backwards compatibility with the old debug env var.
	debugValue := getConfigFunc("DEBUG")
	if debugValue == "true" {
		log.Warn(log.CAKC081)
		return "debug"
	}

	// If neither log level nor debug are configured, return default
	return defaultLogLevel
}
