package config

import (
	"fmt"
	"io/ioutil"
	"os"
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
	"CONJUR_AUTHN_URL",
	"CONJUR_ACCOUNT",
	"CONJUR_AUTHN_LOGIN",
	"MY_POD_NAMESPACE",
	"MY_POD_NAME",
}

// ReadFileFunc defines the interface for reading an SSL Certificate from the env
type ReadFileFunc func(filename string) ([]byte, error)

// NewFromEnv returns a config FromEnv using the standard file reader for reading certs
func NewFromEnv() (*Config, error) {
	return FromEnv(ioutil.ReadFile)
}

// FromEnv returns a new authenticator configuration object
func FromEnv(readFileFunc ReadFileFunc) (*Config, error) {
	var err error

	configureLogLevel()

	// Fill config with 'simple' values from environment
	config, err := populateConfig()
	if err != nil {
		return nil, err
	}

	// Load CA cert from Environment
	config.SSLCertificate, err = readSSLCert(readFileFunc)
	if err != nil {
		return nil, log.RecordedError(log.CAKC021, err)
	}

	// Load Username from Environment
	config.Username, err = NewUsername(os.Getenv("CONJUR_AUTHN_LOGIN"))
	if err != nil {
		return nil, err
	}

	return config, nil
}

func configureLogLevel() {
	validVal := "true"
	val := os.Getenv("DEBUG")
	if val == validVal {
		log.EnableDebugMode()
	} else if val != "" {
		// In case "DEBUG" is configured with incorrect value
		log.Warn(log.CAKC034, val, validVal)
	}
}

func readSSLCert(readFile ReadFileFunc) ([]byte, error) {
	SSLCert := os.Getenv("CONJUR_SSL_CERTIFICATE")
	SSLCertPath := os.Getenv("CONJUR_CERT_FILE")
	if SSLCert == "" && SSLCertPath == "" {
		return nil, log.RecordedError(log.CAKC007)
	}

	if SSLCert != "" {
		return []byte(SSLCert), nil
	}
	return readFile(SSLCertPath)
}

func populateConfig() (*Config, error) {
	// Check that required environment variables are set
	for _, envvar := range requiredEnvVariables {
		if os.Getenv(envvar) == "" {
			return nil, log.RecordedError(log.CAKC009, envvar)
		}
	}

	config := &Config{
		Account:       os.Getenv("CONJUR_ACCOUNT"),
		ContainerMode: os.Getenv("CONTAINER_MODE"),
		PodName:       os.Getenv("MY_POD_NAME"),
		PodNamespace:  os.Getenv("MY_POD_NAMESPACE"),
		URL:           os.Getenv("CONJUR_AUTHN_URL"),
	}

	// Only versions '4' & '5' are allowed, with '5' being used as the default
	config.ConjurVersion = DefaultConjurVersion
	switch os.Getenv("CONJUR_VERSION") {
	case "4":
		config.ConjurVersion = "4"
	case "5":
		break // Stick with default
	case "":
		break // Stick with default
	default:
		return nil, log.RecordedError(log.CAKC021, fmt.Errorf("invalid conjur version"))
	}

	// Parse token refresh rate if one is provided from env
	tokenRefreshTimeout, err := utils.DurationFromEnvOrDefault(
		"CONJUR_TOKEN_TIMEOUT",
		DefaultTokenRefreshTimeout,
		nil,
	)
	if err != nil {
		return nil, err
	}
	config.TokenRefreshTimeout = tokenRefreshTimeout

	config.TokenFilePath = DefaultTokenFilePath
	// If CONJUR_TOKEN_FILE_PATH is defined in the env we take its value
	if envVal := os.Getenv("CONJUR_AUTHN_TOKEN_FILE"); envVal != "" {
		config.TokenFilePath = envVal
	}

	config.ClientCertPath = DefaultClientCertPath
	// If CONJUR_CLIENT_CERT_PATH is defined in the env we take its value
	if envVal := os.Getenv("CONJUR_CLIENT_CERT_PATH"); envVal != "" {
		config.ClientCertPath = envVal
	}

	config.InjectCertLogPath = DefaultInjectCertLogPath

	// Parse client cert retry count limit if one is provided from env
	clientCertRetryCountLimit, err := utils.IntFromEnvOrDefault(
		"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT",
		DefaultClientCertRetryCountLimit,
		nil,
	)
	if err != nil {
		return nil, err
	}
	config.ClientCertRetryCountLimit = clientCertRetryCountLimit

	return config, nil
}
