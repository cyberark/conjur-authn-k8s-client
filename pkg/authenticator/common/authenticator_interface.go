package common

import (
	"context"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
)

type AuthenticatorInterface interface {
	Init(config *ConfigurationInterface) (AuthenticatorInterface, error)
	InitWithAccessToken(config *ConfigurationInterface, token access_token.AccessToken) (AuthenticatorInterface, error)
	Authenticate() error
	AuthenticateWithContext(ctx context.Context) error
	GetAccessToken() access_token.AccessToken
}
