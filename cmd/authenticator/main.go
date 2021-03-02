package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/factory"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

func main() {
	log.Info(log.CAKC048, authenticator.FullVersionName)

	var err error

	authn, errMsg := factory.NewAuthenticatorFromEnv()
	if errMsg != "" {
		printErrorAndExit(errMsg)
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
			err := authn.Authenticate()
			if err != nil {
				return log.RecordedError(log.CAKC016)
			}

			if authn.GlobalConfig().ContainerMode == "init" {
				os.Exit(0)
			}

			log.Info(log.CAKC047, authn.GlobalConfig().TokenRefreshTimeout)

			fmt.Println()
			time.Sleep(authn.GlobalConfig().TokenRefreshTimeout)

			// Reset exponential backoff
			expBackoff.Reset()
		}
	}, expBackoff)

	if err != nil {
		printErrorAndExit(log.CAKC031)
	}
}

func printErrorAndExit(errorMessage string) {
	log.Error(errorMessage)
	os.Exit(1)
}
