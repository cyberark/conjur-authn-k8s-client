package creators

import (
	"fmt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	jwtAuthenitcator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/jwt"
	"strings"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
	k8sAuthenitcator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/k8s"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

func NewAuthenticator(conf common.ConfigurationInterface) (common.AuthenticatorInterface, error) {
	authn, error := GetAuthenticator(conf)
	if error != nil {
		return nil, error
	}
	return authn.Init(&conf)
}

func NewAuthenticatorWithAccessToken(conf common.ConfigurationInterface, token access_token.AccessToken) (common.AuthenticatorInterface, error) {
	var error error
	var authn common.AuthenticatorInterface
	authn, error = GetAuthenticator(conf)
	if error != nil {
		return nil, error
	}
	return authn.InitWithAccessToken(&conf, token)
}

func GetAuthenticator(conf common.ConfigurationInterface) (common.AuthenticatorInterface, error) {
	if strings.Compare(conf.GetAuthenticationType(), k8sAuthenitcator.AuthnType) == 0 {
		return &k8sAuthenitcator.Authenticator{}, nil
	} else if strings.Compare(conf.GetAuthenticationType(), jwtAuthenitcator.AuthnType) == 0 {
		return &jwtAuthenitcator.Authenticator{}, nil
	}
	return nil, fmt.Errorf(log.CAKC064)
}
