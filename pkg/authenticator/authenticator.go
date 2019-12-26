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
	"time"

	"github.com/fullsailor/pkcs7"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token/file"
	authnConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

var oidExtensionSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
var bufferTime = 30 * time.Second

// Authenticator contains the configuration and client
// for the authentication connection to Conjur
type Authenticator struct {
	client      *http.Client
	privateKey  *rsa.PrivateKey
	AccessToken access_token.AccessToken
	Config      authnConfig.Config
	PublicCert  *x509.Certificate
}

const (
	nameTypeEmail = 1
	nameTypeDNS   = 2
	nameTypeURI   = 6
	nameTypeIP    = 7
)

func New(config authnConfig.Config) (*Authenticator, error) {
	accessToken, err := file.NewAccessToken(config.TokenFilePath)
	if err != nil {
		return nil, log.RecordedError(log.CAKC001E)
	}

	return NewWithAccessToken(config, accessToken)
}

func NewWithAccessToken(config authnConfig.Config, accessToken access_token.AccessToken) (*Authenticator, error) {
	signingKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, log.RecordedError(log.CAKC030E, err.Error())
	}

	client, err := newHTTPSClient(config.SSLCertificate, nil, nil)
	if err != nil {
		return nil, err
	}

	return &Authenticator{
		client:      client,
		privateKey:  signingKey,
		AccessToken: accessToken,
		Config:      config,
	}, nil
}

// GenerateCSR prepares the CSR
func (auth *Authenticator) GenerateCSR(commonName string) ([]byte, error) {
	sanURIString, err := generateSANURI(auth.Config.PodNamespace, auth.Config.PodName)
	sanURI, err := url.Parse(sanURIString)
	if err != nil {
		return nil, err
	}

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

	log.InfoLogger.Printf(log.CAKC007I, auth.Config.Username)

	csrRawBytes, err := auth.GenerateCSR(auth.Config.Username.Suffix)

	csrBytes := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrRawBytes,
	})

	req, err := LoginRequest(auth.Config.URL, auth.Config.ConjurVersion, csrBytes, auth.Config.Username.Prefix)
	if err != nil {
		return err
	}

	resp, err := auth.client.Do(req)
	if err != nil {
		return log.RecordedError(log.CAKC028E, err.Error())
	}

	err = EmptyResponse(resp)
	if err != nil {
		return log.RecordedError(log.CAKC029E, err.Error())
	}

	// load client cert
	certPEMBlock, err := ioutil.ReadFile(auth.Config.ClientCertPath)
	if err != nil {
		if os.IsNotExist(err) {
			return log.RecordedError(log.CAKC011E, auth.Config.ClientCertPath)
		}

		return log.RecordedError(log.CAKC012E, err.Error())
	}

	certDERBlock, certPEMBlock := pem.Decode(certPEMBlock)
	cert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return log.RecordedError(log.CAKC013E, auth.Config.ClientCertPath, err.Error())
	}

	auth.PublicCert = cert

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

	log.InfoLogger.Printf(log.CAKC008I, certExpiresOn)
	log.InfoLogger.Printf(log.CAKC009I, currentDate)
	log.InfoLogger.Printf(log.CAKC010I, bufferTime)

	return currentDate.Add(bufferTime).After(certExpiresOn)
}

// Authenticate sends Conjur an authenticate request and returns
// the response data. Also manages state of certificates.
func (auth *Authenticator) Authenticate() ([]byte, error) {
	if !auth.IsLoggedIn() {
		log.InfoLogger.Printf(log.CAKC005I)

		if err := auth.Login(); err != nil {
			return nil, log.RecordedError(log.CAKC015E)
		}

		log.InfoLogger.Printf(log.CAKC002I)
	}

	if auth.IsCertExpired() {
		log.InfoLogger.Printf(log.CAKC004I)

		if err := auth.Login(); err != nil {
			return nil, err
		}

		log.InfoLogger.Printf(log.CAKC003I)
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
		auth.Config.Username.FullUsername,
	)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, log.RecordedError(log.CAKC027E, err.Error())
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

	err = auth.AccessToken.Write(content)
	if err != nil {
		return err
	}

	log.InfoLogger.Printf(log.CAKC001I)

	return nil
}

// generateSANURI returns the formatted uri(SPIFFEE format for now) for the certificate.
func generateSANURI(namespace, podname string) (string, error) {
	if namespace == "" || podname == "" {
		return "", log.RecordedError(log.CAKC008E, namespace, podname)
	}
	return fmt.Sprintf("spiffe://cluster.local/namespace/%s/podname/%s", namespace, podname), nil
}

func marshalSANs(dnsNames, emailAddresses []string, ipAddresses []net.IP, uris []*url.URL) ([]byte, error) {
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
		return nil, log.RecordedError(log.CAKC026E, err.Error())
	}

	decodedPEM, err = p7.Decrypt(publicCert, privateKey)
	if err != nil {
		return nil, log.RecordedError(log.CAKC025E, err.Error())
	}

	return decodedPEM, nil
}
