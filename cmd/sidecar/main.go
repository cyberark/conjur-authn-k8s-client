package main

import (
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
	authnConfigProvider "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	secretsConfigProvider "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/config"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/sidecar/logging"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/storage"
	storageConfigProvider "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
	"os"
	"time"
)

// logging
var errLogger = log.ErrorLogger
var infoLogger = log.InfoLogger

func main() {
	var err error

	// Initialize configurations
	authnConfig, err := authnConfigProvider.NewFromEnv()
	if err != nil {
		errLogger.Printf(err.Error())
		os.Exit(1)
	}

	secretsConfig, err := secretsConfigProvider.NewFromEnv()
	if err != nil {
		errLogger.Printf("Failure creating secrets config: %s", err.Error())
		os.Exit(1)
	}

	storageConfig, err := storageConfigProvider.NewFromEnv()
	if err != nil {
		errLogger.Printf(err.Error())
		os.Exit(1)
	}

	if storageConfig.StoreType == storageConfigProvider.K8S && authnConfig.ContainerMode != "init" {
		infoLogger.Printf("Store type 'K8S' must run as an init container")
		os.Exit(1)
	}

	storageHandler, err := storage.NewStorageHandler(*storageConfig, *secretsConfig)
	if err != nil {
		errLogger.Printf(err.Error())
		os.Exit(1)
	}

	authn, err := authenticator.New(*authnConfig, storageHandler.AccessTokenHandler)
	if err != nil {
		errLogger.Printf(err.Error())
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
			infoLogger.Printf(fmt.Sprintf("Authenticating as %s ...", authn.Config.Username))
			authnResp, err := authn.Authenticate()
			if err != nil {
				errLogger.Printf("Failure authenticating: %s", err.Error())
				return err
			}

			err = authn.ParseAuthenticationResponse(authnResp)
			if err != nil {
				errLogger.Printf("Failure parsing authentication response: %s", err.Error())
				return err
			}

			err = storageHandler.SecretsHandler.HandleSecrets()
			if err != nil {
				errLogger.Printf("Failed to handle secrets: %s", err.Error())
				return err
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

			infoLogger.Printf("Waiting for %s to re-authenticate and fetch secrets.",
				authn.Config.TokenRefreshTimeout)

			fmt.Println()
			time.Sleep(authn.Config.TokenRefreshTimeout)
		}
	}, expBackoff)

	if err != nil {
		errLogger.Printf("Backoff exhausted: %s", err.Error())
		// Deleting the retrieved Conjur access token in case we got an error after retrieval.
		// if the access token is already deleted the action should not fail
		err = storageHandler.AccessTokenHandler.Delete()
		if err != nil {
			errLogger.Printf("failed to delete access token: %s", err.Error())
		}
		os.Exit(1)
	}
}
