package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
	authnConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// logging
var errLogger = log.ErrorLogger
var infoLogger = log.InfoLogger

func main() {
	printVersion()

	var err error

	config, err := authnConfig.NewFromEnv()
	if err != nil {
		printErrorAndExit(log.CAKC018E)
	}

	// Create new Authenticator
	authn, err := authenticator.New(*config)
	if err != nil {
		printErrorAndExit(log.CAKC019E)
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
			infoLogger.Printf(log.CAKC006I, authn.Config.Username)
			resp, err := authn.Authenticate()
			if err != nil {
				return log.RecordedError(log.CAKC016E)
			}

			err = authn.ParseAuthenticationResponse(resp)
			if err != nil {
				return log.RecordedError(log.CAKC020E)
			}

			if authn.Config.ContainerMode == "init" {
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
		printErrorAndExit(log.CAKC031E)
	}
}

func printErrorAndExit(errorMessage string) {
	errLogger.Printf(errorMessage)
	os.Exit(1)
}

func printVersion(){
	infoLogger.Printf("Kubernetes Authenticator Client v%s starting up...", authenticator.FullVersionName)
}