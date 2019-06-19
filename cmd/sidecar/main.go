package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
	authnConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/secrets"
	secretsConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/config"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/sidecar/logging"
)

// logging
var errLogger = log.ErrorLogger
var infoLogger = log.InfoLogger

func main() {
	var err error

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

	// Create new Authenticator
	authn, err := authenticator.New(*configAuthn)
	if err != nil {
		errLogger.Printf(err.Error())
		os.Exit(1)
	}

	configSecrets, err := secretsConfig.NewFromEnv(clientCertPath, tokenFilePath)
	if err != nil {
		errLogger.Printf(err.Error())
		os.Exit(1)
	}

	// Create new Secrets
	secrets, err := secrets.New(*configSecrets)
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

			secretsResp, err := secrets.FetchSecrets()
			if err != nil {
				errLogger.Printf("Failure fetching secrets: %s", err.Error())
				return err
			}

			err = secrets.HandleSecretsResponse(secretsResp)
			if err != nil {
				errLogger.Printf("Failure handling secrets response: %s", err.Error())
				return err
			}

			if authn.Config.ContainerMode == "init" {
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
		os.Exit(1)
	}
}
