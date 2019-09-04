package storage

import (
	"fmt"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
	secretsConfigProvider "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/config"
	secretsHandlers "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/handlers"
	storageConfigProvider "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
)

type StorageHandler struct {
	AccessTokenHandler access_token.AccessTokenHandler
	SecretsHandler     secretsHandlers.SecretsHandler
}

func NewStorageHandler(storageConfig storageConfigProvider.Config) (*StorageHandler, error) {
	var infoLogger = log.InfoLogger

	var accessTokenHandler access_token.AccessTokenHandler
	var secretsHandler secretsHandlers.SecretsHandler
	var err error

	if storageConfig.StoreType == storageConfigProvider.K8S {
		infoLogger.Printf(fmt.Sprintf(log.CAKC001I, storageConfigProvider.K8S))

		secretsConfig, err := secretsConfigProvider.NewFromEnv()
		if err != nil {
			return nil, log.PrintAndReturnError(log.CAKC003E)
		}

		accessTokenHandler, err = access_token.NewAccessTokenMemory()
		if err != nil {
			return nil, log.PrintAndReturnError(log.CAKC004E)
		}

		secretsHandler, err = secretsHandlers.NewSecretHandlerK8sUseCase(*secretsConfig, accessTokenHandler)
		if err != nil {
			return nil, log.PrintAndReturnError(log.CAKC001E)
		}
	} else if storageConfig.StoreType == storageConfigProvider.None {
		accessTokenHandler, err = access_token.NewAccessTokenFile(storageConfig)
		if err != nil {
			return nil, log.PrintAndReturnError(log.CAKC002E)
		}

		var secretHandlerNoneUseCase secretsHandlers.SecretHandlerNoneUseCase
		secretsHandler = &secretHandlerNoneUseCase
	} else {
		// although this is checked when creating `storageConfig.StoreType` we check this here for code clarity and future dev guard
		return nil, log.PrintAndReturnError(log.CAKC005E, storageConfig.StoreType)
	}

	return &StorageHandler{
		AccessTokenHandler: accessTokenHandler,
		SecretsHandler:     secretsHandler,
	}, nil
}
