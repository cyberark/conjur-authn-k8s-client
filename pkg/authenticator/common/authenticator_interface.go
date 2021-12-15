package common

import (
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
)

type AuthenticatorInterface interface {
	Init(config *ConfigurationInterface) (AuthenticatorInterface, error)
	InitWithAccessToken(config *ConfigurationInterface, token access_token.AccessToken) (AuthenticatorInterface, error)
	Authenticate() error
	GetAccessToken() access_token.AccessToken
}
