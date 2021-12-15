package creators

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
	k8sAuthenitcator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/k8s"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

const AuthnURLVarName string = "CONJUR_AUTHN_URL"

// AuthnSettings represents a group of authenticator client configuration settings.
type AuthnSettings map[string]string

// NewConfigFromEnv returns a config ConfigFromEnv using the standard file reader for reading certs
func NewConfigFromEnv() (common.ConfigurationInterface, error) {
	return ConfigFromEnv(ioutil.ReadFile)
}

// ConfigFromEnv returns a new authenticator configuration object
func ConfigFromEnv(readFileFunc common.ReadFileFunc) (common.ConfigurationInterface, error) {
	authnUrl := os.Getenv(AuthnURLVarName)
	conf, error := GetConfiguration(authnUrl)
	if error != nil {
		return nil, error
	}
	envSettings := GatherSettings(conf, os.Getenv)

	errLogs := envSettings.Validate(conf, readFileFunc)
	if len(errLogs) > 0 {
		logErrors(errLogs)
		return nil, errors.New(log.CAKC061)
	}

	conf.LoadConfig(envSettings)
	return conf, nil
}

func GetConfiguration(url string) (common.ConfigurationInterface, error) {
	if strings.Contains(url, k8sAuthenitcator.AuthnType) {
		return &k8sAuthenitcator.Config{}, nil
	}
	return nil, fmt.Errorf(log.CAKC063, url)
}

// GatherSettings retrieves authenticator client configuration settings from a slice
// of arbitrary `func(key string) string` functions. Values received from 'Getter' functions
// are prioritized in the order that the functions are provided.
func GatherSettings(conf common.ConfigurationInterface, getters ...func(key string) string) AuthnSettings {
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
func (settings AuthnSettings) Validate(conf common.ConfigurationInterface, readFileFunc common.ReadFileFunc) []error {
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
