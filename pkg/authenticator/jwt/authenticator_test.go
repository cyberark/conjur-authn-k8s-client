package jwt

import (
	"testing"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token/memory"
	"github.com/stretchr/testify/require"
)

func TestEdgeUrl(t *testing.T) {
	t.Run("with a conjur edge url", func(t *testing.T) {
		config := Config{}
		config.Common.URL = "https://edge.example.com/api/authn-jwt/service-id"
		config.Common.Account = "conjur"
		config.Common.SSLCertificate = []byte("")
		config.JWTTokenFilePath = "my-jwt-token"

		at, _ := memory.NewAccessToken()

		authn, err := NewWithAccessToken(config, at)

		require.NoError(t, err)
		require.NotNil(t, authn)

		t.Run("includes /api", func(t *testing.T) {
			req, err := authn.conjurClient.JWTAuthenticateRequest("my-jwt-token", "host-id")
			require.NoError(t, err)
			require.NotNil(t, req)
			// Ensure /api isn't stripped
			require.Contains(t, req.URL.String(), "/api/")
		})
	})
}
