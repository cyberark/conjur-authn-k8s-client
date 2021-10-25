package authenticator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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

func TestAuthenticator(t *testing.T) {
	Convey("IsCertExpired", t, func() {
		Convey("Returns false if cert is not expired", func() {
			goodCert, err := parseCert("testdata/good_cert.crt")
			So(err, ShouldBeNil)

			authn := Authenticator{
				PublicCert: goodCert,
			}

			So(authn.IsCertExpired(), ShouldEqual, false)
		})

		Convey("Returns true if cert is expired", func() {
			expiredCert, err := parseCert("testdata/expired_cert.crt")
			So(err, ShouldBeNil)

			authn := Authenticator{
				PublicCert: expiredCert,
			}

			So(authn.IsCertExpired(), ShouldEqual, true)
		})
	})

	Convey("GenerateCSR", t, func() {
		// Create a minimal authenticator with a minimal config
		authnConfig := config.Config{
			PodName:      "testPod",
			PodNamespace: "testNameSpace",
		}

		signingKey, _ := rsa.GenerateKey(rand.Reader, 1024)
		authn := &Authenticator{
			Config:     authnConfig,
			privateKey: signingKey,
		}

		Convey("Given a common-name", func() {
			commonName := "host.path.to.policy"
			csr, err := authn.GenerateCSR(commonName)
			Convey("Finishes without raising an error", func() {
				So(err, ShouldBeNil)
			})

			// decrypt the CSR
			csrDecrypted, _ := x509.ParseCertificateRequest(csr)
			Convey("Inserts the common-name in the subject", func() {
				So(csrDecrypted.Subject.CommonName, ShouldEqual, commonName)
			})
		})
	})

	Convey("consumeInjectClientCertError", t, func() {
		path := "/tmp/test_file"
		dataStr := "some\ntext\n"
		err := ioutil.WriteFile(path, []byte(dataStr), 0644)
		if err != nil {
			t.FailNow()
		}

		content := consumeInjectClientCertError(path)

		Convey("Gets the content from the file", func() {
			So(content, ShouldResemble, dataStr)
		})

		Convey("Deletes the file", func() {
			_, err := os.Stat(path)
			So(os.IsNotExist(err), ShouldBeTrue)
		})
	})

	t.Run("retrieves access token", func(t *testing.T) {
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
