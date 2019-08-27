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

		Convey("Given a non-expired certificate to authenticate with Conjur", func() {
			goodCert, err := parseCert("testdata/good_cert.crt")

			Convey("Finishes without raising an error and returns certificate", func() {
				So(err, ShouldBeNil)
			})

			Convey("Returns the false that the certificate is not expired", func() {
				authn := Authenticator{
					PublicCert: goodCert,
				}
				So(authn.IsCertExpired(), ShouldEqual, false)
			})
		})

		Convey("Given an expired certificate to authenticate with Conjur", func() {
			expiredCert, err := parseCert("testdata/expired_cert.crt")

			Convey("Finishes without raising an error and returns certificate", func() {
				So(err, ShouldBeNil)
			})

			Convey("Returns true that the certificate is expired", func() {
				authn := Authenticator{
					PublicCert: expiredCert,
				}
				So(authn.IsCertExpired(), ShouldEqual, true)
			})
		})
	})
}
