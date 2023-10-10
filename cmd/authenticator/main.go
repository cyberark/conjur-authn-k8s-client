package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"

	"github.com/cenkalti/backoff"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-opentelemetry-tracer/pkg/trace"
)

const (
	tracerName        = "conjur-authn-k8s-client"
	tracerService     = "conjur-authn-k8s-client"
	tracerEnvironment = "production"
	tracerID          = 1
)

func main() {
	// Note: This will log even if the log level is set to "warn" or "error" since that's loaded after this
	log.Info(log.CAKC048, authenticator.FullVersionName)

	var err error

	config, err := config.NewConfigFromEnv()
	if err != nil {
		printErrorAndExit(log.CAKC018)
	}

	// Create a Tracer and parent Span
	tracerType, collectorUrl := trace.TypeFromEnv()
	ctx, tracer, cleanup, err := trace.Create(
		tracerType,
		trace.TracerProviderConfig{
			TracerName:        tracerName,
			TracerService:     tracerService,
			TracerEnvironment: tracerEnvironment,
			TracerID:          tracerID,
			CollectorURL:      collectorUrl,
			ConsoleWriter:     os.Stdout,
		},
	)
	if err != nil {
		printErrorAndExit(err.Error())
	}
	defer cleanup(ctx)

	// Create new Authenticator
	authn, err := authenticator.NewAuthenticator(config)
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
			err := authn.AuthenticateWithContext(ctx, tracer)
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
