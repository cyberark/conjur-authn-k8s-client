package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// Config defines the configuration parameters
// for the authentication requests
type Config struct {
	Account             string
	ClientCertPath      string
	ContainerMode       string
	ConjurVersion       string
	PodName             string
	PodNamespace        string
	SSLCertificate      []byte
	TokenFilePath       string
	TokenRefreshTimeout time.Duration
	URL                 string
	Username            string
}

// DefaultTokenRefreshTimeout is the default time the system waits to
// reauthenticate on error
const (
	ClientCertPathDefault      = "/etc/conjur/ssl/client.pem"
	DefaultTokenRefreshTimeout = 6 * time.Minute
	TokenFilePathDefault       = "/run/conjur/access-token"
)

// New returns a new authenticator configuration object
func NewFromEnv() (*Config, error) {
	var err error

	// Check that required environment variables are set
	for _, envvar := range []string{
		"CONJUR_AUTHN_URL",
		"CONJUR_ACCOUNT",
		"CONJUR_AUTHN_LOGIN",
		"MY_POD_NAMESPACE",
		"MY_POD_NAME",
	} {
		if os.Getenv(envvar) == "" {
			err = fmt.Errorf("environment variable %s must be provided", envvar)
			return nil, err
		}
	}

	// Read flags
	tokenFilePath := flag.String("t", TokenFilePathDefault,
		"Path to Conjur access token")
	clientCertPath := flag.String("c", ClientCertPathDefault,
		"Path to client certificate")
	flag.Parse()

	// Load CA cert
	caCert, err := readSSLCert()
	if err != nil {
		return nil, err
	}

	// Load configuration from the environment
	authnURL := os.Getenv("CONJUR_AUTHN_URL")
	account := os.Getenv("CONJUR_ACCOUNT")
	authnLogin := os.Getenv("CONJUR_AUTHN_LOGIN")
	podNamespace := os.Getenv("MY_POD_NAMESPACE")
	podName := os.Getenv("MY_POD_NAME")

	containerMode := os.Getenv("CONTAINER_MODE")

	conjurVersion := os.Getenv("CONJUR_VERSION")
	if len(conjurVersion) == 0 {
		conjurVersion = "5"
	}

	// Parse token refresh rate if one is provided from env
	tokenRefreshTimeout := DefaultTokenRefreshTimeout
	tokenRefreshTimeoutString := os.Getenv("CONJUR_TOKEN_TIMEOUT")
	if len(tokenRefreshTimeoutString) > 0 {
		parsedTokenRefreshTimeout, err := time.ParseDuration(tokenRefreshTimeoutString)
		if err != nil {
			return nil, err
		}

		tokenRefreshTimeout = parsedTokenRefreshTimeout
	}

	// If CONJUR_TOKEN_FILE_PATH is defined in the env we take its value
	if envVal := os.Getenv("CONJUR_AUTHN_TOKEN_FILE"); envVal != "" {
		tokenFilePath = &envVal
	}

	// If CONJUR_CLIENT_CERT_PATH is defined in the env we take its value
	if envVal := os.Getenv("CONJUR_CLIENT_CERT_PATH"); envVal != "" {
		clientCertPath = &envVal
	}

	return &Config{
		Account:             account,
		ClientCertPath:      *clientCertPath,
		ContainerMode:       containerMode,
		ConjurVersion:       conjurVersion,
		PodName:             podName,
		PodNamespace:        podNamespace,
		SSLCertificate:      caCert,
		TokenFilePath:       *tokenFilePath,
		TokenRefreshTimeout: tokenRefreshTimeout,
		URL:                 authnURL,
		Username:            authnLogin,
	}, nil
}

func readSSLCert() ([]byte, error) {
	SSLCert := os.Getenv("CONJUR_SSL_CERTIFICATE")
	SSLCertPath := os.Getenv("CONJUR_CERT_FILE")
	if SSLCert == "" && SSLCertPath == "" {
		err := fmt.Errorf(
			"at least one of CONJUR_SSL_CERTIFICATE and CONJUR_CERT_FILE must be provided")
		return nil, err
	}

	if SSLCert != "" {
		return []byte(SSLCert), nil
	}
	return ioutil.ReadFile(SSLCertPath)
}
