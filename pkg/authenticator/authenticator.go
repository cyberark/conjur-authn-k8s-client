package authenticator

import (
	"context"
)

type Authenticator interface {
	Authenticate() error
	AuthenticateWithContext(ctx context.Context) error
}
