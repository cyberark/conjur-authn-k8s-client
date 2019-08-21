package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
	authnConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	storageHandler "github.com/cyberark/conjur-authn-k8s-client/pkg/storage"
	storageConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/secrets"
	secretsConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/config"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/sidecar/logging"
)

// logging
var errLogger = log.ErrorLogger
var infoLogger = log.InfoLogger

func main() {
	var err error
	var secretsHandler *secrets.Secrets

	const(
		none = "none"
		k8s = "k8s_secrets"
	)

	// Parse any flags for client cert / token paths, and set default values if not passed
	clientCertPath := flag.String("c", "/etc/conjur/ssl/client.pem",
		"Path to client certificate")
	tokenFilePath := flag.String("t", "/run/conjur/access-token",
		"Path to Conjur access token")

	configAuthn, err := authnConfig.NewFromEnv(clientCertPath, tokenFilePath)
	if err != nil {
		errLogger.Printf(err.Error())
		os.Exit(1)
	}

	if configAuthn.ContainerMode == "init" {
		os.Exit(0)
	}

	// Create new Authenticator
	authn, err := authenticator.New(*configAuthn)
	if err != nil {
		errLogger.Printf(err.Error())
		os.Exit(1)
	}

	// Create new Storage
	configStorage, err := storageConfig.NewFromEnv()
	if err != nil {
		errLogger.Printf(err.Error())
		os.Exit(1)
	}
	stor, err := storageHandler.NewStorageHandler(*configStorage)
	if err != nil {
		errLogger.Printf(err.Error())
		os.Exit(1)
	}

	// TODO: account for if SECRET_DESTINATION is set to k8s but not an init container. Must check this case early.
	conjurSecretsDest := none
	/* if conjurSecretsDest == 1  &&  authn.Config.ContainerMode != "init" {
	     // notify on invalid configuration and exit
	     infoLogger.Printf("Appropriate message for not supporting sidecar with this workflow ....")
	     os.Exit(1)
	} */

	// TODO: move these code segments to their respective files
	//  Code that will handle the placement of access token
	if conjurSecretsDest == none {
		configSecrets, err := secretsConfig.NewFromEnv(tokenFilePath)
		if err != nil {
			errLogger.Printf("Failure creating secrets config: %s", err.Error())
			os.Exit(1)
		}

		// Create new Secrets
		secretsHandler, err = secrets.New(*configSecrets)
		if err != nil {
			errLogger.Printf("Failure creating secrets: %s", err.Error())
			os.Exit(1)
		}
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

			// TODO: move these code segments to their respective files
			//  Code that will handle the retrieval and updating of Conjur to K8s secrets
			// ------------ START LOGIC ------------
			k8sSecretsMap, err := secretsHandler.RetrieveK8sSecrets()
			if err != nil {
				errLogger.Printf("Failure retrieving k8s secrets: %s", err.Error())
				return err
			}

			k8sSecretsMap, err = secretsHandler.UpdateK8sSecretsMapWithConjurSecrets(k8sSecretsMap)
			if err != nil {
				errLogger.Printf("Failure updating K8s Secrets map: %s", err.Error())
				return err
			}

			err = secretsHandler.PatchK8sSecrets(k8sSecretsMap)
			if err != nil {
				errLogger.Printf("Failure patching K8s Secrets: %s", err.Error())
				return err
			}
			// ------------ END LOGIC ------------

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
		os.Exit(1)
	}
}