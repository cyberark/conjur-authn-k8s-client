package authenticator

import (
	"context"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-opentelemetry-tracer/pkg/trace"
)

type Authenticator interface {
	Authenticate() error
	AuthenticateWithContext(ctx context.Context, tracer trace.Tracer) error
	GetAccessToken() access_token.AccessToken
}
