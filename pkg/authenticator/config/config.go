package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// Config defines the configuration parameters
// for the authentication requests
type Config struct {
	ContainerMode       string
	ConjurVersion       string
	Account             string
	URL                 string
	Username            string
	PodName             string
	PodNamespace        string
	SSLCertificate      []byte
	ClientCertPath      string
	TokenRefreshTimeout time.Duration
}

// DefaultTokenRefreshTimeout is the default time the system waits to
// reauthenticate on error
const (
	DefaultTokenRefreshTimeout = 6 * time.Minute
	ClientCertPathDefault      = "/etc/conjur/ssl/client.pem"
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
			err = fmt.Errorf("Environment variable %s must be provided", envvar)
			return nil, err
		}
	}

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

	clientCertPath := ClientCertPathDefault
	// If CONJUR_CLIENT_CERT_PATH not configured take default value
	if envVal := os.Getenv("CONJUR_CLIENT_CERT_PATH"); envVal != "" {
		clientCertPath = envVal
	}

	return &Config{
		ContainerMode:       containerMode,
		ConjurVersion:       conjurVersion,
		Account:             account,
		URL:                 authnURL,
		Username:            authnLogin,
		PodName:             podName,
		PodNamespace:        podNamespace,
		SSLCertificate:      caCert,
		ClientCertPath:      clientCertPath,
		TokenRefreshTimeout: tokenRefreshTimeout,
	}, nil
}

func readSSLCert() ([]byte, error) {
	SSLCert := os.Getenv("CONJUR_SSL_CERTIFICATE")
	SSLCertPath := os.Getenv("CONJUR_CERT_FILE")
	if SSLCert == "" && SSLCertPath == "" {
		err := fmt.Errorf(
			"At least one of CONJUR_SSL_CERTIFICATE and CONJUR_CERT_FILE must be provided")
		return nil, err
	}

	if SSLCert != "" {
		return []byte(SSLCert), nil
	}
	return ioutil.ReadFile(SSLCertPath)
}
