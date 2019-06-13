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
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fullsailor/pkcs7"

	authnConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
)

var oidExtensionSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
var bufferTime = 30 * time.Second

// Authenticator contains the configuration and client
// for the authentication connection to Conjur
type Authenticator struct {
	Config     authnConfig.Config
	privateKey *rsa.PrivateKey
	PublicCert *x509.Certificate
	client     *http.Client
}

const (
	nameTypeEmail = 1
	nameTypeDNS   = 2
	nameTypeURI   = 6
	nameTypeIP    = 7
)

// New returns a new Authenticator
func New(config authnConfig.Config) (auth *Authenticator, err error) {
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

	// Generate CSR
	commonName := strings.Replace(usernameCSR, "/", ".", -1)
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

	return x509.CreateCertificateRequest(rand.Reader, &template, auth.privateKey)
}

// Login sends Conjur a CSR and verifies that the client cert is
// successfully retrieved
func (auth *Authenticator) Login() error {

	InfoLogger.Printf(fmt.Sprintf("Logging in as %s.", auth.Config.Username))

	csrRawBytes, err := auth.GenerateCSR()

	csrBytes := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrRawBytes,
	})
	req, err := LoginRequest(auth.Config.URL, auth.Config.ConjurVersion, csrBytes)
	if err != nil {
		return err
	}

	InfoLogger.Printf("Sending login request")
	resp, err := auth.client.Do(req)
	if err != nil {
		return err
	}

	InfoLogger.Printf("Checking for empty login response")
	err = EmptyResponse(resp)
	if err != nil {
		return err
	}

	// load client cert
	InfoLogger.Printf("Loading client certificate")
	certPEMBlock, err := ioutil.ReadFile(auth.Config.ClientCertPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("client certificate not found at %s", auth.Config.ClientCertPath)
		}

		return err
	}

	InfoLogger.Printf("Parsing client certificate")
	certDERBlock, certPEMBlock := pem.Decode(certPEMBlock)
	cert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return err
	}

	auth.PublicCert = cert

	InfoLogger.Printf("Removing client certificate")
	// clean up the client cert so it's only available in memory
	os.Remove(auth.Config.ClientCertPath)

	return nil
}

// Returns true if we are logged in (have a cert)
func (auth *Authenticator) IsLoggedIn() bool {
	return auth.PublicCert != nil
}

// Returns true if certificate is expired or close to expiring
func (auth *Authenticator) IsCertExpired() bool {
	certExpiresOn := auth.PublicCert.NotAfter.UTC()
	currentDate := time.Now().UTC()

	InfoLogger.Printf("Cert expires: %v", certExpiresOn)
	InfoLogger.Printf("Current date: %v", currentDate)
	InfoLogger.Printf("Buffer time:  %v", bufferTime)

	return currentDate.Add(bufferTime).After(certExpiresOn)
}

// Authenticate sends Conjur an authenticate request and returns
// the response data. Also manages state of certificates.
func (auth *Authenticator) Authenticate() ([]byte, error) {
	if !auth.IsLoggedIn() {
		InfoLogger.Printf("Not logged in. Trying to log in...")

		if err := auth.Login(); err != nil {
			ErrorLogger.Printf("Login failed: %v", err.Error())
			return nil, err
		}

		InfoLogger.Printf("Logged in")
	}

	if auth.IsCertExpired() {
		InfoLogger.Printf("Certificate expired. Re-logging in...")

		if err := auth.Login(); err != nil {
			return nil, err
		}

		InfoLogger.Printf("Logged in. Continuing authentication.")
	}

	privDer := x509.MarshalPKCS1PrivateKey(auth.privateKey)
	keyPEMBlock := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDer})

	certPEMBlock := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: auth.PublicCert.Raw})

	client, err := newHTTPSClient(auth.Config.SSLCertificate, certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, err
	}

	req, err := AuthenticateRequest(
		auth.Config.URL,
		auth.Config.ConjurVersion,
		auth.Config.Account,
		auth.Config.Username,
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
		content, err = decodeFromPEM(response, auth.PublicCert, auth.privateKey)
		if err != nil {
			return err
		}
	} else if auth.Config.ConjurVersion == "5" {
		content = response
	}

	// InfoLogger.Printf("Writing token %v to shared volume ...", content)
	err = ioutil.WriteFile(auth.Config.TokenFilePath, content, 0644)
	if err != nil {
		return err
	}

	InfoLogger.Printf("Successfully authenticated!")

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
