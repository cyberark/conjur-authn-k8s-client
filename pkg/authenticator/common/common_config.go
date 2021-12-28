package common

import (
	"fmt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"strconv"
	"time"
)

// Config defines the configuration parameters common for both authentications
type Config struct {
	Account                   string
	ClientCertPath            string
	ClientCertRetryCountLimit int
	ContainerMode             string
	SSLCertificate            []byte
	TokenFilePath             string
	TokenRefreshTimeout       time.Duration
	URL                       string
	Username                  *Username
}

// LoadConfig is a constructor for common Config object
func (config *Config) LoadConfig(settings map[string]string) {
	for key, value := range settings {
		switch key {
		case "CONJUR_ACCOUNT":
			config.Account = value
		case "CONJUR_AUTHN_LOGIN":
			if len(value) == 0 {
				config.Username = nil
			} else {
				username, _ := NewUsername(value)
				config.Username = username
			}
		case "CONJUR_AUTHN_URL":
			config.URL = value
		case "CONJUR_SSL_CERTIFICATE":
			config.SSLCertificate = []byte(value)
		case "CONTAINER_MODE":
			config.ContainerMode = value
		case "CONJUR_AUTHN_TOKEN_FILE":
			config.TokenFilePath = value
		case "CONJUR_CLIENT_CERT_PATH":
			config.ClientCertPath = value
		case "CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT":
			limit, _ := strconv.Atoi(value)
			config.ClientCertRetryCountLimit = limit
		case "CONJUR_TOKEN_TIMEOUT":
			timeout, _ := durationFromString(key, value)
			config.TokenRefreshTimeout = timeout
		}
	}
}

func durationFromString(key, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf(log.CAKC060, key, value)
	}
	return duration, nil
}
