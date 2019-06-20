package config

import (
	"fmt"
	"os"
	"time"
)

// Config defines the configuration parameters
// for the authentication requests
type Config struct {
	PodNamespace    string
	TokenFilePath   string
	SecretsYamlPath string
	KubeSecretName  string
	SecretsDirPath  string
}

// DefaultTokenRefreshTimeout is the default time the system waits to
// reauthenticate on error
const DefaultTokenRefreshTimeout = 6 * time.Minute

// New returns a new authenticator configuration object
func NewFromEnv(tokenPath *string) (*Config, error) {
	var err error

	// Check that required environment variables are set
	for _, envvar := range []string{
		"CONJUR_APPLIANCE_URL",
		"CONJUR_ACCOUNT",
		"CONJUR_AUTHN_LOGIN",
		"CONJUR_AUTHN_URL",
		"MY_POD_NAMESPACE",
	} {
		if os.Getenv(envvar) == "" {
			err = fmt.Errorf("Environment variable %s must be provided", envvar)
			return nil, err
		}
	}

	// Load configuration from the environment
	podNamespace := os.Getenv("MY_POD_NAMESPACE")

	secretsYamlPath := os.Getenv("SECRETS_YML_PATH")
	if len(secretsYamlPath) == 0 {
		secretsYamlPath = "secrets.yml"
	}

	kubeSecretName := os.Getenv("K8S_SECRET_NAME")
	if len(kubeSecretName) == 0 {
		kubeSecretName = "dap-secret"
	}

	secretsDirPath := os.Getenv("SECRETS_DIR_PATH")
	if len(secretsDirPath) == 0 {
		secretsDirPath = "/run/conjur/secrets"
	}

	return &Config{
		PodNamespace:    podNamespace,
		TokenFilePath:   *tokenPath,
		SecretsYamlPath: secretsYamlPath,
		KubeSecretName:  kubeSecretName,
		SecretsDirPath:  secretsDirPath,
	}, nil
}
