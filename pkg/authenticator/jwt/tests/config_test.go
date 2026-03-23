package tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/jwt"
)

var environmentValues = map[string]string{
	"CONJUR_AUTHN_URL":       "authn-jwt",
	"CONJUR_ACCOUNT":         "testAccount",
	"CONJUR_CERT_FILE":       "testSSLCertFile.txt",
	"CONJUR_SSL_CERTIFICATE": "testSSLCert",
	"JWT_TOKEN_PATH":         "good_jwt.token",
}

var annotationValues = map[string]string{
	"conjur.org/log-level":      "debug",
	"conjur.org/container-mode": "init",
}

var envToAnnot = map[string]string{
	"LOG_LEVEL":      "conjur.org/log-level",
	"CONTAINER_MODE": "conjur.org/container-mode",
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
				"JWT_TOKEN_PATH":          "good_jwt.token",
				"CONJUR_AUTHN_LOGIN":      "",
				"CONJUR_ACCOUNT":          "testAccount",
				"CONJUR_AUTHN_URL":        "authn-jwt",
				"CONJUR_CERT_FILE":        "testSSLCertFile.txt",
				"CONJUR_SSL_CERTIFICATE":  "testSSLCert",
				"CONTAINER_MODE":          "init",  // provided by annotation
				"LOG_LEVEL":               "debug", // provided by annotation
				"DEBUG":                   "",
				"CONJUR_AUTHN_TOKEN_FILE": jwt.DefaultTokenFilePath,
				"CONJUR_HTTP_TIMEOUT":     jwt.DefaultHTTPTimeout,
				"CONJUR_TOKEN_TIMEOUT":    jwt.DefaultTokenRefreshTimeout,
			},
		},
		{
			description: "if the first getter function returns empty strings, fallback to the next functions, and eventually an empty string",
			annotFunc:   emptyAnnotations,
			expected: config.AuthnSettings{
				"JWT_TOKEN_PATH":          "good_jwt.token",
				"CONJUR_AUTHN_LOGIN":      "",
				"CONJUR_AUTHN_URL":        "authn-jwt",
				"CONJUR_ACCOUNT":          "testAccount",
				"CONJUR_CERT_FILE":        "testSSLCertFile.txt",
				"CONJUR_SSL_CERTIFICATE":  "testSSLCert",
				"LOG_LEVEL":               "",
				"DEBUG":                   "",
				"CONTAINER_MODE":          "",
				"CONJUR_AUTHN_TOKEN_FILE": jwt.DefaultTokenFilePath,
				"CONJUR_HTTP_TIMEOUT":     jwt.DefaultHTTPTimeout,
				"CONJUR_TOKEN_TIMEOUT":    jwt.DefaultTokenRefreshTimeout,
			},
		},
	}

	for _, tc := range TestCases {
		t.Run(tc.description, func(t *testing.T) {
			resultMap := config.GatherSettings(&jwt.Config{}, tc.annotFunc, fromEnv)
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
				"JWT_TOKEN_PATH":         "good_jwt.token",
				"CONJUR_AUTHN_URL":       "authn-jwt",
				"CONJUR_ACCOUNT":         "testAccount",
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

func TestHTTPTimeoutConfiguration(t *testing.T) {
	testCases := []struct {
		description        string
		httpTimeout        string
		expectedError      bool
		expectedDurationMs int64
	}{
		{
			description:        "default timeout value when not set",
			httpTimeout:        "",
			expectedError:      false,
			expectedDurationMs: 60000, // 60 seconds
		},
		{
			description:        "valid timeout value",
			httpTimeout:        "30s",
			expectedError:      false,
			expectedDurationMs: 30000, // 30 seconds
		},
		{
			description:        "valid timeout with minutes",
			httpTimeout:        "1m30s",
			expectedError:      false,
			expectedDurationMs: 90000, // 90 seconds
		},
		{
			description:   "invalid timeout value returns error",
			httpTimeout:   "invalid",
			expectedError: true,
		},
	}

	baseEnv := map[string]string{
		"JWT_TOKEN_PATH":         "good_jwt.token",
		"CONJUR_AUTHN_URL":       "authn-jwt",
		"CONJUR_ACCOUNT":         "testAccount",
		"CONJUR_SSL_CERTIFICATE": "samplecert",
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			env := baseEnv
			if tc.httpTimeout != "" {
				env["CONJUR_HTTP_TIMEOUT"] = tc.httpTimeout
			}
			setEnv(env)

			cfg, err := config.ConfigFromEnv(successfulMockReadFile)

			if tc.expectedError {
				assert.Nil(t, cfg)
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, cfg)
				assert.Nil(t, err)
				assert.Equal(t, tc.expectedDurationMs, cfg.GetHTTPTimeout().Milliseconds())
			}

			unsetEnv(env)
		})
	}
}
