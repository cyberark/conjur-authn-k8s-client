package secrets

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/cyberark/conjur-api-go/conjurapi"
	secretsConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/config"
	sidecar "github.com/cyberark/conjur-authn-k8s-client/pkg/sidecar"
	"github.com/cyberark/summon/secretsyml"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	//
	// Uncomment to load all auth plugins
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

type SecretResponse struct {
	Secrets []Secret
}

type Secret struct {
	Key         string
	SecretBytes []byte
}

const (
	nameTypeEmail = 1
	nameTypeDNS   = 2
	nameTypeURI   = 6
	nameTypeIP    = 7
)

// New returns a new Secrets
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

func (secrets *Secrets) FetchSecrets() (*SecretResponse, error) {
	var provider Provider

	InfoLogger.Printf("Fetching secret values...")
	// Get access token created by authenticator
	tokenData, err := ioutil.ReadFile("/run/conjur/access-token")
	if err != nil {
		return nil, err
	}
	InfoLogger.Printf("Read token data.")

	InfoLogger.Printf("Reading secrets.yml...")
	// Pull secrets from Conjur
	// See: https://github.com/cyberark/cloudfoundry-conjur-buildpack/blob/master/conjur-env/main.go
	secretsRequest, err := secretsyml.ParseFromFile("secrets.yml", "", nil)
	if err != nil {
		return nil, err
	}
	InfoLogger.Printf("secrets.yml read.")

	tempFactory := NewTempFactory("")

	type Result struct {
		key   string
		bytes []byte
		error
	}

	// Run provider calls concurrently
	results := make(chan Result, len(secretsRequest))
	var wg sync.WaitGroup

	// Lazy loading provider
	for _, spec := range secretsRequest {
		if provider == nil && spec.IsVar() {
			provider, err = NewProvider(tokenData)
			if err != nil {
				return nil, err
			}
		}
	}

	for key, spec := range secretsRequest {
		wg.Add(1)
		go func(key string, spec secretsyml.SecretSpec) {
			var (
				secretBytes []byte
				err         error
			)

			if spec.IsVar() {
				secretBytes, err = provider.RetrieveSecret(spec.Path)

				if spec.IsFile() {
					fname := tempFactory.Push(secretBytes)
					secretBytes = []byte(fname)
				}
			} else {
				// If the spec isn't a variable, use its value as-is
				secretBytes = []byte(spec.Path)
			}

			results <- Result{key, secretBytes, err}
			wg.Done()
			return
		}(key, spec)
	}
	wg.Wait()
	close(results)

	secretResults := make([]Secret, len(secretsRequest))

	// Put secrets in a data structure
	idx := 0
	for result := range results {
		if result.error == nil {
			InfoLogger.Printf("Transforming secret %s", result.key)
			secretResults[idx] = Secret{
				Key:         result.key,
				SecretBytes: result.bytes,
			}
		} else {
			return nil, fmt.Errorf("Error fetching secret '%s' - %s", result.key, result.error)
		}
		idx++
	}

	return &SecretResponse{
		Secrets: secretResults,
	}, nil
}

func (secrets *Secrets) HandleSecretsResponse(response *SecretResponse) error {
	InfoLogger.Printf("Writing secrets to destinations...")

	err := storeSecretsInVolume(&secrets.Config, response.Secrets)
	if err != nil {
		return fmt.Errorf("Error writing secrets to volume: %s", err)
	}

	err = storeSecretsInKubeSecrets(&secrets.Config, response.Secrets)
	if err != nil {
		return fmt.Errorf("Error writing secrets to K8s secrets: %s", err)
	}

	return nil
}

func NewProvider(tokenData []byte) (Provider, error) {
	InfoLogger.Printf("Creating Conjur client...")
	config, err := conjurapi.LoadConfig()
	if err != nil {
		return nil, err
	}

	client, err := conjurapi.NewClientFromToken(config, string(tokenData))
	if err != nil {
		return nil, err
	}
	InfoLogger.Printf("Conjur client created.")

	return client, nil
}

type Provider interface {
	RetrieveSecret(string) ([]byte, error)
}

func storeSecretsInVolume(config *secretsConfig.Config, secrets []Secret) error {
	secretsDir := "/run/conjur/secrets"

	// Create the directory if it doesn't exist
	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		InfoLogger.Printf("Creating secrets directory: %s", secretsDir)
		os.MkdirAll(secretsDir, 755)
	}

	for _, secret := range secrets {
		InfoLogger.Printf("Storing secret, %s, in volume '%s'", secret.Key, secretsDir)
		err := ioutil.WriteFile(path.Join(secretsDir, secret.Key), secret.SecretBytes, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func storeSecretsInKubeSecrets(config *secretsConfig.Config, secrets []Secret) error {
	secretName := "dap-secrets"

	// Write secrets to K8s secrets manager
	// See: https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration
	// creates the in-cluster config
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	secretsData := make(map[string][]byte)
	for _, secret := range secrets {
		secretsData[secret.Key] = secret.SecretBytes
	}

	// creates the clientset
	InfoLogger.Print("Creating Kubernetes client")
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	InfoLogger.Printf("Creating the Kubernetes secret '%s'", secretName)
	_, err = kubeClient.CoreV1().Secrets(config.PodNamespace).Create(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: secretsData,
	})

	if err != nil {
		return err
	}

	InfoLogger.Printf("Secret created")

	return nil
}
