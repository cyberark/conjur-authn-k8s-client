package jwt

import (
	"testing"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token/memory"
	"github.com/stretchr/testify/require"
)

func TestNewWithAccessToken(t *testing.T) {
	testCases := []struct {
		name            string
		url             string
		expectedBaseURL string
		shouldError     bool
	}{
		{
			name:            "Conjur Edge URL with /api prefix",
			url:             "https://edge.example.com/api/authn-jwt/my-service",
			expectedBaseURL: "https://edge.example.com/api",
			shouldError:     false,
		},
		{
			name:            "Conjur URL without /api prefix",
			url:             "https://conjur.example.com/authn-jwt/my-service",
			expectedBaseURL: "https://conjur.example.com",
			shouldError:     false,
		},
		{
			name:            "Proxy URL with custom path",
			url:             "https://proxy.example.com/custom/path/authn-jwt/my-service",
			expectedBaseURL: "https://proxy.example.com/custom/path",
			shouldError:     false,
		},
		{
			name:            "Conjur URL with trailing slash",
			url:             "https://conjur.example.com/authn-jwt/my-service/",
			expectedBaseURL: "https://conjur.example.com",
			shouldError:     false,
		},
		{
			name:            "Conjur URL with multiple slashes",
			url:             "https://conjur.example.com//authn-jwt//my-service",
			expectedBaseURL: "https://conjur.example.com",
			shouldError:     false,
		},
		{
			name:            "URL with port number",
			url:             "https://conjur.example.com:8443/authn-jwt/my-service",
			expectedBaseURL: "https://conjur.example.com:8443",
			shouldError:     false,
		},
		{
			name:        "Invalid URL - missing service ID",
			url:         "https://conjur.example.com/authn-jwt",
			shouldError: true,
		},
		{
			name:        "Invalid URL - wrong authenticator type",
			url:         "https://conjur.example.com/authn-k8s/my-service",
			shouldError: true,
		},
		{
			name:        "Invalid URL - malformed",
			url:         "://invalid-url",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := Config{}
			config.Common.URL = tc.url
			config.Common.Account = "conjur"
			config.Common.SSLCertificate = []byte("")
			config.JWTTokenFilePath = "my-jwt-token"

			at, _ := memory.NewAccessToken()

			authn, err := NewWithAccessToken(config, at)

			if tc.shouldError {
				require.Error(t, err)
				require.Nil(t, authn)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, authn)

			// Verify the base URL was constructed correctly
			req, err := authn.conjurClient.JWTAuthenticateRequest("my-jwt-token", "host-id")
			require.NoError(t, err)
			require.NotNil(t, req)

			expectedURL := tc.expectedBaseURL + "/authn-jwt/my-service/conjur/host-id/authenticate"
			require.Equal(t, expectedURL, req.URL.String())
		})
	}
}
