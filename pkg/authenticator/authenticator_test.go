package authenticator

import (
	"bytes"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token/memory"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

type assertFunc func(t *testing.T,
	authn *Authenticator,
	err error,
	loginCsr *x509.CertificateRequest,
	loginCsrErr error,
	logTxt string,
)

func TestAuthenticator_Authenticate(t *testing.T) {
	testCases := []struct {
		name               string
		podName            string
		podNamespace       string
		skipWritingCSRFile bool
		assert             assertFunc
	}{
		{
			name:         "happy path",
			podName:      "testPodName",
			podNamespace: "testPodNamespace",
			assert: func(t *testing.T, authn *Authenticator, err error, loginCsr *x509.CertificateRequest, loginCsrErr error, _ string) {
				assert.NoError(t, err)

				// Check the CSR
				assert.NotNil(t, loginCsr)
				assert.NoError(t, loginCsrErr)

				// Assert on common name
				assert.Equal(t, "test-user", loginCsr.Subject.CommonName)

				// Assert on spiffe SAN
				assert.Len(t, loginCsr.Extensions, 1)
				var sans []asn1.RawValue
				_, _ = asn1.Unmarshal(loginCsr.Extensions[0].Value, &sans)
				assert.Len(t, sans, 1)

				assert.Equal(
					t,
					"spiffe://cluster.local/namespace/testPodNamespace/podname/testPodName",
					string(sans[0].Bytes),
				)

				// Check that the access token was set correctly
				token, _ := authn.AccessToken.Read()
				assert.Equal(t, token, []byte("some token"))
			},
		},
		{
			name:         "empty podname",
			podName:      "",
			podNamespace: "",
			assert: func(t *testing.T, authn *Authenticator, err error, loginCsr *x509.CertificateRequest, _ error, _ string) {
				assert.NoError(t, err)

				// Assert empty spiffe
				assert.Len(t, loginCsr.Extensions, 1)
				var sans []asn1.RawValue
				_, _ = asn1.Unmarshal(loginCsr.Extensions[0].Value, &sans)
				assert.Len(t, sans, 1)
				assert.Equal(t, "", string(sans[0].Bytes))
			},
		},
		{
			name:         "expired cert",
			podName:      "testPodName",
			podNamespace: "testPodNamespace",
			assert: func(t *testing.T, authn *Authenticator, err error, _ *x509.CertificateRequest, _ error, _ string) {
				assert.NoError(t, err)
				// Set the expiration date to now, and try to authenticate again
				// This will cause the authenticator to try to refresh the cert
				authn.PublicCert.NotAfter = time.Now()
				err = authn.Authenticate()
				assert.NoError(t, err)

				// Check that the cert was renewed
				assert.True(t, authn.PublicCert.NotAfter.After(time.Now()))
			},
		},
		{
			name:               "injects log on failure",
			podName:            "testPodName",
			podNamespace:       "testPodNamespace",
			skipWritingCSRFile: true,
			assert: func(t *testing.T, _ *Authenticator, err error, _ *x509.CertificateRequest, _ error, logTxt string) {
				assert.Error(t, err)
				// Check logs for the expected error
				assert.Contains(t, logTxt, "error writing csr file")
			},
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
			var loginCsr *x509.CertificateRequest
			var loginCsrErr error

			// Start up a test server to mock the Conjur server's auth endpoints
			ts := NewTestAuthServer(clientCertPath, certLogPath, "some token", tc.skipWritingCSRFile)
			ts.HandleLogin = func(csr *x509.CertificateRequest, err error) {
				loginCsr = csr
				loginCsrErr = err
			}
			defer ts.Server.Close()

			// Create an authenticator with dummy config
			at, _ := memory.NewAccessToken()
			sslcert := pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: ts.Server.Certificate().Raw,
			})
			username, _ := config.NewUsername("host/test-user")

			cfg := config.Config{
				Account:                   "account",
				ClientCertPath:            clientCertPath,
				ClientCertRetryCountLimit: 0,
				ContainerMode:             "doesntmatter",
				ConjurVersion:             "5",
				InjectCertLogPath:         certLogPath,
				PodName:                   tc.podName,
				PodNamespace:              tc.podNamespace,
				SSLCertificate:            sslcert,
				TokenFilePath:             tokenPath,
				TokenRefreshTimeout:       0,
				URL:                       ts.Server.URL,
				Username:                  username,
			}

			// EXERCISE
			authn, err := NewWithAccessToken(cfg, at)
			if !assert.NoError(t, err) {
				return
			}

			// Intercept the logs to check for the cert placement error
			var logTxt bytes.Buffer
			log.ErrorLogger.SetOutput(&logTxt)

			// Call the main method of the authenticator. This is where most of the internal implementation happens
			err = authn.Authenticate()

			// ASSERT
			tc.assert(t, authn, err, loginCsr, loginCsrErr, logTxt.String())
		})
	}
}
