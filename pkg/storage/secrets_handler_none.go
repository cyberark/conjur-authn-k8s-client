package storage

import "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"


type SecretHandlerNone struct {
	AccessToken	AccessTokenHandler
	ConjurSecrets	ConjurSecretsFetcher
	K8sSecrets	SecretsStoreK8s
}

func NewSecretHandlerNone(config config.Config, AccessToken AccessTokenHandler) (SecretsHandler *SecretHandlerNone, err error){
	return
}

func (secretHandlerNone *SecretHandlerNone) HandleSecrets() error {
	return nil
}
