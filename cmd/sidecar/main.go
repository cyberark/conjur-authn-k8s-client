package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
	authnConfigProvider "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/storage"
	storageConfigProvider "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
)

// logging
var errorLogger = log.ErrorLogger
var infoLogger = log.InfoLogger

func main() {
	var err error

	// Initialize configurations
	authnConfig, err := authnConfigProvider.NewFromEnv()
	if err != nil {
		printErrorAndExit(log.CAKC045E)
	}

	storageConfig, err := storageConfigProvider.NewFromEnv()
	if err != nil {
		printErrorAndExit(log.CAKC046E)
	}

	if storageConfig.StoreType == storageConfigProvider.K8S && authnConfig.ContainerMode != "init" {
		printErrorAndExit(log.CAKC047E)
	}

	storageHandler, err := storage.NewStorageHandler(*storageConfig)
	if err != nil {
		printErrorAndExit(log.CAKC048E)
	}

	authn, err := authenticator.New(*authnConfig, storageHandler.AccessTokenHandler)
	if err != nil {
		printErrorAndExit(log.CAKC049E)
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
				return log.PrintAndReturnError(log.CAKC050E)
			}

			err = authn.ParseAuthenticationResponse(authnResp)
			if err != nil {
				return log.PrintAndReturnError(log.CAKC051E)
			}

			err = storageHandler.SecretsHandler.HandleSecrets()
			if err != nil {
				return log.PrintAndReturnError(log.CAKC052E)
			}

			err = storageHandler.AccessTokenHandler.Delete()
			if err != nil {
				return log.PrintAndReturnError(log.CAKC065E, err.Error())
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
		// Deleting the retrieved Conjur access token in case we got an error after retrieval.
		// if the access token is already deleted the action should not fail
		err = storageHandler.AccessTokenHandler.Delete()
		if err != nil {
			errorLogger.Printf(log.CAKC054E)
		}

		printErrorAndExit(log.CAKC053E)
	}
}

func printErrorAndExit(errorMessage string) {
	errorLogger.Printf(errorMessage)
	os.Exit(1)
}
