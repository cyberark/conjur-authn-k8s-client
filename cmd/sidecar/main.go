package main

import (
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
	authnConfigProvider "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/storage"
	storageConfigProvider "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
	"os"
	"time"
)

// logging
var errorLogger = log.ErrorLogger
var infoLogger = log.InfoLogger

func main() {
	var err error

	// Initialize configurations
	authnConfig, err := authnConfigProvider.NewFromEnv()
	if err != nil {
		errorLogger.Printf(log.CAKC045E)
		os.Exit(1)
	}

	storageConfig, err := storageConfigProvider.NewFromEnv()
	if err != nil {
		errorLogger.Printf(log.CAKC046E)
		os.Exit(1)
	}

	if storageConfig.StoreType == storageConfigProvider.K8S && authnConfig.ContainerMode != "init" {
		errorLogger.Printf(log.CAKC047E)
		os.Exit(1)
	}

	storageHandler, err := storage.NewStorageHandler(*storageConfig)
	if err != nil {
		errorLogger.Printf(log.CAKC048E)
		os.Exit(1)
	}

	authn, err := authenticator.New(*authnConfig, storageHandler.AccessTokenHandler)
	if err != nil {
		errorLogger.Printf(log.CAKC049E)
		os.Exit(1)
	}

	// Configure exponential backoff
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 2 * time.Second
	expBackoff.RandomizationFactor = 0.5
	expBackoff.Multiplier = 2
	expBackoff.MaxInterval = 15 * time.Second
	expBackoff.MaxElapsedTime = 2 * time.Minute

	err = backoff.Retry(func() error {
		for {
			infoLogger.Printf(fmt.Sprintf(log.CAKC019I, authn.Config.Username))
			authnResp, err := authn.Authenticate()
			if err != nil {
				return log.PrintAndReturnError(log.CAKC050E, err, false)
			}

			err = authn.ParseAuthenticationResponse(authnResp)
			if err != nil {
				return log.PrintAndReturnError(log.CAKC051E, err, false)
			}

			err = storageHandler.SecretsHandler.HandleSecrets()
			if err != nil {
				return log.PrintAndReturnError(log.CAKC052E, err, false)
			}

			err = storageHandler.AccessTokenHandler.Delete()
			if err != nil {
				return err
			}

			if authnConfig.ContainerMode == "init" {
				os.Exit(0)
			}

			// Reset exponential backoff
			expBackoff.Reset()

			infoLogger.Printf(log.CAKC013I, authn.Config.TokenRefreshTimeout)

			fmt.Println()
			time.Sleep(authn.Config.TokenRefreshTimeout)
		}
	}, expBackoff)

	if err != nil {
		errorLogger.Printf(log.CAKC053E)
		// Deleting the retrieved Conjur access token in case we got an error after retrieval.
		// if the access token is already deleted the action should not fail
		err = storageHandler.AccessTokenHandler.Delete()
		if err != nil {
			errorLogger.Printf(log.CAKC054E)
		}
		os.Exit(1)
	}
}
