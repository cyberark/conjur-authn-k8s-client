package authenticator

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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
}
