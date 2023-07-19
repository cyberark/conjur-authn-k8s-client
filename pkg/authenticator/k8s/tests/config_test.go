package tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/k8s"
)

var environmentValues = map[string]string{
	"CONJUR_AUTHN_URL":       "filepath",
	"CONJUR_ACCOUNT":         "testAccount",
	"CONJUR_AUTHN_LOGIN":     "host",
	"CONJUR_CERT_FILE":       "testSSLCertFile.txt",
	"CONJUR_SSL_CERTIFICATE": "testSSLCert",
	"MY_POD_NAMESPACE":       "testNameSpace",
	"MY_POD_NAME":            "testPodName",
}

var annotationValues = map[string]string{
	"conjur.org/authn-identity": "host/anotherHost",
	"conjur.org/log-level":      "debug",
	"conjur.org/container-mode": "init",
}

var envToAnnot = map[string]string{
	"CONJUR_AUTHN_LOGIN": "conjur.org/authn-identity",
	"LOG_LEVEL":          "conjur.org/log-level",
	"CONTAINER_MODE":     "conjur.org/container-mode",
}

func setEnv(env map[string]string) {
	for key, value := range env {
		os.Setenv(key, value)
	}
}

func unsetEnv(env map[string]string) {
	for key := range env {
		os.Setenv(key, "")
	}
}

func TestGatherSettings(t *testing.T) {
	TestCases := []struct {
		description string
		annotFunc   func(string) string
		expected    config.AuthnSettings
	}{
		{
			description: "functions are ordered by priority: first function overrides second, which overrides third",
			annotFunc:   fromAnnotations,
			expected: config.AuthnSettings{
				"CONJUR_ACCOUNT":                       "testAccount",
				"CONJUR_AUTHN_LOGIN":                   "host/anotherHost", // provided by annotation
				"CONJUR_AUTHN_URL":                     "filepath",
				"CONJUR_CERT_FILE":                     "testSSLCertFile.txt",
				"CONJUR_SSL_CERTIFICATE":               "testSSLCert",
				"CONTAINER_MODE":                       "init",  // provided by annotation
				"LOG_LEVEL":                            "debug", // provided by annotation
				"DEBUG":                                "",
				"MY_POD_NAME":                          "testPodName",
				"MY_POD_NAMESPACE":                     "testNameSpace",
				"CONJUR_AUTHN_TOKEN_FILE":              k8s.DefaultTokenFilePath,
				"CONJUR_CLIENT_CERT_PATH":              k8s.DefaultClientCertPath,
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": k8s.DefaultClientCertRetryCountLimit,
				"CONJUR_TOKEN_TIMEOUT":                 k8s.DefaultTokenRefreshTimeout,
			},
		},
		{
			description: "if the first getter function returns empty strings, fallback to the next functions, and eventually an empty string",
			annotFunc:   emptyAnnotations,
			expected: config.AuthnSettings{
				"CONJUR_AUTHN_URL":                     "filepath",
				"CONJUR_ACCOUNT":                       "testAccount",
				"CONJUR_AUTHN_LOGIN":                   "host",
				"CONJUR_CERT_FILE":                     "testSSLCertFile.txt",
				"CONJUR_SSL_CERTIFICATE":               "testSSLCert",
				"MY_POD_NAMESPACE":                     "testNameSpace",
				"MY_POD_NAME":                          "testPodName",
				"LOG_LEVEL":                            "",
				"DEBUG":                                "",
				"CONTAINER_MODE":                       "",
				"CONJUR_CLIENT_CERT_PATH":              k8s.DefaultClientCertPath,
				"CONJUR_AUTHN_TOKEN_FILE":              k8s.DefaultTokenFilePath,
				"CONJUR_TOKEN_TIMEOUT":                 k8s.DefaultTokenRefreshTimeout,
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": k8s.DefaultClientCertRetryCountLimit,
			},
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			resultMap := config.GatherSettings(&k8s.Config{}, tc.annotFunc, fromEnv)
			assert.Equal(t, tc.expected, resultMap)
		})
	}
}

func TestFromEnv(t *testing.T) {
	TestCases := []struct {
		description string
		env         map[string]string
		expectErr   bool
	}{
		{
			description: "happy path",
			env: map[string]string{
				"CONJUR_AUTHN_URL":       "authn-k8s",
				"CONJUR_ACCOUNT":         "testAccount",
				"CONJUR_AUTHN_LOGIN":     "host/test-user",
				"MY_POD_NAME":            "testPodName",
				"MY_POD_NAMESPACE":       "testNameSpace",
				"CONJUR_SSL_CERTIFICATE": "samplecert",
			},
			expectErr: false,
		},
		{
			description: "bad settings return nil Config and error message",
			env:         map[string]string{},
			expectErr:   true,
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			// SETUP & EXERCISE
			setEnv(tc.env)
			config, err := config.ConfigFromEnv(successfulMockReadFile)

			// ASSERT
			if tc.expectErr {
				assert.Nil(t, config)
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, config)
				assert.Nil(t, err)
			}
			unsetEnv(tc.env)
		})
	}
}

func successfulMockReadFile(filename string) ([]byte, error) {
	return []byte{}, nil
}

func fromEnv(key string) string {
	return environmentValues[key]
}

func fromAnnotations(key string) string {
	annot := envToAnnot[key]
	return annotationValues[annot]
}

func emptyAnnotations(key string) string {
	return ""
}
