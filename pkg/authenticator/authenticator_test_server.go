package authenticator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

// testServer creates, for testing purposes, an http server on a random port that mocks conjur's
// login and authenticate endpoints.
func testServer(clientCertPath, expectedTokenValue string) *httptest.Server {
	authnCACertificate := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
	}
	authnCAPrivKey, _ := rsa.GenerateKey(rand.Reader, 4096)

	var loginCsr *x509.CertificateRequest
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.HasSuffix(r.URL.Path, "/authenticate") {
			// Peer certificate from mutual auth

			// Respond with a dummy token
			w.WriteHeader(201)
			w.Write([]byte(expectedTokenValue))
		}

		if strings.HasPrefix(r.URL.Path, "/inject_client_cert") {
			// Login request

			body, _ := ioutil.ReadAll(r.Body)
			defer r.Body.Close()

			csrBlock, _ := pem.Decode(body)
			loginCsr, _ = x509.ParseCertificateRequest(csrBlock.Bytes)

			w.WriteHeader(201)

			// Create client certificate template
			clientCRTTemplate := x509.Certificate{
				Signature:          loginCsr.Signature,
				SignatureAlgorithm: loginCsr.SignatureAlgorithm,

				PublicKeyAlgorithm: loginCsr.PublicKeyAlgorithm,
				PublicKey:          loginCsr.PublicKey,

				SerialNumber: big.NewInt(2),
				Issuer:       authnCACertificate.Subject,
				Subject:      loginCsr.Subject,
				NotBefore:    time.Now(),
				NotAfter:     time.Now().Add(24 * time.Hour),
				KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
				ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			// Create client certificate from template and CA public key
			clientCRTRaw, _ := x509.CreateCertificate(
				rand.Reader,
				&clientCRTTemplate,
				authnCACertificate,
				loginCsr.PublicKey,
				authnCAPrivKey,
			)
			// Save the certificate
			err := ioutil.WriteFile(
				clientCertPath,
				pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCRTRaw}),
				os.ModePerm,
			)
			if err != nil {
				panic(err)
			}
		}

	}))
	ts.StartTLS()
	ts.TLS.ClientAuth = tls.RequestClientCert

	return ts
}
