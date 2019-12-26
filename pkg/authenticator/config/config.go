package config

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
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
	Username            *Username
}

const (
	DefaultClientCertPath = "/etc/conjur/ssl/client.pem"
	DefaultTokenFilePath  = "/run/conjur/access-token"

	// DefaultTokenRefreshTimeout is the default time the system waits to reauthenticate on error
	DefaultTokenRefreshTimeout = 6 * time.Minute
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
			return nil, log.RecordedError(log.CAKC009E, envvar)
		}
	}

	// Load CA cert
	caCert, err := readSSLCert()
	if err != nil {
		return nil, log.RecordedError(log.CAKC021E, err.Error())
	}

	// Load configuration from the environment
	authnURL := os.Getenv("CONJUR_AUTHN_URL")
	account := os.Getenv("CONJUR_ACCOUNT")
	authnLogin, err := NewUsername(os.Getenv("CONJUR_AUTHN_LOGIN"))
	if err != nil {
		return nil, err
	}

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
			return nil, log.RecordedError(log.CAKC010E, err.Error())
		}

		tokenRefreshTimeout = parsedTokenRefreshTimeout
	}

	tokenFilePath := DefaultTokenFilePath
	// If CONJUR_TOKEN_FILE_PATH is defined in the env we take its value
	if envVal := os.Getenv("CONJUR_AUTHN_TOKEN_FILE"); envVal != "" {
		tokenFilePath = envVal
	}

	clientCertPath := DefaultClientCertPath
	// If CONJUR_CLIENT_CERT_PATH is defined in the env we take its value
	if envVal := os.Getenv("CONJUR_CLIENT_CERT_PATH"); envVal != "" {
		clientCertPath = envVal
	}

	return &Config{
		Account:             account,
		ClientCertPath:      clientCertPath,
		ContainerMode:       containerMode,
		ConjurVersion:       conjurVersion,
		PodName:             podName,
		PodNamespace:        podNamespace,
		SSLCertificate:      caCert,
		TokenFilePath:       tokenFilePath,
		TokenRefreshTimeout: tokenRefreshTimeout,
		URL:                 authnURL,
		Username:            authnLogin,
	}, nil
}

func readSSLCert() ([]byte, error) {
	SSLCert := os.Getenv("CONJUR_SSL_CERTIFICATE")
	SSLCertPath := os.Getenv("CONJUR_CERT_FILE")
	if SSLCert == "" && SSLCertPath == "" {
		return nil, log.RecordedError(log.CAKC007E)
	}

	if SSLCert != "" {
		return []byte(SSLCert), nil
	}
	return ioutil.ReadFile(SSLCertPath)
}
