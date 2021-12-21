package common

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// ReadFileFunc defines the interface for reading an SSL Certificate from the env
type ReadFileFunc func(filename string) ([]byte, error)

func validTimeout(key, timeoutStr string) error {
	_, err := durationFromString(key, timeoutStr)
	return err
}

func validInt(key, value string) error {
	_, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf(log.CAKC060, key, value)
	}
	return nil
}

func validUsername(key, value string) error {
	_, err := NewUsername(value)
	return err
}

func validConjurVersion(key, version string) error {
	// Only versions '4' & '5' are allowed, with '5' being used as the default
	switch version {
	case "4":
		break
	case "5":
		break
	default:
		return fmt.Errorf(log.CAKC060, key, version)
	}

	return nil
}

func ValidateSetting(key string, value string) error {
	switch key {
	case "CONJUR_AUTHN_LOGIN":
		return validUsername(key, value)
	case "CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT":
		return validInt(key, value)
	case "CONJUR_TOKEN_TIMEOUT":
		return validTimeout(key, value)
	case "CONJUR_VERSION":
		return validConjurVersion(key, value)
	case "JWT_TOKEN_PATH":
		return validatePath(value)
	default:
		return nil
	}
}

func ReadSSLCert(settings map[string]string, readFile ReadFileFunc) ([]byte, error) {
	SSLCert := settings["CONJUR_SSL_CERTIFICATE"]
	SSLCertPath := settings["CONJUR_CERT_FILE"]
	if SSLCert == "" && SSLCertPath == "" {
		return nil, errors.New(log.CAKC007)
	}

	if SSLCert != "" {
		return []byte(SSLCert), nil
	}
	return readFile(SSLCertPath)
}

func validatePath(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf(log.CAKC065, path)
	}
	return nil
}
