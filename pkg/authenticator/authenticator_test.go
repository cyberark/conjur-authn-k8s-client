package authenticator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token/memory"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
)

func parseCert(filename string) (*x509.Certificate, error) {
	certPEMBlock, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	certDERBlock, certPEMBlock := pem.Decode(certPEMBlock)
	cert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func TestAuthenticator_GenerateCSR(t *testing.T) {
	// SETUP
	// Create a minimal authenticator with a minimal config
	authnConfig := config.Config{
		PodName:      "testPodName",
		PodNamespace: "testPodNamespace",
	}
	signingKey, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.NoError(t, err)

	authn := &Authenticator{
		Config:     authnConfig,
		privateKey: signingKey,
	}

	// EXERCISE
	// Generate the CSR
	csr, err := authn.GenerateCSR("host.path.to.policy")
	assert.NoError(t, err)

	// ASSERT
	// Parse the generated CSR using the stdlib to ensure
	// it has the properties we expect
	parsedCsr, err := x509.ParseCertificateRequest(csr)
	assert.NoError(t, err)

	// Assert on common name
	assert.Equal(t, "host.path.to.policy", parsedCsr.Subject.CommonName)

	// Assert on spiffe SAN
	assert.Len(t, parsedCsr.Extensions, 1)
	var sans []asn1.RawValue
	_, _ = asn1.Unmarshal(parsedCsr.Extensions[0].Value, &sans)
	assert.Len(t, sans, 1)
	assert.Equal(
		t,
		"spiffe://cluster.local/namespace/testPodNamespace/podname/testPodName",
		string(sans[0].Bytes),
	)
}

func TestAuthenticator_IsCertExpired(t *testing.T) {
	t.Run("Active cert", func(t *testing.T) {
		// SETUP
		activeCert, err := parseCert("testdata/good_cert.crt")
		assert.NoError(t, err)
		authn := Authenticator{
			PublicCert: activeCert,
		}

		// EXERCISE
		isExpired := authn.IsCertExpired()

		// ASSERT
		assert.False(t, isExpired)
	})

	t.Run("Expired cert", func(t *testing.T) {
		// SETUP
		expiredCert, err := parseCert("testdata/expired_cert.crt")
		assert.NoError(t, err)
		authn := Authenticator{
			PublicCert: expiredCert,
		}

		// EXERCISE
		isExpired := authn.IsCertExpired()

		// ASSERT
		assert.True(t, isExpired)
	})
}

func Test_consumeInectClientCertError(t *testing.T) {
	// SETUP
	path := "/tmp/test_file"
	expectedErr := "some\ntext\n"
	err := ioutil.WriteFile(path, []byte(expectedErr), 0644)
	assert.NoError(t, err)

	// EXERCISE
	consumedErr := consumeInjectClientCertError(path)

	// ASSERT
	assert.Equal(t, expectedErr, consumedErr)
	assert.NoFileExists(t, path)

}

func TestAuthenticator_Authenticate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// Create a temporary file for storing the client cert. This will allow multiple tests to run in parallel
		tmpDir := t.TempDir()
		clientCertPath := filepath.Join(tmpDir, "etc:conjur:ssl:client.pem")
		certLogPath := filepath.Join(tmpDir, "tmp:conjur_copy_text_output.log")
		tokenPath := filepath.Join(tmpDir, "run:conjur:access-token")

		// Start up a test server to mock the Conjur server's auth endpoints
		ts := testServer(clientCertPath, "some token")
		defer ts.Close()

		// Create an authenticator with dummy config
		at, _ := memory.NewAccessToken()
		sslcert := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: ts.Certificate().Raw,
		})
		username, _ := config.NewUsername("host/test-user")

		cfg := config.Config{
			Account:                   "account",
			ClientCertPath:            clientCertPath,
			ClientCertRetryCountLimit: 1,
			ContainerMode:             "doesntmatter",
			ConjurVersion:             "5",
			InjectCertLogPath:         certLogPath,
			PodName:                   "somepodname",
			PodNamespace:              "somepodnamespace",
			SSLCertificate:            sslcert,
			TokenFilePath:             tokenPath,
			TokenRefreshTimeout:       0,
			URL:                       ts.URL,
			Username:                  username,
		}

		authn, err := NewWithAccessToken(cfg, at)
		if !assert.NoError(t, err) {
			return
		}

		err = authn.Authenticate()
		if !assert.NoError(t, err) {
			return
		}

		// Check that the access token was set correctly
		token, _ := authn.AccessToken.Read()
		assert.Equal(t, token, []byte("some token"))
	})
}
