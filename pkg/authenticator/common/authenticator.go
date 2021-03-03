package common

import "github.com/cyberark/conjur-authn-k8s-client/pkg/config"

// Authenticator represents an authenticator interface
type Authenticator interface {
	Authenticate() error
	GlobalConfig() config.Config
	Init() (Authenticator, string)
	CanHandle(string) bool
}
