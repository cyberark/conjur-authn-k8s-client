package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/utils"
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

// Default settings (this comment added to satisfy linter)
const (
	DefaultClientCertPath    = "/etc/conjur/ssl/client.pem"
	DefaultInjectCertLogPath = "/tmp/conjur_copy_text_output.log"
	DefaultTokenFilePath     = "/run/conjur/access-token"

	DefaultConjurVersion = "5"

	// DefaultTokenRefreshTimeout is the default time the system waits to reauthenticate on error
	DefaultTokenRefreshTimeout = "6m0s"

	// DefaultClientCertRetryCountLimit is the amount of times we wait after successful
	// login for the client certificate file to exist, where each time we wait for a second.
	DefaultClientCertRetryCountLimit = "10"
)

var requiredEnvVariables = []string{
	"CONJUR_ACCOUNT",
	"CONJUR_AUTHN_LOGIN",
	"CONJUR_AUTHN_URL",
	"MY_POD_NAME",
	"MY_POD_NAMESPACE",
}

var optionalEnvVariables = []string{
	"CONJUR_AUTHN_TOKEN_FILE",
	"CONJUR_CERT_FILE",
	"CONJUR_CLIENT_CERT_PATH",
	"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT",
	"CONJUR_SSL_CERTIFICATE",
	"CONJUR_TOKEN_TIMEOUT",
	"CONJUR_VERSION",
	"DEBUG",
}

var optionalAnnotations = []string{
	"conjur.org/authn-identity",
	"conjur.org/container-mode",
	"conjur.org/debug-logging",
}

// ReadFileFunc defines the interface for reading an SSL Certificate from the env
type ReadFileFunc func(filename string) ([]byte, error)

// GatherSettings returns a string-to-string map of all provided environment
// variables and parsed, valid annotations that are concerned with authenticator
// client config.
func GatherSettings(annotations map[string]string) map[string]string {
	masterMap := make(map[string]string)

	for _, envvar := range requiredEnvVariables {
		value := os.Getenv(envvar)
		if value != "" {
			masterMap[envvar] = value
		}
	}

	for _, envvar := range optionalEnvVariables {
		value := os.Getenv(envvar)
		if value != "" {
			masterMap[envvar] = value
		}
	}

	for _, annotation := range optionalAnnotations {
		value := annotations[annotation]
		if value != "" {
			masterMap[annotation] = value
		}
	}

	return masterMap
}

// ValidateSettings confirms that the provided environment variable and annotation
// settings yield a valid authenticator client configuration. Returns a list of
// Error logs, and a list of Info logs.
func ValidateSettings(settings map[string]string, readFileFunc ReadFileFunc) ([]error, []error) {
	errorList := []error{}
	infoList := []error{}

	// confirm required envvars
	for _, envvar := range requiredEnvVariables {
		if settings[envvar] == "" {
			errorList = append(errorList, fmt.Errorf(log.CAKC009, envvar))
		}
	}

	// confirm settings that can be set with either envvars or annotations
	settingsWithMultipleSources := map[string][]string{
		"ContainerMode": {"CONTAINER_MODE", "conjur.org/container-mode"},
		"Username":      {"CONJUR_AUTHN_LOGIN", "conjur.org/authn-identity"},
	}

	for setting, sources := range settingsWithMultipleSources {
		envSetting := settings[sources[0]]
		annotSetting := settings[sources[1]]

		if envSetting != "" && annotSetting != "" {
			infoList = append(infoList, fmt.Errorf(log.CAKC060, setting, sources[0], sources[1]))
		}
		if envSetting == "" && annotSetting == "" && setting == "Username" {
			errorList = append(errorList, fmt.Errorf(log.CAKC061, setting, sources[0], sources[1]))
		}
	}

	// confirm valid CONJUR_VERSION
	conjurVersion := settings["CONJUR_VERSION"]
	if err := validConjurVersion(conjurVersion); err != nil && conjurVersion != "" {
		errorList = append(errorList, err)
	}

	// confirm either CONJUR_SSL_CERTIFICATE or CONJUR_CERT_FILE
	// if CONJUR_SSL_CERTIFICATE is empty, overwrite it with the contents of CONJUR_CERT_FILE
	cert, err := readSSLCert(readFileFunc, settings)
	if err != nil {
		errorList = append(errorList, fmt.Errorf(log.CAKC021, err))
	} else if settings["CONJUR_SSL_CERTIFICATE"] == "" {
		settings["CONJUR_SSL_CERTIFICATE"] = string(cert)
	}

	// Parse token refresh rate if one is provided from env
	tokenRefreshTimeout, err := utils.DurationFromEnvOrDefault(
		"CONJUR_TOKEN_TIMEOUT",
		DefaultTokenRefreshTimeout,
		nil,
	)
	if err != nil {
		errorList = append(errorList, err)
	} else {
		settings["CONJUR_TOKEN_TIMEOUT"] = tokenRefreshTimeout.String()
	}

	// Parse client cert retry count limit if one is provided from env
	clientCertRetryCountLimit, err := utils.IntFromEnvOrDefault(
		"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT",
		DefaultClientCertRetryCountLimit,
		nil,
	)
	if err != nil {
		errorList = append(errorList, err)
	} else {
		settings["CONJUR_CLIENT_CERT_RETRY_LIMIT"] = strconv.Itoa(clientCertRetryCountLimit)
	}

	return errorList, infoList
}

// NewFromEnv returns a config created only from envvar settings, using
// the standard file reader for reading certs
func NewFromEnv() (*Config, error) {
	return FromEnv(ioutil.ReadFile)
}

// FromEnv returns a new authenticator configuration object using only envvar settings
func FromEnv(readFileFunc ReadFileFunc) (*Config, error) {
	envSettings := GatherSettings(map[string]string{})

	errLogs, infoLogs := ValidateSettings(envSettings, readFileFunc)
	logErrors(errLogs, infoLogs)
	if len(errLogs) > 0 {
		return nil, errors.New(log.CAKC062)
	}

	return NewConfig(envSettings), nil
}

// NewConfig returns a new authenticator configuration object from a
// string-to-string map of pre-validated authenticator settings compiled
// from environment variables and supplied annotations
func NewConfig(settings map[string]string) *Config {
	logLevel := settings["conjur.org/debug-logging"]
	if logLevel == "" {
		logLevel = settings["DEBUG"]
	}
	configureLogLevel(logLevel)

	config := populateConfig(settings)

	config.SSLCertificate = []byte(settings["CONJUR_SSL_CERTIFICATE"])

	usernameStr := settings["conjur.org/authn-identity"]
	if usernameStr == "" {
		usernameStr = settings["CONJUR_AUTHN_LOGIN"]
	}
	config.Username, _ = NewUsername(usernameStr)

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

func validConjurVersion(version string) error {
	validVersions := []string{"4", "5"}

	for _, validVersion := range validVersions {
		if version == validVersion {
			return nil
		}
	}

	return fmt.Errorf(log.CAKC021, "invalid conjur version")
}

func readSSLCert(readFile ReadFileFunc, settings map[string]string) ([]byte, error) {
	SSLCert := settings["CONJUR_SSL_CERTIFICATE"]
	SSLCertPath := settings["CONJUR_CERT_FILE"]
	if SSLCert == "" && SSLCertPath == "" {
		return nil, log.RecordedError(log.CAKC007)
	}

	if SSLCert != "" {
		return []byte(SSLCert), nil
	}
	return readFile(SSLCertPath)
}

func populateConfig(settings map[string]string) *Config {
	config := &Config{
		Account:       settings["CONJUR_ACCOUNT"],
		ContainerMode: settings["CONTAINER_MODE"],
		PodName:       settings["MY_POD_NAME"],
		PodNamespace:  settings["MY_POD_NAMESPACE"],
		URL:           settings["CONJUR_AUTHN_URL"],
	}

	annotContainerMode := settings["conjur.org/container-mode"]
	if annotContainerMode != "" {
		config.ContainerMode = annotContainerMode
	}

	// Only versions '4' & '5' are allowed, with '5' being used as the default
	config.ConjurVersion = DefaultConjurVersion
	if settings["CONJUR_VERSION"] != "" {
		config.ConjurVersion = settings["CONJUR_VERSION"]
	}

	config.TokenRefreshTimeout, _ = time.ParseDuration(settings["CONJUR_TOKEN_TIMEOUT"])

	config.TokenFilePath = DefaultTokenFilePath
	if envVal := settings["CONJUR_AUTHN_TOKEN_FILE"]; envVal != "" {
		config.TokenFilePath = envVal
	}

	config.ClientCertPath = DefaultClientCertPath
	if envVal := settings["CONJUR_CLIENT_CERT_PATH"]; envVal != "" {
		config.ClientCertPath = envVal
	}

	config.InjectCertLogPath = DefaultInjectCertLogPath

	config.ClientCertRetryCountLimit, _ = strconv.Atoi(settings["CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT"])

	return config
}

func logErrors(errLogs []error, infoLogs []error) {
	for _, err := range infoLogs {
		log.Info(err.Error())
	}
	for _, err := range errLogs {
		log.Error(err.Error())
	}
}
