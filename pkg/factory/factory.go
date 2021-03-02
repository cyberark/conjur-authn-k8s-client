package factory

import (
	"os"
	"strings"

	gcpAuthenticator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/gcp"
	gcpAuthnConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/gcp/config"
	k8sAuthenticator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/k8s"
	k8sAuthnConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/k8s/config"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/config"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

const (
	// K8sAuthnType is k8s
	K8sAuthnType = "k8s"
	// GCPAuthnType is gcp
	GCPAuthnType = "gcp"
	// DefaultAuthnType defaults to k8s authentication
	DefaultAuthnType = K8sAuthnType
	// AuthnTypeEnvironmentVariable environment variable set for authentication type
	AuthnTypeEnvironmentVariable = "CONJUR_AUTHN_TYPE"
)

// Authenticator represents an authenticator interface
type Authenticator interface {
	Authenticate() error
	GlobalConfig() config.Config
}

// NewAuthenticatorFromEnv returns desired authenticator type
func NewAuthenticatorFromEnv() (Authenticator, string) {
	var authn Authenticator
	var errMsg string
	configureLogLevel()

	authnType := getAuthnType(os.Getenv(AuthnTypeEnvironmentVariable), DefaultAuthnType)

	switch authnType {
	case K8sAuthnType:
		authn, errMsg = initK8s()
	case GCPAuthnType:
		authn, errMsg = initGCP()
	}

	return authn, errMsg
}

func getAuthnType(authnType string, defaultType string) string {
	if strings.ToLower(authnType) == GCPAuthnType {
		return GCPAuthnType
	}
	return K8sAuthnType
}

func initK8s() (Authenticator, string) {
	log.Debug(log.CAKC058)
	config, err := k8sAuthnConfig.NewFromEnv()
	if err != nil {
		return nil, log.CAKC018
	}

	authn, err := k8sAuthenticator.New(*config)
	if err != nil {
		return nil, log.CAKC019
	}

	return authn, ""

}

func initGCP() (Authenticator, string) {
	log.Debug(log.CAKC059)
	config, err := gcpAuthnConfig.NewFromEnv()
	if err != nil {
		return nil, log.CAKC018
	}

	authn, err := gcpAuthenticator.New(*config)
	if err != nil {
		return nil, log.CAKC019
	}

	return authn, ""
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
