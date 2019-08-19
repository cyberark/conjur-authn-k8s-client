package config

import (
	"os"
)

type Config struct {
	StoreType     string
	TokenFilePath string
}

const (
	k8s  = "k8s_secrets"
	none = "none"
)

// should we check here what the ENV says and act accordingly with values?
func NewFromEnv() (*Config, error) {
	storeType := none
	tokenFilePath := "run/conjur/access-token"
	if os.Getenv("SECRETS_DESTINATION") == k8s {
		storeType = k8s
		tokenFilePath = ""
	}
	// else old flow
	return &Config{
		StoreType:     storeType,
		TokenFilePath: tokenFilePath,
	}, nil
}
