package config

import (
	"errors"
	"fmt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
	"io/ioutil"
	"os"
	"strings"

	k8sAuthenticator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/k8s"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

const authnURLVarName string = "CONJUR_AUTHN_URL"

// AuthnSettings represents a group of authenticator client configuration settings.
type AuthnSettings map[string]string

// NewConfigFromEnv returns a config ConfigFromEnv using the standard file reader for reading certs
func NewConfigFromEnv() (Configuration, error) {
	return ConfigFromEnv(ioutil.ReadFile)
}

// ConfigFromEnv returns a new authenticator configuration object
func ConfigFromEnv(readFileFunc common.ReadFileFunc) (Configuration, error) {
	authnUrl := os.Getenv(authnURLVarName)
	conf, error := getConfiguration(authnUrl)
	if error != nil {
		return nil, error
	}
	envSettings := GatherSettings(conf, os.Getenv)

	errLogs := envSettings.validate(conf, readFileFunc)
	if len(errLogs) > 0 {
		logErrors(errLogs)
		return nil, errors.New(log.CAKC061)
	}

	conf.LoadConfig(envSettings)
	return conf, nil
}

func getConfiguration(url string) (Configuration, error) {
	if strings.Contains(url, k8sAuthenticator.AuthnType) {
		return &k8sAuthenticator.Config{}, nil
	}
	return nil, fmt.Errorf(log.CAKC063, url)
}

// GatherSettings retrieves authenticator client configuration settings from a slice
// of arbitrary `func(key string) string` functions. Values received from 'Getter' functions
// are prioritized in the order that the functions are provided.
func GatherSettings(conf Configuration, getters ...func(key string) string) AuthnSettings {
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

	return settings
}

// Validate confirms that the given AuthnSettings yield a valid authenticator
// client configuration. Returns a list of Error logs.
func (settings AuthnSettings) validate(conf Configuration, readFileFunc common.ReadFileFunc) []error {
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
