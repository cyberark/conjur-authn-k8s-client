package authenticator

import (
	"fmt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
	k8sAuthenitcator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/k8s"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

func NewAuthenticator(conf common.ConfigurationInterface) (common.AuthenticatorInterface, error) {
	authn, error := getAuthenticator(conf)
	if error != nil {
		return nil, error
	}
	return authn.Init(&conf)
}

func NewAuthenticatorWithAccessToken(conf common.ConfigurationInterface, token access_token.AccessToken) (common.AuthenticatorInterface, error) {
	var error error
	var authn common.AuthenticatorInterface
	authn, error = getAuthenticator(conf)
	if error != nil {
		return nil, error
	}
	return authn.InitWithAccessToken(&conf, token)
}

func getAuthenticator(conf common.ConfigurationInterface) (common.AuthenticatorInterface, error) {
	if conf.GetAuthenticationType() == k8sAuthenitcator.AuthnType {
		return &k8sAuthenitcator.Authenticator{}, nil
	}
	return nil, fmt.Errorf(log.CAKC064)
}
