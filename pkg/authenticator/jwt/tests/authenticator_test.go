// +build integration

package tests

import (
	"bytes"
	"context"
	"encoding/pem"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token/memory"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/jwt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

const tmpJwtTokenPath = "good_jwt.token"

type assertFunc func(t *testing.T,
	authn *jwt.Authenticator,
	err error,
)

func TestAuthenticator_Authenticate(t *testing.T) {
	testCases := []struct {
		name               string
		jwtTokenPath       string
		assert             assertFunc
		skipWritingCSRFile bool
		wrongUrl           bool
	}{
		{
			name:         "happy path",
			jwtTokenPath: tmpJwtTokenPath,
			assert: func(t *testing.T, authn *jwt.Authenticator, err error) {
				assert.NoError(t, err)

				// Check that the access token was set correctly
				token, _ := authn.GetAccessToken().Read()
				assert.Equal(t, token, []byte("some token"))
			},
		},
		{
			name:         "wrong url given",
			jwtTokenPath: tmpJwtTokenPath,
			assert: func(t *testing.T, authn *jwt.Authenticator, err error) {
				assert.NotNil(t, err)
				assert.True(t, strings.Contains(err.Error(), "Failed to send https authenticate request or receive response"))
			},
			wrongUrl: true,
		},
		{
			name: "token doesn't exist",
			assert: func(t *testing.T, authn *jwt.Authenticator, err error) {
				assert.NotNil(t, err)
				assert.True(t, strings.Contains(err.Error(), "Failed to read JWT from"))
			},
			jwtTokenPath: "/tmp/nonExistingPath",
		},
		{
			name: "Token path is empty",
			assert: func(t *testing.T, authn *jwt.Authenticator, err error) {
				assert.NotNil(t, err)
				assert.True(t, strings.Contains(err.Error(), "Failed to read JWT from"))
			},
			jwtTokenPath: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// SETUP
			// Create a temporary file for storing the client cert. This will allow multiple tests to run in parallel
			tmpDir := t.TempDir()
			clientCertPath := filepath.Join(tmpDir, "etc:conjur:ssl:client.pem")
			certLogPath := filepath.Join(tmpDir, "tmp:conjur_copy_text_output.log")
			tokenPath := filepath.Join(tmpDir, "run:conjur:access-token")

			// Start up a test server to mock the Conjur server's auth endpoints
			ts := common.NewTestAuthServer(clientCertPath, certLogPath, "some token", tc.skipWritingCSRFile)

			defer ts.Server.Close()

			// Create an authenticator with dummy config
			at, _ := memory.NewAccessToken()

			sslcert := pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: ts.Server.Certificate().Raw,
			})

			cfg := jwt.Config{
				JWTTokenFilePath: tc.jwtTokenPath,
				Common: common.Config{
					SSLCertificate:            sslcert,
					TokenFilePath:             tokenPath,
					TokenRefreshTimeout:       0,
					URL:                       ts.Server.URL,
					Username:                  nil,
					Account:                   "account",
					ClientCertPath:            clientCertPath,
					ClientCertRetryCountLimit: 0,
					ContainerMode:             "doesntmatter",
				},
			}

			if tc.wrongUrl {
				cfg.Common.URL = "http://wrong-url"
			}

			// EXERCISE
			authn, err := jwt.NewWithAccessToken(cfg, at)
			if !assert.NoError(t, err) {
				return
			}

			// Intercept the logs to check for the cert placement error
			var logTxt bytes.Buffer
			log.ErrorLogger.SetOutput(&logTxt)

			// Call the main method of the authenticator. This is where most of the internal implementation happens
			err = authn.AuthenticateWithContext(context.Background())

			// ASSERT
			tc.assert(t, authn, err)
		})
	}
}
