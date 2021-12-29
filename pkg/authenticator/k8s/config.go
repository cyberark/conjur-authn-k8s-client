package k8s

import (
	"fmt"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// Config defines the configuration parameters
// for the authentication requests
type Config struct {
	Common            common.Config
	InjectCertLogPath string
	PodName           string
	PodNamespace      string
	ConjurVersion     string
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
	AuthnType                        = "authn-k8s"
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
}

func durationFromString(key, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf(log.CAKC060, key, value)
	}
	return duration, nil
}

func (config *Config) LoadConfig(settings map[string]string) {
	config.Common = common.Config{}
	config.Common.LoadConfig(settings)

	for key, value := range settings {
		switch key {
		case "MY_POD_NAME":
			config.PodName = value
		case "MY_POD_NAMESPACE":
			config.PodNamespace = value
		case "CONJUR_VERSION":
			config.ConjurVersion = value
		}
	}
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
