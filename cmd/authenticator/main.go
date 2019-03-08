package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
	authnConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
)

// AuthenticateCycleDuration is the default time the system waits to
// reauthenticate on error
const AuthenticateCycleDuration = 6 * time.Minute

// logging
var errLogger = authenticator.ErrorLogger
var infoLogger = authenticator.InfoLogger

func main() {

	var err error

	// Parse any flags for client cert / token paths, and set default values if not passed
	clientCertPath := flag.String("c", "/etc/conjur/ssl/client.pem",
		"Path to client certificate")
	tokenFilePath := flag.String("t", "/run/conjur/access-token",
		"Path to Conjur access token")

	config, err := authnConfig.NewFromEnv(clientCertPath, tokenFilePath)
	handleMainError(err)

	// Create new Authenticator
	authn, err := authenticator.New(*config)
	handleMainError(err)

	// Configure exponential backoff
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 2 * time.Second
	expBackoff.RandomizationFactor = 0.5
	expBackoff.Multiplier = 2
	expBackoff.MaxInterval = 15 * time.Second
	expBackoff.MaxElapsedTime = 2 * time.Minute

	err = backoff.Retry(func() error {
		if err = authn.Login(); err != nil {
			errLogger.Printf("on login: %v", err.Error())
			return err
		}

		for {
			infoLogger.Printf(fmt.Sprintf("authenticating as %s ...", authn.Config.Username))
			resp, err := authn.Authenticate()
			if err != nil {
				errLogger.Printf("on authenticate: %s", err.Error())
				return err
			}

			infoLogger.Printf("valid authentication response")
			err = authn.ParseAuthenticationResponse(resp)
			if err != nil {
				errLogger.Printf("on response parse: %s", err.Error())
				return err
			}

			if authn.Config.ContainerMode == "init" {
				os.Exit(0)
			}

			// Reset exponential backoff
			expBackoff.Reset()

			infoLogger.Printf("waiting for %s to re-authenticate.", AuthenticateCycleDuration)
			fmt.Println()
			time.Sleep(AuthenticateCycleDuration)
		}
	}, expBackoff)
	if err != nil {
		// Handle error.
		errLogger.Printf("backoff exhausted: %s", err.Error())
	}
}

func handleMainError(err error) {
	if err != nil {
		errLogger.Printf(err.Error())
		os.Exit(1)
	}
}
