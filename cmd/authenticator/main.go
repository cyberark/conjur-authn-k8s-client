package main

import (
	"flag"
	"fmt"
	"io/ioutil"
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

	// Load CA cert
	ConjurCACert, err := readSSLCert()
	handleMainError(err)

	config, err := authnConfig.NewFromEnv(ConjurCACert, clientCertPath, tokenFilePath)
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
			if authn.IsCertExpired() {
				infoLogger.Printf("Certificate expired. Re-logging in...")

				if err = authn.Login(); err != nil {
					return err
				}

				infoLogger.Printf("Logged in. Continuing authentication.")
			}

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

func readSSLCert() ([]byte, error) {
	SSLCert := os.Getenv("CONJUR_SSL_CERTIFICATE")
	SSLCertPath := os.Getenv("CONJUR_CERT_FILE")
	if SSLCert == "" && SSLCertPath == "" {
		return nil, fmt.Errorf(
			"at least one of CONJUR_SSL_CERTIFICATE and CONJUR_CERT_FILE must be provided")
	}

	if SSLCert != "" {
		return []byte(SSLCert), nil
	}
	return ioutil.ReadFile(SSLCertPath)
}

func handleMainError(err error) {
	if err != nil {
		errLogger.Printf(err.Error())
		os.Exit(1)
	}
}
