package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config defines the configuration parameters
// for the authentication requests
type Config struct {
	PodNamespace       string
	RequiredK8sSecrets []string
}

// DefaultTokenRefreshTimeout is the default time the system waits to
// reauthenticate on error
const DefaultTokenRefreshTimeout = 6 * time.Minute

const CONJUR_MAP_KEY = "conjur-map"

// New returns a new authenticator configuration object
func NewFromEnv() (*Config, error) {
	var err error

	// Check that required environment variables are set
	for _, envvar := range []string{
		"CONJUR_APPLIANCE_URL",
		"CONJUR_ACCOUNT",
		"CONJUR_AUTHN_LOGIN",
		"CONJUR_AUTHN_URL",
		"MY_POD_NAMESPACE",
		"K8S_SECRETS",
	} {
		if os.Getenv(envvar) == "" {
			err = fmt.Errorf("environment variable %s must be provided", envvar)
			return nil, err
		}
	}

	// Load configuration from the environment
	podNamespace := os.Getenv("MY_POD_NAMESPACE")

	// Split the comma-separated list into an array
	requiredK8sSecrets := strings.Split(os.Getenv("K8S_SECRETS"), ",")

	return &Config{
		PodNamespace:       podNamespace,
		RequiredK8sSecrets: requiredK8sSecrets,
	}, nil
}
