package storage

import (
	"fmt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	secretsConfigProvider "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/config"
	secretsHandlers "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/handlers"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/sidecar/logging"
	storageConfigProvider "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
)

// TODO: understand why we need underscore here, otherwise "Storage_Handler is not a type"
type Storage_Handler struct {
	AccessTokenHandler access_token.AccessTokenHandler
	SecretsHandler     secretsHandlers.SecretsHandler
}

func NewStorageHandler(storageConfig storageConfigProvider.Config) (StorageHandler *Storage_Handler, err error) {
	var errLogger = log.ErrorLogger
	var infoLogger = log.InfoLogger

	secretsConfig, err := secretsConfigProvider.NewFromEnv()
	if err != nil {
		errLogger.Printf("Failure creating secrets config: %s", err.Error())
		return nil, err
	}

	var accessTokenHandler access_token.AccessTokenHandler
	var secretsHandler secretsHandlers.SecretsHandler

	if storageConfig.StoreType == storageConfigProvider.K8S {
		infoLogger.Printf(fmt.Sprintf("Storage configuration is %s ", storageConfigProvider.K8S))

		accessTokenHandler, err = access_token.NewAccessTokenMemory()
		if err != nil {
			return nil, fmt.Errorf("error creating access token object, reason: %s", err)
		}

		secretsHandler, err = secretsHandlers.NewSecretHandlerK8sUseCase(*secretsConfig, accessTokenHandler)
		if err != nil {
			return nil, fmt.Errorf("error secret handler object, reason: %s", err)
		}
	} else if storageConfig.StoreType == storageConfigProvider.None {
		accessTokenHandler, err = access_token.NewAccessTokenFile(storageConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating access token object, reason: %s", err)
		}

		var secretHandlerNoneUseCase secretsHandlers.SecretHandlerNoneUseCase
		secretsHandler = &secretHandlerNoneUseCase
	} else {
		// although this is checked when creating `storageConfig.StoreType` we check this here for code clarity and future dev guard
		errLogger.Printf("Store type %s is invalid", storageConfig.StoreType)
		return nil, err
	}

	return &Storage_Handler{
		AccessTokenHandler: accessTokenHandler,
		SecretsHandler:     secretsHandler,
	}, nil
}
