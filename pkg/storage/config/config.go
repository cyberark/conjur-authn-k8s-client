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

// TODO if SECRET_DESTINATION is k8s then Config struct fields are updated
// Otherwise (if none or non-existent) the returned Config struct will have the access token path to file
// This implementation is not consistent with how we do it in other parts of the code because this implementation
// will be for both `k8s` and `none` flows
func NewFromEnv() (*Config, error) {
	storeType := none
	tokenFilePath := "run/conjur/access-token"
	if os.Getenv("SECRETS_DESTINATION") == k8s {
		storeType = k8s
		tokenFilePath = ""
	}
	return &Config{
		StoreType:     storeType,
		TokenFilePath: tokenFilePath,
	}, nil
}
