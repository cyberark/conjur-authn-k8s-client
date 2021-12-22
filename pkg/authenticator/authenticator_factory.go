package authenticator

import (
	"fmt"
	"net/http"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

func NewAuthenticator(conf config.ConfigurationInterface) (AuthenticatorInterface, error) {
	authn, client, error := getAuthenticator(conf)
	if error != nil {
		return nil, error
	}
	return authn.Init(&conf, client)
}

func NewAuthenticatorWithAccessToken(conf config.ConfigurationInterface,
	token access_token.AccessToken) (AuthenticatorInterface, error) {

	var error error
	var authn AuthenticatorInterface
	authn, error = getAuthenticator(conf)
	if error != nil {
		return nil, error
	}
	return authn.InitWithAccessToken(&conf, client, token)
}

func getAuthenticator(conf config.ConfigurationInterface) (AuthenticatorInterface,
	*http.Client, error) {

	var authenticator AuthenticatorInterface

	// Get authenticator based on configured authenticator type
	switch conf.GetAuthenticationType() {
	case k8sAuthenticator.AuthnType:
		authenticator = &k8sAuthenticator.Authenticator{}
	default:
		return nil, nil, fmt.Errorf(log.CAKC064)
	}

	// Create an HTTP client
	client, err := newHTTPSClient(config.Common.SSLCertificate, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	return authenticator, client, nil
}
