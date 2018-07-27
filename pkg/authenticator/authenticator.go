package authenticator

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/fullsailor/pkcs7"
)

var oidExtensionSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}

// AuthenticatorConfig defines the configuration parameters
// for the authentication requests
type AuthenticatorConfig struct {
	ConjurVersion  string
	Account        string
	URL            string
	Username       string
	PodName        string
	PodNamespace   string
	SSLCertificate []byte
	ClientCertPath string
	TokenFilePath  string
}

// Authenticator contains the configuration and client
// for the authentication connection to Conjur
type Authenticator struct {
	Config     AuthenticatorConfig
	privateKey *rsa.PrivateKey
	publicCert *x509.Certificate
	client     *http.Client
}

const (
	nameTypeEmail = 1
	nameTypeDNS   = 2
	nameTypeURI   = 6
	nameTypeIP    = 7
)

// New returns a new Authenticator
func New(config AuthenticatorConfig) (auth *Authenticator, err error) {
	signingKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}

	client, err := newHTTPSClient(config.SSLCertificate, nil, nil)
	if err != nil {
		return nil, err
	}

	return &Authenticator{
		Config:     config,
		client:     client,
		privateKey: signingKey,
	}, nil
}

// GenerateCSR prepares the CSR
func (auth *Authenticator) GenerateCSR() ([]byte, error) {
	sanURIString, err := generateSANURI(auth.Config.PodNamespace, auth.Config.PodName)
	sanURI, err := url.Parse(sanURIString)
	if err != nil {
		return nil, err
	}

	// The CSR only uses the :namespace/:resource_type/:resource_id part of the username
	usernameSplit := strings.Split(auth.Config.Username, "/")
	usernameCSR := strings.Join(usernameSplit[len(usernameSplit)-3:], "/")

	return generateCSR(auth.privateKey, usernameCSR, sanURI)
}

// Login sends Conjur a CSR and verifies that the client cert is
// successfully retrieved
func (auth *Authenticator) Login() error {

	log.Printf(fmt.Sprintf("logging in as %s.", auth.Config.Username))

	csrRawBytes, err := auth.GenerateCSR()

	csrBytes := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrRawBytes,
	})
	req, err := LoginRequest(auth.Config.URL, auth.Config.ConjurVersion, csrBytes)
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
	certPEMBlock, err := ioutil.ReadFile(auth.Config.ClientCertPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("client certificate not found at %s", auth.Config.ClientCertPath)
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
	os.Remove(auth.Config.ClientCertPath)

	return nil
}

// Authenticate sends Conjur an authenticate request and returns
// the response data
func (auth *Authenticator) Authenticate() ([]byte, error) {
	privDer := x509.MarshalPKCS1PrivateKey(auth.privateKey)
	keyPEMBlock := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDer})

	certPEMBlock := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: auth.publicCert.Raw})

	client, err := newHTTPSClient(auth.Config.SSLCertificate, certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, err
	}

	req, err := AuthenticateRequest(
		auth.Config.URL,
		auth.Config.ConjurVersion,
		auth.Config.Account,
		auth.Config.Username,
		certPEMBlock,
	)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return DataResponse(resp)
}

// ParseAuthenticationResponse takes the response from the Authenticate
// request, decrypts if needed, and writes to the token file
func (auth *Authenticator) ParseAuthenticationResponse(response []byte) error {

	var content []byte
	var err error

	// Token is only encrypted in Conjur v4
	if auth.Config.ConjurVersion == "4" {
		content, err = decodeFromPEM(response, auth.publicCert, auth.privateKey)
		if err != nil {
			return err
		}
	} else if auth.Config.ConjurVersion == "5" {
		content = response
	}

	// log.Printf("writing token %v to shared volume ...", content)
	err = ioutil.WriteFile(auth.Config.TokenFilePath, content, 0644)
	if err != nil {
		return err
	}

	log.Printf("successfully authenticated.")

	return nil
}

// generateSANURI returns the formatted uri(SPIFFEE format for now) for the certificate.
func generateSANURI(namespace, podname string) (string, error) {
	if namespace == "" || podname == "" {
		return "", fmt.Errorf(
			"namespace or podname can't be empty namespace=%v podname=%v", namespace, podname)
	}
	return fmt.Sprintf("spiffe://cluster.local/namespace/%s/podname/%s", namespace, podname), nil
}

func marshalSANs(dnsNames, emailAddresses []string, ipAddresses []net.IP, uris []*url.URL) (derBytes []byte, err error) {
	var rawValues []asn1.RawValue
	for _, name := range dnsNames {
		rawValues = append(rawValues, asn1.RawValue{Tag: nameTypeDNS, Class: asn1.ClassContextSpecific, Bytes: []byte(name)})
	}
	for _, email := range emailAddresses {
		rawValues = append(rawValues, asn1.RawValue{Tag: nameTypeEmail, Class: asn1.ClassContextSpecific, Bytes: []byte(email)})
	}
	for _, rawIP := range ipAddresses {
		// If possible, we always want to encode IPv4 addresses in 4 bytes.
		ip := rawIP.To4()
		if ip == nil {
			ip = rawIP
		}
		rawValues = append(rawValues, asn1.RawValue{Tag: nameTypeIP, Class: asn1.ClassContextSpecific, Bytes: ip})
	}
	for _, uri := range uris {
		rawValues = append(rawValues, asn1.RawValue{Tag: nameTypeURI, Class: asn1.ClassContextSpecific, Bytes: []byte(uri.String())})
	}
	return asn1.Marshal(rawValues)
}

func decodeFromPEM(PEMBlock []byte, publicCert *x509.Certificate, privateKey crypto.PrivateKey) ([]byte, error) {
	var decodedPEM []byte

	tokenDerBlock, _ := pem.Decode(PEMBlock)
	p7, err := pkcs7.Parse(tokenDerBlock.Bytes)
	if err != nil {
		return nil, err
	}

	decodedPEM, err = p7.Decrypt(publicCert, privateKey)
	if err != nil {
		return nil, err
	}

	return decodedPEM, nil
}

func generateCSR(privateKey *rsa.PrivateKey, id string, sanURI *url.URL) ([]byte, error) {
	commonName := strings.Replace(id, "/", ".", -1)
	subj := pkix.Name{
		CommonName: commonName,
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	subjectAltNamesValue, err := marshalSANs(nil, nil, nil, []*url.URL{
		sanURI,
	})
	if err != nil {
		return nil, err
	}

	extSubjectAltName := pkix.Extension{
		Id:       oidExtensionSubjectAltName,
		Critical: false,
		Value:    subjectAltNamesValue,
	}
	template.ExtraExtensions = []pkix.Extension{extSubjectAltName}

	return x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
}
