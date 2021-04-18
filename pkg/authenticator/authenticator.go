package authenticator

import (
	"fmt"
	"os"
	"strings"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
	gcpAuthenticator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/gcp"

	k8sAuthenticator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/k8s"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

const (
	// DefaultAuthnType defaults to k8s authentication
	DefaultAuthnType = k8sAuthenticator.AuthnType
	// AuthnTypeEnvironmentVariable environment variable set for authentication type
	AuthnTypeEnvironmentVariable = "CONJUR_AUTHN_TYPE"
)

// NewAuthenticatorFromEnv returns desired authenticator type
func NewAuthenticatorFromEnv() (common.Authenticator, string) {
	configureLogLevel()

	authnStategies := registeredAuthenticators()
	authnType := getAuthnType()

	for _, authnStategy := range authnStategies {
		if authnStategy.CanHandle(authnType) {
			return authnStategy.Init()
		}
	}

	return nil, fmt.Sprintf(log.CAKC060, authnType)
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

func registeredAuthenticators() []common.Authenticator {
	return []common.Authenticator{
		&gcpAuthenticator.Authenticator{},
		&k8sAuthenticator.Authenticator{},
	}
}

func getAuthnType() string {
	authnType := os.Getenv(AuthnTypeEnvironmentVariable)
	if authnType == "" {
		return DefaultAuthnType
	}
	return strings.ToLower(authnType)
}
