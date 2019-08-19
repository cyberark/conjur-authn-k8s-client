package storage

import (
	"github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
)

type SecretsHandler interface {
	HandleSecrets() error
}

type AccessTokenHandler interface {
	Read()	([]byte, error)
	Write()	error
	Delete()	error
}

// Need underscore here unsure why
type Storage_Handler struct {
	AccessToken AccessTokenHandler
	Secrets SecretsHandler
}

func NewStorageHandler(config config.Config) (StorageHandler *Storage_Handler, err error) {
	// check what StoreType was in original store
	if config.StoreType == "k8s_secrets" {
		// create new AccessTokenMemory object
		tokenInMemory, err := NewAccessTokenMemory(config)
		// create new SecretHandler object
		secretsHandlerK8s, err := NewSecretHandlerK8s(config, tokenInMemory)
		return &Storage_Handler{
			AccessToken: tokenInMemory,
			Secrets:     secretsHandlerK8s,
		}, nil
	} else {
		//	then the StoreType is none and old way of saving access token in file is run
		// check if initContainer or sidecar here or at the beginning?
		tokenInFile, err := NewAccessTokenFile(config)
		secretsHandlerNone, err := NewSecretHandlerNone(config, tokenInFile)
		return &Storage_Handler{
			AccessToken: tokenInFile,
			Secrets:     secretsHandlerNone,
		}, nil
	}
}