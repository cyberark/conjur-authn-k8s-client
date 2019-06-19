package secrets

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"net/http"
	"time"

	secretsConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/config"
	sidecar "github.com/cyberark/conjur-authn-k8s-client/pkg/sidecar"
)

var oidExtensionSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
var bufferTime = 30 * time.Second

// Secrets contains the configuration and client
// for the authentication connection to Conjur
type Secrets struct {
	Config     secretsConfig.Config
	privateKey *rsa.PrivateKey
	PublicCert *x509.Certificate
	client     *http.Client
}

type secretResponse struct {
	secrets []secret
}

type secret struct {
	key         string
	secretBytes []byte
}

const (
	nameTypeEmail = 1
	nameTypeDNS   = 2
	nameTypeURI   = 6
	nameTypeIP    = 7
)

// New returns a new Authenticator
func New(config secretsConfig.Config) (secrets *Secrets, err error) {
	signingKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}

	client, err := sidecar.NewHTTPSClient(config.SSLCertificate, nil, nil)
	if err != nil {
		return nil, err
	}

	return &Secrets{
		Config:     config,
		client:     client,
		privateKey: signingKey,
	}, nil
}

func (secrets *Secrets) FetchSecrets() ([]byte, error) {
	// Get access token created by authenticator

	// Pull secrets from Conjur
	// See: https://github.com/cyberark/cloudfoundry-conjur-buildpack/blob/master/conjur-env/main.go

	// Put secrets in a data structure

	return nil, nil
}

func (secrets *Secrets) HandleSecretsResponse(response []byte) error {
	// Write secrets to K8s secrets manager
	// See: https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration

	// Write secrets to volume
	return nil
}
