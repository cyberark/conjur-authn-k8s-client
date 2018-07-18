package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var oidExtensionSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}

type AuthenticatorConfig struct {
	ConjurVersion  string
	Account        string
	URL            string
	Username       string
	PodName        string
	PodNamespace   string
	SSLCertificate []byte
}

type Authenticator struct {
	AuthenticatorConfig
	privateKey *rsa.PrivateKey
	publicCert *x509.Certificate
	client     *http.Client
}

func NewAuthenticator(config AuthenticatorConfig) (auth *Authenticator, err error) {
	signingKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}

	client, err := newHTTPSClient(config.SSLCertificate, nil, nil)
	if err != nil {
		return nil, err
	}

	return &Authenticator{AuthenticatorConfig: config, client: client, privateKey: signingKey}, nil
}

func (auth *Authenticator) GenerateCSR() ([]byte, error) {
	sanURIString, err := generateSANURI(auth.PodNamespace, auth.PodName)
	sanURI, err := url.Parse(sanURIString)
	if err != nil {
		return nil, err
	}

	// The CSR only uses the :namespace/:resource_type/:resource_id part of the username
	usernameSplit := strings.Split(auth.Username, "/")
	usernameCSR := strings.Join(usernameSplit[len(usernameSplit)-3:], "/")

	return generateCSR(auth.privateKey, usernameCSR, sanURI)
}

func (auth *Authenticator) Login() error {
	csrRawBytes, err := auth.GenerateCSR()

	csrBytes := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrRawBytes,
	})
	req, err := LoginRequest(auth.URL, auth.ConjurVersion, csrBytes)
	if err != nil {
		return err
	}

	resp, err := auth.client.Do(req)

	if err != nil {
		return err
	}

	err = EmptyResponse(resp)
	if err != nil {
		return err
	}
	// load client cert
	certPEMBlock, err := ioutil.ReadFile(CLIENT_CERT_PATH)
	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("client certificate not found at %s", CLIENT_CERT_PATH)
		}

		return err
	}

	certDERBlock, certPEMBlock := pem.Decode(certPEMBlock)
	cert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return err
	}

	auth.publicCert = cert

	// clean up the client cert so it's only available in memory
	os.Remove(CLIENT_CERT_PATH)

	return nil
}

func (auth *Authenticator) Authenticate() ([]byte, error) {
	privDer := x509.MarshalPKCS1PrivateKey(auth.privateKey)
	keyPEMBlock := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDer})

	certPEMBlock := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: auth.publicCert.Raw})

	client, err := newHTTPSClient(auth.SSLCertificate, certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, err
	}

	req, err := AuthenticateRequest(auth.URL, auth.ConjurVersion, auth.Account, auth.Username, certPEMBlock)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return DataResponse(resp)
}
