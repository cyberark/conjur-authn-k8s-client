package storage

import (
	"fmt"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/sidecar/logging"
	storageConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
)

type SecretsHandler interface {
	HandleSecrets() error
}

type AccessTokenHandler interface {
	Read() ([]byte, error)
	Write(Data []byte) error
	Delete() error
}

// TODO: understand why we need underscore here, otherwise "StorageHandler is not a type"
type Storage_Handler struct {
	AccessToken AccessTokenHandler
	Secrets     SecretsHandler
}

func NewStorageHandler(config storageConfig.Config) (StorageHandler *Storage_Handler, err error) {
	var infoLogger = log.InfoLogger

	// check what StoreType was in original store
	if config.StoreType == storageConfig.K8S {
		infoLogger.Printf(fmt.Sprintf("Storage configuration is %s ", storageConfig.K8S))

		// create new AccessTokenMemory object
		tokenInMemory, err := NewAccessTokenMemory(config)
		if err != nil {
			return nil, fmt.Errorf("error creating access token object, reason: %s", err)
		}
		// create new SecretHandler object
		secretsHandlerK8s, err := NewSecretHandlerK8s(config, tokenInMemory)
		if err != nil {
			return nil, fmt.Errorf("error secret handler object, reason: %s", err)
		}
		return &Storage_Handler{
			AccessToken: tokenInMemory,
			Secrets:     secretsHandlerK8s,
		}, nil
	} else {
		//	then the StoreType is none and original way of saving access token in file is run
		// check if initContainer or sidecar here or at the beginning?
		infoLogger.Printf(fmt.Sprintf("Storage configuration is %s ", storageConfig.None))

		tokenInFile, err := NewAccessTokenFile(config)
		if err != nil {
			return nil, fmt.Errorf("error creating access token object, reason: %s", err)
		}
		secretsHandlerNone, err := NewSecretHandlerNone(config, tokenInFile)
		if err != nil {
			return nil, fmt.Errorf("error secret handler object, reason: %s", err)
		}
		return &Storage_Handler{
			AccessToken: tokenInFile,
			Secrets:     secretsHandlerNone,
		}, nil
	}
}
