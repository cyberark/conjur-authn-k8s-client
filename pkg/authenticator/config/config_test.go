package config

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	logger "github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/stretchr/testify/assert"
)

var environmentValues = map[string]string{
	"CONJUR_AUTHN_URL":       "filepath",
	"CONJUR_ACCOUNT":         "testAccount",
	"CONJUR_AUTHN_LOGIN":     "host",
	"CONJUR_CERT_FILE":       "testSSLCertFile.txt",
	"CONJUR_SSL_CERTIFICATE": "testSSLCert",
	"CONJUR_VERSION":         "",
	"MY_POD_NAMESPACE":       "testNameSpace",
	"MY_POD_NAME":            "testPodName",
}

var annotationValues = map[string]string{
	"conjur.org/authn-identity": "host/anotherHost",
	"conjur.org/debug-logging":  "true",
	"conjur.org/container-mode": "init",
}

var envToAnnot = map[string]string{
	"CONJUR_AUTHN_LOGIN": "conjur.org/authn-identity",
	"DEBUG":              "conjur.org/debug-logging",
	"CONTAINER_MODE":     "conjur.org/container-mode",
}

func assertGoodConfig(expected *Config) func(*testing.T, *Config) {
	return func(t *testing.T, result *Config) {
		assert.Equal(t, expected, result)
	}
}

type errorAssertFunc func(*testing.T, []error)

func assertEmptyErrorList() errorAssertFunc {
	return func(t *testing.T, errorList []error) {
		assert.Empty(t, errorList)
	}
}

func assertErrorInList(err error) errorAssertFunc {
	return func(t *testing.T, errorList []error) {
		assert.Contains(t, errorList, err)
	}
}

func assertErrorNotInList(err error) errorAssertFunc {
	return func(t *testing.T, errorList []error) {
		assert.NotContains(t, errorList, err)
	}
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
		expected    AuthnSettings
	}{
		{
			description: "functions are ordered by priority: first function overrides second, which overrides third",
			annotFunc:   fromAnnotations,
			expected: AuthnSettings{
				"CONJUR_ACCOUNT":                       "testAccount",
				"CONJUR_AUTHN_LOGIN":                   "host/anotherHost", // provided by annotation
				"CONJUR_AUTHN_URL":                     "filepath",
				"CONJUR_CERT_FILE":                     "testSSLCertFile.txt",
				"CONJUR_SSL_CERTIFICATE":               "testSSLCert",
				"CONTAINER_MODE":                       "init", // provided by annotation
				"DEBUG":                                "true", // provided by annotation
				"MY_POD_NAME":                          "testPodName",
				"MY_POD_NAMESPACE":                     "testNameSpace",
				"CONJUR_AUTHN_TOKEN_FILE":              DefaultTokenFilePath,
				"CONJUR_CLIENT_CERT_PATH":              DefaultClientCertPath,
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": DefaultClientCertRetryCountLimit,
				"CONJUR_TOKEN_TIMEOUT":                 DefaultTokenRefreshTimeout,
				"CONJUR_VERSION":                       DefaultConjurVersion,
			},
		},
		{
			description: "if the first getter function returns empty strings, fallback to the next functions, and eventually an empty string",
			annotFunc:   emptyAnnotations,
			expected: AuthnSettings{
				"CONJUR_AUTHN_URL":                     "filepath",
				"CONJUR_ACCOUNT":                       "testAccount",
				"CONJUR_AUTHN_LOGIN":                   "host",
				"CONJUR_CERT_FILE":                     "testSSLCertFile.txt",
				"CONJUR_SSL_CERTIFICATE":               "testSSLCert",
				"MY_POD_NAMESPACE":                     "testNameSpace",
				"MY_POD_NAME":                          "testPodName",
				"DEBUG":                                "",
				"CONTAINER_MODE":                       "",
				"CONJUR_CLIENT_CERT_PATH":              DefaultClientCertPath,
				"CONJUR_AUTHN_TOKEN_FILE":              DefaultTokenFilePath,
				"CONJUR_VERSION":                       DefaultConjurVersion,
				"CONJUR_TOKEN_TIMEOUT":                 DefaultTokenRefreshTimeout,
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": DefaultClientCertRetryCountLimit,
			},
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			resultMap := GatherSettings(tc.annotFunc, fromEnv)
			assert.Equal(t, tc.expected, resultMap)
		})
	}
}

func TestValidate(t *testing.T) {
	TestCases := []struct {
		description string
		settings    AuthnSettings
		assert      errorAssertFunc
	}{
		{
			description: "happy path",
			settings: AuthnSettings{
				// required variables
				"CONJUR_AUTHN_URL":   "filepath",
				"CONJUR_ACCOUNT":     "testAccount",
				"CONJUR_AUTHN_LOGIN": "host",
				"MY_POD_NAME":        "testPodName",
				"MY_POD_NAMESPACE":   "testNameSpace",
				// correct value types
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": "7",
				"CONJUR_TOKEN_TIMEOUT":                 "6m0s",
				"CONTAINER_MODE":                       "init",
				// certificate provided
				"CONJUR_SSL_CERTIFICATE": "samplecertificate",
				// valid version
				"CONJUR_VERSION": "5",
			},
			assert: assertEmptyErrorList(),
		},
		{
			description: "error raised for missing required setting",
			settings:    AuthnSettings{},
			assert:      assertErrorInList(fmt.Errorf(logger.CAKC062, "CONJUR_AUTHN_URL")),
		},
		{
			description: "error raised for invalid username",
			settings: AuthnSettings{
				"CONJUR_AUTHN_URL":   "filepath",
				"CONJUR_ACCOUNT":     "testAccount",
				"CONJUR_AUTHN_LOGIN": "bad-username",
				"MY_POD_NAME":        "testPodName",
				"MY_POD_NAMESPACE":   "testNameSpace",
			},
			assert: assertErrorInList(fmt.Errorf(logger.CAKC032, "bad-username")),
		},
		{
			description: "error raised for invalid retry count limit",
			settings: AuthnSettings{
				"CONJUR_AUTHN_URL":                     "filepath",
				"CONJUR_ACCOUNT":                       "testAccount",
				"CONJUR_AUTHN_LOGIN":                   "host",
				"MY_POD_NAME":                          "testPodName",
				"MY_POD_NAMESPACE":                     "testNameSpace",
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": "seven",
			},
			assert: assertErrorInList(fmt.Errorf(logger.CAKC060, "CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT", "seven")),
		},
		{
			description: "error raised for invalid timeout",
			settings: AuthnSettings{
				"CONJUR_AUTHN_URL":                     "filepath",
				"CONJUR_ACCOUNT":                       "testAccount",
				"CONJUR_AUTHN_LOGIN":                   "host",
				"MY_POD_NAME":                          "testPodName",
				"MY_POD_NAMESPACE":                     "testNameSpace",
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": "7",
				"CONJUR_TOKEN_TIMEOUT":                 "seventeen",
			},
			assert: assertErrorInList(fmt.Errorf(logger.CAKC060, "CONJUR_TOKEN_TIMEOUT", "seventeen")),
		},
		{
			description: "error raised for invalid certificate",
			settings: AuthnSettings{
				"CONJUR_AUTHN_URL":                     "filepath",
				"CONJUR_ACCOUNT":                       "testAccount",
				"CONJUR_AUTHN_LOGIN":                   "host",
				"MY_POD_NAME":                          "testPodName",
				"MY_POD_NAMESPACE":                     "testNameSpace",
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": "7",
				"CONJUR_TOKEN_TIMEOUT":                 "6m0s",
				"CONTAINER_MODE":                       "init",
				"CONJUR_SSL_CERTIFICATE":               "",
				"CONJUR_CERT_FILE":                     "",
			},
			assert: assertErrorInList(errors.New(logger.CAKC007)),
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			errLogs := tc.settings.Validate(successfulMockReadFile)
			tc.assert(t, errLogs)
		})
	}
}

func TestNewConfig(t *testing.T) {
	TestCases := []struct {
		description string
		settings    AuthnSettings
		expected    *Config
		assert      func(*testing.T, *Config)
	}{
		{
			description: "happy path",
			settings: AuthnSettings{
				// required variables
				"CONJUR_AUTHN_URL":   "filepath",
				"CONJUR_ACCOUNT":     "testAccount",
				"CONJUR_AUTHN_LOGIN": "host/test-user",
				"MY_POD_NAME":        "testPodName",
				"MY_POD_NAMESPACE":   "testNameSpace",
				// correct value types
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": "7",
				"CONJUR_TOKEN_TIMEOUT":                 "6m0s",
				"CONTAINER_MODE":                       "init",
				// certificate provided
				"CONJUR_SSL_CERTIFICATE": "samplecertificate",
				// defaults provided
				"CONJUR_AUTHN_TOKEN_FILE": DefaultTokenFilePath,
				"CONJUR_CLIENT_CERT_PATH": DefaultClientCertPath,
				"CONJUR_VERSION":          DefaultConjurVersion,
			},
			assert: assertGoodConfig(&Config{
				Account:                   "testAccount",
				ClientCertPath:            DefaultClientCertPath,
				ClientCertRetryCountLimit: 7,
				ContainerMode:             "init",
				ConjurVersion:             "5",
				InjectCertLogPath:         DefaultInjectCertLogPath,
				PodName:                   "testPodName",
				PodNamespace:              "testNameSpace",
				SSLCertificate:            []byte("samplecertificate"),
				TokenFilePath:             DefaultTokenFilePath,
				TokenRefreshTimeout:       360000000000,
				URL:                       "filepath",
				Username: &Username{
					FullUsername: "host/test-user",
					Prefix:       "host",
					Suffix:       "test-user",
				},
			}),
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			config := tc.settings.NewConfig()
			tc.assert(t, config)
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
				"CONJUR_AUTHN_URL":       "filepath",
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
			setEnv(tc.env)
			config, err := FromEnv(successfulMockReadFile)

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

func TestConjurVersion(t *testing.T) {
	TestCases := []struct {
		description string
		version     string
		expVersion  string
		assert      errorAssertFunc
	}{
		{
			description: "Succeeds if version is 4",
			version:     "4",
			expVersion:  "4",
			assert:      assertErrorNotInList(fmt.Errorf(logger.CAKC060, "CONJUR_VERSION", "4")),
		},
		{
			description: "Succeeds if version is 5",
			version:     "5",
			expVersion:  "5",
			assert:      assertErrorNotInList(fmt.Errorf(logger.CAKC060, "CONJUR_VERSION", "5")),
		},
		{
			description: "Sets the default version for an empty value",
			version:     "",
			expVersion:  DefaultConjurVersion,
			assert:      assertErrorNotInList(fmt.Errorf(logger.CAKC060, "CONJUR_VERSION", DefaultConjurVersion)),
		},
		{
			description: "Returns error if version is invalid",
			version:     "3",
			expVersion:  "",
			assert:      assertErrorInList(fmt.Errorf(logger.CAKC060, "CONJUR_VERSION", "3")),
		},
	}

	for _, tc := range TestCases {
		provideVersion := func(key string) string {
			if key == "CONJUR_VERSION" {
				return tc.version
			}
			return ""
		}

		t.Run(tc.description, func(t *testing.T) {
			settings := GatherSettings(provideVersion)
			errLogs := settings.Validate(successfulMockReadFile)
			tc.assert(t, errLogs)
			if tc.expVersion != "" {
				assert.Equal(t, tc.expVersion, settings["CONJUR_VERSION"])
			}
		})
	}
}

func TestDebugLogging(t *testing.T) {
	TestCases := []struct {
		description string
		debugValue  string
		settings    AuthnSettings
	}{
		{
			description: "debug logs are enabled",
			debugValue:  "true",
			settings: AuthnSettings{
				// required variables
				"CONJUR_AUTHN_URL":   "filepath",
				"CONJUR_ACCOUNT":     "testAccount",
				"CONJUR_AUTHN_LOGIN": "host/test-user",
				"MY_POD_NAME":        "testPodName",
				"MY_POD_NAMESPACE":   "testNameSpace",
				// correct value types
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": "7",
				"CONJUR_TOKEN_TIMEOUT":                 "6m0s",
				"CONTAINER_MODE":                       "init",
				// certificate provided
				"CONJUR_SSL_CERTIFICATE": "samplecertificate",
				// defaults provided
				"CONJUR_AUTHN_TOKEN_FILE": DefaultTokenFilePath,
				"CONJUR_CLIENT_CERT_PATH": DefaultClientCertPath,
				"CONJUR_VERSION":          DefaultConjurVersion,
				// debug setting
				"DEBUG": "true",
			},
		},
		{
			description: "debug logs are disabled",
			debugValue:  "",
			settings: AuthnSettings{
				// required variables
				"CONJUR_AUTHN_URL":   "filepath",
				"CONJUR_ACCOUNT":     "testAccount",
				"CONJUR_AUTHN_LOGIN": "host/test-user",
				"MY_POD_NAME":        "testPodName",
				"MY_POD_NAMESPACE":   "testNameSpace",
				// correct value types
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": "7",
				"CONJUR_TOKEN_TIMEOUT":                 "6m0s",
				"CONTAINER_MODE":                       "init",
				// certificate provided
				"CONJUR_SSL_CERTIFICATE": "samplecertificate",
				// defaults provided
				"CONJUR_AUTHN_TOKEN_FILE": DefaultTokenFilePath,
				"CONJUR_CLIENT_CERT_PATH": DefaultClientCertPath,
				"CONJUR_VERSION":          DefaultConjurVersion,
			},
		},
		{
			description: "debug logs are given an incorrect value",
			debugValue:  "garbage",
			settings: AuthnSettings{
				// required variables
				"CONJUR_AUTHN_URL":   "filepath",
				"CONJUR_ACCOUNT":     "testAccount",
				"CONJUR_AUTHN_LOGIN": "host/test-user",
				"MY_POD_NAME":        "testPodName",
				"MY_POD_NAMESPACE":   "testNameSpace",
				// correct value types
				"CONJUR_CLIENT_CERT_RETRY_COUNT_LIMIT": "7",
				"CONJUR_TOKEN_TIMEOUT":                 "6m0s",
				"CONTAINER_MODE":                       "init",
				// certificate provided
				"CONJUR_SSL_CERTIFICATE": "samplecertificate",
				// defaults provided
				"CONJUR_AUTHN_TOKEN_FILE": DefaultTokenFilePath,
				"CONJUR_CLIENT_CERT_PATH": DefaultClientCertPath,
				"CONJUR_VERSION":          DefaultConjurVersion,
				// debug setting
				"DEBUG": "garbage",
			},
		},
	}

	for _, tc := range TestCases {
		var logBuffer bytes.Buffer
		logger.InfoLogger = log.New(&logBuffer, "", 0)

		config := tc.settings.NewConfig()
		assert.NotNil(t, config)

		logMessages := logBuffer.String()
		if tc.debugValue == "true" {
			assert.Contains(t, logMessages, "CAKC052")
			assert.NotContains(t, logMessages, "CAKC034")
		} else if tc.debugValue == "" {
			assert.NotContains(t, logMessages, "CAKC052")
			assert.NotContains(t, logMessages, "CAKC034")
		} else {
			assert.NotContains(t, logMessages, "CAKC052")
			assert.Contains(t, logMessages, "CAKC034")
		}
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
