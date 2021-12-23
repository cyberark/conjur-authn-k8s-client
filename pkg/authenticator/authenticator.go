package authenticator

import (
	"context"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
)

type Authenticator interface {
	Authenticate() error
	AuthenticateWithContext(ctx context.Context) error
	GetAccessToken() access_token.AccessToken
}
