package config

import (
	"os"
)

type Config struct {
	StoreType     string
	TokenFilePath string
}

const (
	K8s  = "k8s_secrets"
	None = "none"
)

// TODO if SECRET_DESTINATION is k8s then Config struct fields are updated
// Otherwise (if none or non-existent) the returned Config struct will have the access token path to file
// This implementation is not consistent with other implementations in the code because here we will take
// both `k8s` and `none` flows into account
func NewFromEnv() (*Config, error) {
	storeType := None
	// TODO: consider moving this configuration to configurable ENV variable
	tokenFilePath := "run/conjur/access-token"
	if os.Getenv("SECRETS_DESTINATION") == K8s {
		storeType = K8s
		tokenFilePath = ""
	}
	return &Config{
		StoreType:     storeType,
		TokenFilePath: tokenFilePath,
	}, nil
}
