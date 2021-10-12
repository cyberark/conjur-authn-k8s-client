package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// Config defines the configuration parameters
// for the authentication requests
type Config struct {
	Account                   string
	ClientCertPath            string
	ClientCertRetryCountLimit int
	ContainerMode             string
	ConjurVersion             string
	InjectCertLogPath         string
	PodName                   string
	PodNamespace              string
	SSLCertificate            []byte
	TokenFilePath             string
	TokenRefreshTimeout       time.Duration
	URL                       string
	Username                  *Username
}

// AuthnSettings represents a group of authenticator client configuration settings.
type AuthnSettings map[string]string

// Default settings (this comment added to satisfy linter)
const (
	DefaultClientCertPath    = "/etc/conjur/ssl/client.pem"
	DefaultInjectCertLogPath = "/tmp/conjur_copy_text_output.log"
	DefaultTokenFilePath     = "/run/conjur/access-token"
	DefaultContainerMode     = "application"

	DefaultConjurVersion = "5"

	// DefaultTokenRefreshTimeout is the default time the system waits to reauthenticate on error
	DefaultTokenRefreshTimeout = "6m0s"

	// DefaultClientCertRetryCountLimit is the amount of times we wait after successful
	// login for the client certificate file to exist, where each time we wait for a second.
	DefaultClientCertRetryCountLimit = "10"
)

var requiredEnvVariables = []string{
	"CONJUR_AUTHN_URL",
	"CONJUR_ACCOUNT",
	"CONJUR_AUTHN_LOGIN",
	"MY_POD_NAMESPACE",
	"MY_POD_NAME",
}

var envVariables = []string{
	"CONJUR_ACCOUNT",
	"CONJUR_AUTHN_LOGIN",
	"CONJUR_AUTHN_TOKEN_FILE",
	"CONJUR_AUTHN_URL",
	"CONJUR_CERT_FILE",
	"CONJUR_CLIENT_CERT_PATH",
	"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT",
	"CONJUR_SSL_CERTIFICATE",
	"CONJUR_TOKEN_TIMEOUT",
	"CONJUR_VERSION",
	"CONTAINER_MODE",
	"DEBUG",
	"MY_POD_NAME",
	"MY_POD_NAMESPACE",
}

var defaultValues = map[string]string{
	"CONJUR_CLIENT_CERT_PATH":              DefaultClientCertPath,
	"CONJUR_AUTHN_TOKEN_FILE":              DefaultTokenFilePath,
	"CONJUR_VERSION":                       DefaultConjurVersion,
	"CONJUR_TOKEN_TIMEOUT":                 DefaultTokenRefreshTimeout,
	"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": DefaultClientCertRetryCountLimit,
	"CONTAINER_MODE":                       DefaultContainerMode,
}

func defaultEnv(key string) string {
	return defaultValues[key]
}

// ReadFileFunc defines the interface for reading an SSL Certificate from the env
type ReadFileFunc func(filename string) ([]byte, error)

// NewFromEnv returns a config FromEnv using the standard file reader for reading certs
func NewFromEnv() (*Config, error) {
	return FromEnv(ioutil.ReadFile)
}

// FromEnv returns a new authenticator configuration object
func FromEnv(readFileFunc ReadFileFunc) (*Config, error) {
	envSettings := GatherSettings(os.Getenv)

	errLogs := envSettings.Validate(readFileFunc)
	if len(errLogs) > 0 {
		logErrors(errLogs)
		return nil, errors.New(log.CAKC061)
	}

	return envSettings.NewConfig(), nil
}

// GatherSettings retrieves authenticator client configuration settings from a slice
// of arbitrary `func(key string) string` functions. Values received from 'Getter' functions
// are prioritized in the order that the functions are provided.
func GatherSettings(getters ...func(key string) string) AuthnSettings {
	getters = append(getters, defaultEnv)

	getEnv := func(key string) string {
		var val string
		for _, getter := range getters {
			val = getter(key)
			if len(val) > 0 {
				return val
			}
		}
		return ""
	}

	settings := make(AuthnSettings)

	for _, key := range envVariables {
		value := getEnv(key)
		settings[key] = value
	}

	return settings
}

// Validate confirms that the given AuthnSettings yield a valid authenticator
// client configuration. Returns a list of Error logs.
func (settings AuthnSettings) Validate(readFileFunc ReadFileFunc) []error {
	errorLogs := []error{}

	// ensure required values exist
	for _, key := range requiredEnvVariables {
		if settings[key] == "" {
			errorLogs = append(errorLogs, fmt.Errorf(log.CAKC062, key))
		}
	}

	// ensure provided values are of the correct type
	for _, key := range envVariables {
		err := validateSetting(key, settings[key])
		if err != nil {
			errorLogs = append(errorLogs, err)
		}
	}

	// ensure that the certificate settings are valid
	cert, err := readSSLCert(settings, readFileFunc)
	if err != nil {
		errorLogs = append(errorLogs, err)
	} else {
		if settings["CONJUR_SSL_CERTIFICATE"] == "" {
			settings["CONJUR_SSL_CERTIFICATE"] = string(cert)
		}
	}

	return errorLogs
}

// NewConfig provides a new authenticator configuration from an AuthnSettings map.
func (settings AuthnSettings) NewConfig() *Config {
	configureLogLevel(settings["DEBUG"])

	config := &Config{}
	config.InjectCertLogPath = DefaultInjectCertLogPath

	for key, value := range settings {
		switch key {
		case "CONJUR_ACCOUNT":
			config.Account = value
		case "CONJUR_AUTHN_LOGIN":
			username, _ := NewUsername(value)
			config.Username = username
		case "CONJUR_AUTHN_URL":
			config.URL = value
		case "CONJUR_SSL_CERTIFICATE":
			config.SSLCertificate = []byte(value)
		case "CONTAINER_MODE":
			config.ContainerMode = value
		case "MY_POD_NAME":
			config.PodName = value
		case "MY_POD_NAMESPACE":
			config.PodNamespace = value
		case "CONJUR_AUTHN_TOKEN_FILE":
			config.TokenFilePath = value
		case "CONJUR_CLIENT_CERT_PATH":
			config.ClientCertPath = value
		case "CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT":
			limit, _ := strconv.Atoi(value)
			config.ClientCertRetryCountLimit = limit
		case "CONJUR_TOKEN_TIMEOUT":
			timeout, _ := durationFromString(key, value)
			config.TokenRefreshTimeout = timeout
		case "CONJUR_VERSION":
			config.ConjurVersion = value
		}
	}

	return config
}

func configureLogLevel(level string) {
	validVal := "true"
	if level == validVal {
		log.EnableDebugMode()
	} else if level != "" {
		// In case "DEBUG" is configured with incorrect value
		log.Warn(log.CAKC034, level, validVal)
	}
}

func logErrors(errLogs []error) {
	for _, err := range errLogs {
		log.Error(err.Error())
	}
}

func validTimeout(key, timeoutStr string) error {
	_, err := durationFromString(key, timeoutStr)
	return err
}

func durationFromString(key, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf(log.CAKC060, key, value)
	}
	return duration, nil
}

func validInt(key, value string) error {
	_, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf(log.CAKC060, key, value)
	}
	return nil
}

func validUsername(key, value string) error {
	_, err := NewUsername(value)
	return err
}

func validContainerMode(key, value string) error {
	for _, validMode := range []string{"init", "application"} {
		if value == validMode {
			return nil
		}
	}
	return fmt.Errorf(log.CAKC060, key, value)
}

func validConjurVersion(key, version string) error {
	// Only versions '4' & '5' are allowed, with '5' being used as the default
	switch version {
	case "4":
		break
	case "5":
		break
	default:
		return fmt.Errorf(log.CAKC060, key, version)
	}

	return nil
}

func validateSetting(key string, value string) error {
	switch key {
	case "CONJUR_AUTHN_LOGIN":
		return validUsername(key, value)
	case "CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT":
		return validInt(key, value)
	case "CONJUR_TOKEN_TIMEOUT":
		return validTimeout(key, value)
	case "CONJUR_VERSION":
		return validConjurVersion(key, value)
	case "CONTAINER_MODE":
		return validContainerMode(key, value)
	default:
		return nil
	}
}

func readSSLCert(settings map[string]string, readFile ReadFileFunc) ([]byte, error) {
	SSLCert := settings["CONJUR_SSL_CERTIFICATE"]
	SSLCertPath := settings["CONJUR_CERT_FILE"]
	if SSLCert == "" && SSLCertPath == "" {
		return nil, errors.New(log.CAKC007)
	}

	if SSLCert != "" {
		return []byte(SSLCert), nil
	}
	return readFile(SSLCertPath)
}
