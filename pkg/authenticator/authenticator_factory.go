package authenticator

import (
	"fmt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token/file"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	jwtAuthenticator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/jwt"
	k8sAuthenticator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/k8s"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// NewAuthenticator creates an instance of the Authenticator interface based on configured authenticator type.
func NewAuthenticator(conf config.Configuration) (Authenticator, error) {
	accessToken, error := file.NewAccessToken(conf.GetTokenFilePath())
	if error != nil {
		return nil, error
	}
	return getAuthenticator(conf, accessToken)
}

// NewAuthenticatorWithAccessToken creates an instance of the Authenticator interface based on configured authenticator type
// and access token
func NewAuthenticatorWithAccessToken(conf config.Configuration, token access_token.AccessToken) (Authenticator, error) {
	return getAuthenticator(conf, token)
}

func getAuthenticator(conf config.Configuration, token access_token.AccessToken) (Authenticator, error) {
	switch c := conf.(type) {
	case *k8sAuthenticator.Config:
		return k8sAuthenticator.NewWithAccessToken(*c, token)
	case *jwtAuthenticator.Config:
		return jwtAuthenticator.NewWithAccessToken(*c, token)
	default:
		return nil, fmt.Errorf(log.CAKC064)
	}
}
