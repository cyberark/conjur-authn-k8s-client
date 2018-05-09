package main

import (
	"io/ioutil"
	"os"
	"fmt"
	"time"
	"log"
	"github.com/cenkalti/backoff"
)

const CLIENT_CERT_PATH = "/etc/conjur/ssl/client.pem"
const TOKEN_FILE_PATH = "/run/conjur/access-token"
const AUTHENTICATE_CYCLE_DURATION = 6 * time.Minute

// logging
var errLogger = log.New(os.Stderr, "ERROR: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
var infoLogger = log.New(os.Stdout, "INFO: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)

func main() {
	var err error

	for _, envvar := range([]string{
		"CONJUR_AUTHN_URL",
		"CONJUR_AUTHN_LOGIN",
		"MY_POD_NAMESPACE",
		"MY_POD_NAME",
		}) {
		if os.Getenv(envvar) == "" {
			err = fmt.Errorf(
			"%s must be provided", envvar)
			handleMainError(err)
		}
	}

	authnURL := os.Getenv("CONJUR_AUTHN_URL")
	authnLogin := os.Getenv("CONJUR_AUTHN_LOGIN")
	podNamespace := os.Getenv("MY_POD_NAMESPACE")
	podName := os.Getenv("MY_POD_NAME")

	// Load CA cert
	ConjurCACert, err := ReadSSLCert()
	handleMainError(err)

	auth, err := NewAuthenticator(AuthenticatorConfig{
		authnURL,
		authnLogin,
		podName,
		podNamespace,
		ConjurCACert,
	})
	handleMainError(err)

	// configure exponential backoff
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 2 * time.Second
	expBackoff.RandomizationFactor = 0.5
	expBackoff.Multiplier = 2
	expBackoff.MaxInterval = 15 * time.Second
	expBackoff.MaxElapsedTime = 2 * time.Minute

	err = backoff.Retry(func() error {
		err = Login(auth)
		if err != nil {
			errLogger.Printf("on login: %s", err.Error())
			return err
		}

		for {
			err = Authenticate(auth)
			if err != nil {
				if autherr, ok := err.(*AuthenticatorError); ok {
					if autherr.CertExpired() {
						infoLogger.Printf("certificate expired re-logging in.")

						err = Login(auth)
						if err != nil {
							return err
						}

						// if the cert expired and login worked then con
						continue
					}
				} else {
					errLogger.Printf("on authenticate: %s", err.Error())
					return err
				}
			}

			// Reset exponential backoff
			expBackoff.Reset()

			infoLogger.Printf("waiting for 6 minutes to re-authenticate.")
			time.Sleep(AUTHENTICATE_CYCLE_DURATION)
		}
	}, expBackoff)
	if err != nil {
		// Handle error.
		errLogger.Printf("backoff exhausted: %s", err.Error())
	}
}

func Login(auth *Authenticator)(error) {
	infoLogger.Printf(fmt.Sprintf("logging in as %s.", auth.Username))
	return auth.Login()
}

func Authenticate(auth *Authenticator) (error) {
	infoLogger.Printf(fmt.Sprintf("authenticating as %s ...", auth.Username))
	tokenPemBlock, err := auth.Authenticate()
	if err != nil {
		return err
	}
	infoLogger.Printf("valid authentication response.")

	//debugLogger.Printf("decrypting token ...")
	content, err := decodeFromPEM(tokenPemBlock, auth.publicCert, auth.privateKey)
	if err != nil {
		return err
	}
	//debugLogger.Printf("successfully decrypted token.")

	//debugLogger.Printf("writing token to shared volume ...")
	err = ioutil.WriteFile(TOKEN_FILE_PATH, content, 0644)
	if err != nil {
		return err
	}
	//debugLogger.Printf("token, successfully, written token to shared volume.")

	infoLogger.Printf("successfully authenticated.")
	return nil
}

func ReadSSLCert() ([]byte, error) {
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
