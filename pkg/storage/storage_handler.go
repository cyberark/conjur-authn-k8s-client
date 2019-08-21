package storage

import (
	"os"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/sidecar/logging"
)

type SecretsHandler interface {
	HandleSecrets() error
}

type AccessTokenHandler interface {
	Read()	([]byte, error)
	Write()	error
	Delete()	error
}

// TODO: understand why we need underscore here, otherwise "StorageHandler is not a type"
type Storage_Handler struct {
	AccessToken AccessTokenHandler
	Secrets SecretsHandler
}

func NewStorageHandler(config config.Config) (StorageHandler *Storage_Handler, err error) {
	var errLogger = log.ErrorLogger

	// check what StoreType was in original store
	if config.StoreType == "k8s" {
		// create new AccessTokenMemory object
		tokenInMemory, err := NewAccessTokenMemory(config)
		if err != nil {
			errLogger.Printf(err.Error())
			os.Exit(1)
		}
		// create new SecretHandler object
		secretsHandlerK8s, err := NewSecretHandlerK8s(config, tokenInMemory)
		if err != nil {
			errLogger.Printf(err.Error())
			os.Exit(1)
		}
		return &Storage_Handler{
			AccessToken: tokenInMemory,
			Secrets:     secretsHandlerK8s,
		}, nil
	} else {
		//	then the StoreType is none and original way of saving access token in file is run
		// check if initContainer or sidecar here or at the beginning?
		tokenInFile, err := NewAccessTokenFile(config)
		if err != nil {
			errLogger.Printf(err.Error())
			os.Exit(1)
		}
		secretsHandlerNone, err := NewSecretHandlerNone(config, tokenInFile)
		if err != nil {
			errLogger.Printf(err.Error())
			os.Exit(1)
		}
		return &Storage_Handler{
			AccessToken: tokenInFile,
			Secrets:     secretsHandlerNone,
		}, nil
	}
}