package storage

import (
	"github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
)

type SecretsFetcher interface {
}

type SecretsStoreK8s interface {
}

type SecretHandlerK8s struct {
	AccessToken	*AccessTokenHandler
	ConjurSecrets	*SecretsFetcher
	K8sSecrets	*SecretsStoreK8s
}

func NewSecretHandlerK8s(config config.Config, AccessToken AccessTokenHandler) (SecretsHandler *SecretHandlerK8s, err error){
	return
}

func (secrets *SecretHandlerK8s) HandleSecrets() error {
	return nil
}
