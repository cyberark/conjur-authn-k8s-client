package config

import (
	"fmt"
	"os"
)

type Config struct {
	StoreType     string
	TokenFilePath string
}

const (
	K8S                = "k8s_secrets"
	None               = "none"
	SecretsDestination = "SECRETS_DESTINATION"
)

func NewFromEnv(tokenPath *string) (*Config, error) {
	storeType := None
	tokenFilePath := *tokenPath
	secretsDestinationValue := os.Getenv(SecretsDestination)
	if secretsDestinationValue == K8S {
		storeType = K8S
		tokenFilePath = ""
	} else if secretsDestinationValue == "" || secretsDestinationValue == None {
		storeType = None
		tokenFilePath = *tokenPath
	} else {
		// In case SecretsDestination exits and has configured with incorrect value
		return nil, fmt.Errorf("error incorrect value for environmnet variable %s has provided", SecretsDestination)
	}
	return &Config{
		StoreType:     storeType,
		TokenFilePath: tokenFilePath,
	}, nil
}
