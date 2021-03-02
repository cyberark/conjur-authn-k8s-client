package config

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/utils"
)

// Config defines the configuration parameters
// for the authentication requests
type Config struct {
	Account             string
	ContainerMode       string
	SSLCertificate      []byte
	TokenFilePath       string
	TokenRefreshTimeout time.Duration
	URL                 string
	Username            string
}

// Default settings (this comment added to satisfy linter)
const (
	DefaultTokenFilePath = "/run/conjur/access-token"

	// DefaultTokenRefreshTimeout is the default time the system waits to reauthenticate on error
	DefaultTokenRefreshTimeout = "6m0s"
)

var requiredEnvVariables = []string{
	"CONJUR_AUTHN_URL",
	"CONJUR_ACCOUNT",
	"CONJUR_AUTHN_LOGIN",
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

	return config, nil
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
		URL:           os.Getenv("CONJUR_AUTHN_URL"),
		Username:      os.Getenv("CONJUR_AUTHN_LOGIN"),
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

	// Load Username from Environment
	config.Username = os.Getenv("CONJUR_AUTHN_LOGIN")

	return config, nil
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
