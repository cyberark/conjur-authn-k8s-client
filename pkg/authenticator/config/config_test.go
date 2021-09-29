package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	logger "github.com/cyberark/conjur-authn-k8s-client/pkg/log"
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

func setEnvVars(env map[string]string) {
	for key, value := range env {
		err := os.Setenv(key, value)
		if err != nil {
			panic(err)
		}
	}
}

func unsetEnvVars(env map[string]string) {
	for key := range env {
		err := os.Setenv(key, "")
		if err != nil {
			return
		}
	}
}

type errorAssertFunc func(*testing.T, []error, []error)

func assertEmptyErrorList() errorAssertFunc {
	return func(t *testing.T, errorList []error, infoList []error) {
		assert.Empty(t, errorList)
	}
}

func assertInfoInList(err error) errorAssertFunc {
	return func(t *testing.T, errorList []error, infoList []error) {
		assert.Contains(t, infoList, err)
	}
}

func assertErrorInList(err error) errorAssertFunc {
	return func(t *testing.T, errorList []error, infoList []error) {
		assert.Contains(t, errorList, err)
	}
}

func assertGoodConfig(expected *Config, result *Config) func(*testing.T) {
	return func(t *testing.T) {
		assert.Equal(t, expected, result)
	}
}

func TestGatherSettings(t *testing.T) {
	setEnvVars(environmentValues)
	defer unsetEnvVars(environmentValues)

	TestCases := []struct {
		description string
		annotations map[string]string
		expectedMap map[string]string
	}{
		{
			description: "returned map contains envvar and annotation settings",
			annotations: map[string]string{
				"conjur.org/authn-identity": "host/anotherHost",
				"conjur.org/debug-logging":  "true",
				"conjur.org/container-mode": "init",
				"conjur.org/unused":         "value",
			},
			expectedMap: map[string]string{
				"conjur.org/authn-identity": "host/anotherHost",
				"conjur.org/debug-logging":  "true",
				"conjur.org/container-mode": "init",
				"CONJUR_AUTHN_URL":          "filepath",
				"CONJUR_ACCOUNT":            "testAccount",
				"CONJUR_AUTHN_LOGIN":        "host",
				"CONJUR_CERT_FILE":          "testSSLCertFile.txt",
				"CONJUR_SSL_CERTIFICATE":    "testSSLCert",
				"MY_POD_NAMESPACE":          "testNameSpace",
				"MY_POD_NAME":               "testPodName",
			},
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			settings := GatherSettings(tc.annotations)
			assert.Equal(t, tc.expectedMap, settings)
		})
	}
}

func TestValidateSettings(t *testing.T) {
	TestCases := []struct {
		description string
		env         map[string]string
		annotations map[string]string
		assert      errorAssertFunc
	}{
		{
			description: "returned map contains envvar and annotation settings",
			env:         environmentValues,
			annotations: map[string]string{
				"conjur.org/authn-identity": "host/anotherHost",
				"conjur.org/debug-logging":  "true",
				"conjur.org/container-mode": "init",
			},
			assert: assertEmptyErrorList(),
		},
		{
			description: "a required envvar is not included",
			env: map[string]string{
				// "CONJUR_AUTHN_URL":       "filepath",
				"CONJUR_ACCOUNT":         "testAccount",
				"CONJUR_AUTHN_LOGIN":     "host",
				"MY_POD_NAMESPACE":       "testNameSpace",
				"MY_POD_NAME":            "testPodName",
				"CONJUR_SSL_CERTIFICATE": "testSSLCert",
			},
			annotations: map[string]string{},
			assert:      assertErrorInList(fmt.Errorf(logger.CAKC009, "CONJUR_AUTHN_URL")),
		},
		{
			description: "configuration Username is not set by envvar or annotation",
			env: map[string]string{
				"CONJUR_AUTHN_URL": "filepath",
				// "CONJUR_AUTHN_LOGIN": "host",
				"CONJUR_ACCOUNT":         "testAccount",
				"MY_POD_NAMESPACE":       "testNameSpace",
				"MY_POD_NAME":            "testPodName",
				"CONJUR_SSL_CERTIFICATE": "testSSLCert",
			},
			annotations: map[string]string{},
			assert:      assertErrorInList(fmt.Errorf(logger.CAKC061, "Username", "CONJUR_AUTHN_LOGIN", "conjur.org/authn-identity")),
		},
		{
			description: "setting with multiple sources provided by both",
			env: map[string]string{
				"CONJUR_AUTHN_URL":       "filepath",
				"CONJUR_AUTHN_LOGIN":     "host",
				"CONJUR_ACCOUNT":         "testAccount",
				"MY_POD_NAMESPACE":       "testNameSpace",
				"MY_POD_NAME":            "testPodName",
				"CONJUR_SSL_CERTIFICATE": "testSSLCert",
			},
			annotations: map[string]string{
				"conjur.org/authn-identity": "host/anotherHost",
			},
			assert: assertInfoInList(fmt.Errorf(logger.CAKC060, "Username", "CONJUR_AUTHN_LOGIN", "conjur.org/authn-identity")),
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			setEnvVars(tc.env)
			defer unsetEnvVars(tc.env)

			settings := GatherSettings(tc.annotations)
			errLogs, infoLogs := ValidateSettings(settings, successfulMockReadFile)

			tc.assert(t, errLogs, infoLogs)
		})
	}
}

func TestNewConfig(t *testing.T) {
	TestCases := []struct {
		description string
		settings    map[string]string
	}{
		{
			description: "annotations overwrite envvars",
			settings: map[string]string{
				"CONJUR_AUTHN_URL":          "filepath",
				"CONJUR_ACCOUNT":            "testAccount",
				"MY_POD_NAMESPACE":          "testNameSpace",
				"MY_POD_NAME":               "testPodName",
				"CONJUR_SSL_CERTIFICATE":    "testSSLCert",
				"CONTAINER_MODE":            "application",
				"conjur.org/container-mode": "init",
				"CONJUR_AUTHN_LOGIN":        "host",
				"conjur.org/authn-identity": "host/anotherHost",
			},
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			config := NewConfig(tc.settings)
			expectedUsername, _ := NewUsername("host/anotherHost")
			defaultTimeout, _ := time.ParseDuration(DefaultTokenRefreshTimeout)
			defaultCountLimit, _ := strconv.Atoi(DefaultClientCertRetryCountLimit)

			assertGoodConfig(&Config{
				Account:                   "testAccount",
				ContainerMode:             "init",
				ConjurVersion:             "5",
				PodName:                   "testPodName",
				PodNamespace:              "testNameSpace",
				Username:                  expectedUsername,
				URL:                       "filepath",
				SSLCertificate:            []byte("testSSLCert"),
				ClientCertPath:            "",
				TokenRefreshTimeout:       defaultTimeout,
				ClientCertRetryCountLimit: defaultCountLimit,
				InjectCertLogPath:         DefaultInjectCertLogPath,
				TokenFilePath:             DefaultTokenFilePath,
			}, config)
		})
	}
}

func TestDebugLogging(t *testing.T) {
	setEnvVars(environmentValues)
	defer unsetEnvVars(environmentValues)

	TestCases := []struct {
		description     string
		envDebugValue   string
		annotDebugValue string
	}{
		{
			description:     "debug logs are enabled with 'DEBUG=true' envvar",
			envDebugValue:   "true",
			annotDebugValue: "",
		},
		{
			description:     "debug logs are disabled if 'DEBUG' envvar does not equal 'true'",
			envDebugValue:   "invalid value",
			annotDebugValue: "",
		},
		{
			description:     "annotation 'conjur.org/debug-logging' takes precedence over 'DEBUG' envvar",
			envDebugValue:   "invalid value",
			annotDebugValue: "true",
		},
	}

	setEnvVars(environmentValues)
	defer unsetEnvVars(environmentValues)

	for _, tc := range TestCases {
		_ = os.Setenv("DEBUG", tc.envDebugValue)

		settings := GatherSettings(map[string]string{
			"conjur.org/debug-logging": tc.annotDebugValue,
		})

		errLogs, infoLogs := ValidateSettings(settings, successfulMockReadFile)
		assert.Equal(t, 0, len(errLogs))
		assert.Equal(t, 0, len(infoLogs))

		// Replace logger with buffer to test its value
		var logBuffer bytes.Buffer
		logger.InfoLogger = log.New(&logBuffer, "", 0)

		NewConfig(settings)

		logMessages := logBuffer.String()
		if tc.envDebugValue == "invalid value" && tc.annotDebugValue == "" {
			assert.Contains(t, logMessages, "WARN")
			assert.Contains(t, logMessages, "CAKC034")
		} else {
			assert.NotContains(t, logMessages, "CAKC034")
			assert.Contains(t, logMessages, "DEBUG")
			assert.Contains(t, logMessages, "CAKC052")
		}
	}
}

func TestConjurVersion(t *testing.T) {
	setEnvVars(environmentValues)
	defer unsetEnvVars(environmentValues)

	TestCases := []struct {
		description string
		envVersion  string
		expVersion  string
		expErrStr   string
	}{
		{
			description: "Succeeds if version is 4",
			envVersion:  "4",
			expVersion:  "4",
			expErrStr:   "",
		},
		{
			description: "Succeeds if version is 5",
			envVersion:  "5",
			expVersion:  "5",
			expErrStr:   "",
		},
		{
			description: "Sets the default version for an empty value",
			envVersion:  "",
			expVersion:  DefaultConjurVersion,
			expErrStr:   "",
		},
		{
			description: "Returns error if version is invalid",
			envVersion:  "3",
			expVersion:  "",
			expErrStr:   "invalid conjur version",
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			_ = os.Setenv("CONJUR_VERSION", tc.envVersion)

			var logBuffer bytes.Buffer
			if tc.expErrStr != "" {
				// Replace logger with buffer to test its value
				logger.ErrorLogger = log.New(&logBuffer, "", 0)
			}

			config, err := FromEnv(successfulMockReadFile)

			if tc.expErrStr == "" {
				assert.Nil(t, err)
				assert.Equal(t, tc.expVersion, config.ConjurVersion)
			} else {
				assert.NotNil(t, err)
				errorMessages := logBuffer.String()
				assert.Contains(t, errorMessages, tc.expErrStr)
			}

			_ = os.Setenv("CONJUR_VERSION", "")
		})
	}
}

func successfulMockReadFile(filename string) ([]byte, error) {
	return []byte{}, nil
}
