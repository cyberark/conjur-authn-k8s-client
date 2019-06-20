package secrets

import (
	"encoding/asn1"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	secretsConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/config"
	"github.com/cyberark/summon/secretsyml"
)

var oidExtensionSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
var bufferTime = 30 * time.Second

// Secrets contains the configuration and client
// for the authentication connection to Conjur
type Secrets struct {
	Config      secretsConfig.Config
	AccessToken []byte
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
	return &Secrets{
		Config: config,
	}, nil
}

// LoadSecrets reads a provided 'secrets.yml' and attempts
// to retrieve the secrets from Conjur.
func (secrets *Secrets) LoadSecrets() (*SecretResponse, error) {

	InfoLogger.Printf("Loading secrets...")

	secretsYamlPath := secrets.Config.SecretsYamlPath
	secretsMap, err := readSecretsYaml(secretsYamlPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading '%s': %s", secretsYamlPath, err)
	}

	// Read the Conjur access token created by the authenticator
	accessToken, err := ioutil.ReadFile(secrets.Config.TokenFilePath)
	if err != nil {
		return nil, fmt.Errorf("Error reading access token: %s", err)
	}

	secretValues, err := fetchSecretValues(accessToken, secretsMap)
	if err != nil {
		return nil, fmt.Errorf("Error fetching secret values: %s", err)
	}

	return &SecretResponse{
		Secrets: secretValues,
	}, nil
}

func (secrets *Secrets) HandleSecretsResponse(response *SecretResponse) error {
	InfoLogger.Printf("Writing secrets to destinations...")

	secretsDirPath := secrets.Config.SecretsDirPath
	err := storeSecretsInVolume(secretsDirPath, response.Secrets)
	if err != nil {
		return fmt.Errorf("Error writing secrets to volume: %s", err)
	}

	namespace := secrets.Config.PodNamespace
	secretName := secrets.Config.KubeSecretName
	err = storeSecretsInKubeSecrets(namespace, secretName, response.Secrets)
	if err != nil {
		return fmt.Errorf("Error writing secrets to K8s secrets: %s", err)
	}

	return nil
}

type Provider interface {
	RetrieveSecret(string) ([]byte, error)
}

func readSecretsYaml(path string) (secretsyml.SecretsMap, error) {
	InfoLogger.Printf("Reading secrets.yml...")
	secretsMap, err := secretsyml.ParseFromFile(path, "", nil)
	if err != nil {
		return nil, fmt.Errorf("Error reading '%s': %s", path, err)
	}

	return secretsMap, nil
}

func fetchSecretValues(accessToken []byte, secretsMap secretsyml.SecretsMap) ([]Secret, error) {
	var (
		provider Provider
		err      error
	)

	InfoLogger.Printf("Fetching secret values...")

	tempFactory := NewTempFactory("")

	type Result struct {
		key   string
		bytes []byte
		error
	}

	// Run provider calls concurrently
	results := make(chan Result, len(secretsMap))
	var wg sync.WaitGroup

	// Lazy loading provider
	for _, spec := range secretsMap {
		if provider == nil && spec.IsVar() {
			provider, err = conjurProvider(accessToken)
			if err != nil {
				return nil, fmt.Errorf("Error create Conjur secrets provider: %s", err)
			}
		}
	}

	for key, spec := range secretsMap {
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

	secretResults := make([]Secret, len(secretsMap))

	// Put secrets in a data structure
	idx := 0
	for result := range results {
		if result.error == nil {
			secretResults[idx] = Secret{
				Key:         result.key,
				SecretBytes: result.bytes,
			}
		} else {
			return nil, fmt.Errorf("Error fetching secret '%s': %s", result.key, result.error)
		}
		idx++
	}

	return secretResults, nil
}
