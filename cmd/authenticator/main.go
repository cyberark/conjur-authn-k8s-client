package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/creators"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

func main() {
	log.Info(log.CAKC048, common.FullVersionName)

	var err error

	config, err := creators.NewConfigFromEnv()
	if err != nil {
		printErrorAndExit(log.CAKC018)
	}

	// Create new Authenticator
	authn, err := creators.NewAuthenticator(config)
	if err != nil {
		printErrorAndExit(log.CAKC019)
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

			if config.GetContainerMode() == "init" {
				os.Exit(0)
			}

			log.Info(log.CAKC047, config.GetTokenTimeout())

			fmt.Println()
			time.Sleep(config.GetTokenTimeout())

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
