package jwt

import (
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
)

// Config defines the configuration parameters
// for the authentication requests
type Config struct {
	Common           common.Config
	JWTTokenFilePath string
}

// Default settings (this comment added to satisfy linter)
const (
	DefaultClientCertPath = "/etc/conjur/ssl/client.pem"
	DefaultTokenFilePath  = "/run/conjur/access-token"

	// DefaultTokenRefreshTimeout is the default time the system waits to reauthenticate on error
	DefaultTokenRefreshTimeout = "6m0s"

	// DefaultClientCertRetryCountLimit is the amount of times we wait after successful
	// login for the client certificate file to exist, where each time we wait for a second.
	DefaultClientCertRetryCountLimit = "10"

	DefaultJWTTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

	AuthnType = "authn-jwt"
)

var requiredEnvVariables = []string{
	"CONJUR_AUTHN_URL",
	"CONJUR_ACCOUNT",
}

var envVariables = []string{
	"CONJUR_ACCOUNT",
	"CONJUR_AUTHN_TOKEN_FILE",
	"CONJUR_AUTHN_URL",
	"CONJUR_CERT_FILE",
	"CONJUR_SSL_CERTIFICATE",
	"CONJUR_TOKEN_TIMEOUT",
	"CONTAINER_MODE",
	"DEBUG",
	"JWT_TOKEN_PATH",
	"CONJUR_AUTHN_LOGIN",
}

var defaultValues = map[string]string{
	"CONJUR_CLIENT_CERT_PATH":              DefaultClientCertPath,
	"CONJUR_AUTHN_TOKEN_FILE":              DefaultTokenFilePath,
	"CONJUR_TOKEN_TIMEOUT":                 DefaultTokenRefreshTimeout,
	"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": DefaultClientCertRetryCountLimit,
	"JWT_TOKEN_PATH":                       DefaultJWTTokenPath,
}

func (config *Config) LoadConfig(settings map[string]string) {
	config.Common = common.Config{}
	config.Common.LoadConfig(settings)

	if path, exists := settings["JWT_TOKEN_PATH"]; exists {
		config.JWTTokenFilePath = path
	}
}

func (config *Config) GetAuthenticationType() string {
	return AuthnType
}

func (config *Config) GetEnvVariables() []string {
	return envVariables
}

func (config *Config) GetRequiredVariables() []string {
	return requiredEnvVariables
}

func (config *Config) GetDefaultValues() map[string]string {
	return defaultValues
}

func (config *Config) GetContainerMode() string {
	return config.Common.ContainerMode
}

func (config *Config) GetTokenFilePath() string {
	return config.Common.TokenFilePath
}

func (config *Config) GetTokenTimeout() time.Duration {
	return config.Common.TokenRefreshTimeout
}
