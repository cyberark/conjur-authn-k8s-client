package authenticator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

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
}
