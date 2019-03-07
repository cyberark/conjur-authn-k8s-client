package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
)

// AuthenticateCycleDuration is the default time the system waits to
// reauthenticate on error
const AuthenticateCycleDuration = 6 * time.Minute

// logging
var errLogger = log.New(os.Stderr, "ERROR: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
var infoLogger = log.New(os.Stdout, "INFO: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)

func main() {

	var err error

	// Parse any flags for client cert / token paths, and set default values if not passed
	clientCertPath := flag.String("c", "/etc/conjur/ssl/client.pem",
		"Path to client certificate")
	tokenFilePath := flag.String("t", "/run/conjur/access-token",
		"Path to Conjur access token")

	// Check that required environment variables are set
	for _, envvar := range []string{
		"CONJUR_AUTHN_URL",
		"CONJUR_ACCOUNT",
		"CONJUR_AUTHN_LOGIN",
		"MY_POD_NAMESPACE",
		"MY_POD_NAME",
	} {
		if os.Getenv(envvar) == "" {
			err = fmt.Errorf(
				"%s must be provided", envvar)
			handleMainError(err)
		}
	}

	// Load configuration from the environment
	authnURL := os.Getenv("CONJUR_AUTHN_URL")
	account := os.Getenv("CONJUR_ACCOUNT")
	authnLogin := os.Getenv("CONJUR_AUTHN_LOGIN")
	podNamespace := os.Getenv("MY_POD_NAMESPACE")
	podName := os.Getenv("MY_POD_NAME")
	containerMode := os.Getenv("CONTAINER_MODE")
	conjurVersion := os.Getenv("CONJUR_VERSION")
	if len(conjurVersion) == 0 {
		conjurVersion = "5"
	}

	// Load CA cert
	ConjurCACert, err := readSSLCert()
	handleMainError(err)

	// Create new Authenticator
	authn, err := authenticator.New(
		authenticator.Config{
			ConjurVersion:  conjurVersion,
			Account:        account,
			URL:            authnURL,
			Username:       authnLogin,
			PodName:        podName,
			PodNamespace:   podNamespace,
			SSLCertificate: ConjurCACert,
			ClientCertPath: *clientCertPath,
			TokenFilePath:  *tokenFilePath,
		})
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

			if containerMode == "init" {
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
