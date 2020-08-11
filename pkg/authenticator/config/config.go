package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// Config defines the configuration parameters
// for the authentication requests
type Config struct {
	Account                     string
	ClientCertPath              string
	ContainerMode               string
	SupportMultipleApplications string
	ConjurVersion               string
	PodName                     string
	PodNamespace                string
	SSLCertificate              []byte
	TokenFilePath               string
	TokenRefreshTimeout         time.Duration
	URL                         string
	Username                    *Username
}

// Default settings (this comment added to satisfy linter)
const (
	DefaultClientCertPath              = "/etc/conjur/ssl/client.pem"
	DefaultTokenFilePath               = "/run/conjur/access-token"
	DefaultContainerMode               = "sidecar"
	DefaultSupportMultipleApplications = "false"
	DefaultConjurVersion               = "5"

	// DefaultTokenRefreshTimeout is the default time the system waits to reauthenticate on error
	DefaultTokenRefreshTimeout = 6 * time.Minute
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

	// Fill config with 'simple' values from environment
	config, err := populateConfig()
	if err != nil {
		return nil, err
	}

	// Load CA cert from Environment
	config.SSLCertificate, err = readSSLCert(readFileFunc)
	if err != nil {
		return nil, log.RecordedError(log.CAKC021E, err.Error())
	}

	// Load Username from Environment
	config.Username, err = NewUsername(os.Getenv("CONJUR_AUTHN_LOGIN"))
	if err != nil {
		return nil, err
	}

	return config, nil
}

func readSSLCert(readFile ReadFileFunc) ([]byte, error) {
	SSLCert := os.Getenv("CONJUR_SSL_CERTIFICATE")
	SSLCertPath := os.Getenv("CONJUR_CERT_FILE")
	if SSLCert == "" && SSLCertPath == "" {
		return nil, log.RecordedError(log.CAKC007E)
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
			return nil, log.RecordedError(log.CAKC009E, envvar)
		}
	}

	defaultConfig := &Config{
		Account:      os.Getenv("CONJUR_ACCOUNT"),
		PodName:      os.Getenv("MY_POD_NAME"),
		PodNamespace: os.Getenv("MY_POD_NAMESPACE"),
		URL:          os.Getenv("CONJUR_AUTHN_URL"),
	}

	// Only versions '4' & '5' are allowed, with '5' being used as the default
	defaultConfig.ConjurVersion = DefaultConjurVersion
	switch os.Getenv("CONJUR_VERSION") {
	case "4":
		defaultConfig.ConjurVersion = "4"
	case "5":
		break // Stick with default
	case "":
		break // Stick with default
	default:
		return nil, log.RecordedError(log.CAKC021E, fmt.Errorf("invalid conjur version"))
	}

	// Parse token refresh rate if one is provided from env
	defaultConfig.TokenRefreshTimeout = DefaultTokenRefreshTimeout
	tokenRefreshTimeoutString := os.Getenv("CONJUR_TOKEN_TIMEOUT")
	if len(tokenRefreshTimeoutString) > 0 {
		parsedTokenRefreshTimeout, err := time.ParseDuration(tokenRefreshTimeoutString)
		if err != nil {
			return nil, log.RecordedError(log.CAKC010E, err.Error())
		}

		defaultConfig.TokenRefreshTimeout = parsedTokenRefreshTimeout
	}

	defaultConfig.TokenFilePath = DefaultTokenFilePath
	// If CONJUR_TOKEN_FILE_PATH is defined in the env we take its value
	if envVal := os.Getenv("CONJUR_AUTHN_TOKEN_FILE"); envVal != "" {
		defaultConfig.TokenFilePath = envVal
	}

	defaultConfig.ClientCertPath = DefaultClientCertPath
	// If CONJUR_CLIENT_CERT_PATH is defined in the env we take its value
	if envVal := os.Getenv("CONJUR_CLIENT_CERT_PATH"); envVal != "" {
		defaultConfig.ClientCertPath = envVal
	}

	defaultConfig.ContainerMode = DefaultContainerMode
	// If CONTAINER_MODE is defined in the env we take its value
	if envVal := os.Getenv("CONTAINER_MODE"); envVal != "" {
		defaultConfig.ContainerMode = os.Getenv("CONTAINER_MODE")
	}

	defaultConfig.SupportMultipleApplications = DefaultSupportMultipleApplications
	// If SUPPORT_MULTIPLE_APPS is defined in the env we take its value
	if envVal := os.Getenv("SUPPORT_MULTIPLE_APPS"); envVal != "" {
		defaultConfig.SupportMultipleApplications = os.Getenv("SUPPORT_MULTIPLE_APPS")
	}

	return defaultConfig, nil
}
