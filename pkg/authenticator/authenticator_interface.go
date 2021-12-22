package authenticator

import (
	"context"
	"net/http"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
)

type AuthenticatorInterface interface {
	Init(config *ConfigurationInterface, client *http.Client) (AuthenticatorInterface, error)
	InitWithAccessToken(config *ConfigurationInterface, client *http.Client,
		token access_token.AccessToken) (AuthenticatorInterface, error)
	Authenticate() error
	AuthenticateWithContext(ctx context.Context) error
	GetAccessToken() access_token.AccessToken
}
