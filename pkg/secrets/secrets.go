package secrets

import (
	base64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	secretsConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/config"
)

// Secrets contains the configuration and client
// for the authentication connection to Conjur
type Secrets struct {
	Config      secretsConfig.Config
	AccessToken []byte
}

// New returns a new Secrets
func New(config secretsConfig.Config) (secrets *Secrets, err error) {
	return &Secrets{
		Config: config,
	}, nil
}

type K8sSecretsMap struct {
	// Maps a k8s Secret name to a data-entry map that holds the new entries that will be added to the k8s secret.
	// The data-entry map's key represents an entry name and the value is a Conjur variable ID that holds the value
	// of the required k8s secret. After the secret is retrieved from Conjur we replace the variable ID with its
	// corresponding secret value.
	// The variable ID (which is replaced later with the secret) is held as a byte array so we don't hold the secret as
	// clear text string
	K8sSecrets map[string]map[string][]byte

	// Maps a conjur variable id to its place in the k8sSecretsMap. This object helps us to replace
	// the variable IDs with their corresponding secret value in the map
	PathMap map[string]string
}

func (secrets *Secrets) RetrieveK8sSecrets() (*K8sSecretsMap, error) {
	namespace := secrets.Config.PodNamespace
	requiredK8sSecrets := secrets.Config.RequiredK8sSecrets

	k8sSecrets := make(map[string]map[string][]byte)
	PathMap := make(map[string]string)

	for _, secretName := range requiredK8sSecrets {
		k8sSecret, err := retrieveK8sSecret(namespace, secretName)
		if err != nil {
			return nil, fmt.Errorf("error reading k8s secrets: %s", err)
		}

		// Parse its "conjur-map" data entry and store its values in the new-data-entries map
		// This map holds data entries that will be added to the k8s secret after we retrieve their values from Conjur
		newDataEntriesMap := make(map[string][]byte)
		for key, value := range k8sSecret.Secret.Data {
			if key == "conjur-map" {
				// The data value is Base-64 encoded. We decode it before parsing it.
				decodedMap := make([]byte, base64.StdEncoding.DecodedLen(len(value)))
				_, err := base64.StdEncoding.Decode(decodedMap, value)
				if err != nil {
					return nil, fmt.Errorf("error decoding conjur-map of secret %s: %s", secretName, err)
				}

				// Split the conjur-map to k8s secret keys. each value holds a Conjur variable ID
				conjurMapEntries := strings.Split(string(decodedMap), "\n")
				for _, entry := range conjurMapEntries {
					// Parse each secret key and store it in the map
					k8sSecretKeyVal := strings.Split(entry, ":")
					k8sSecretKey := k8sSecretKeyVal[0]
					conjurVariableId := k8sSecretKeyVal[1]
					newDataEntriesMap[k8sSecretKey] = []byte(conjurVariableId)

					// This map will help us later to swap the variable id with the secret value
					PathMap[conjurVariableId] = fmt.Sprintf("%s:%s", secretName, k8sSecretKey)
				}
			}
		}

		// We add the data-entries map to the k8sSecrets map only if the k8s secret has a "conjur-map" data entry
		if len(newDataEntriesMap) > 0 {
			k8sSecrets[secretName] = newDataEntriesMap
		}
	}

	return &K8sSecretsMap{
		K8sSecrets: k8sSecrets,
	}, nil
}

func (secrets *Secrets) UpdateK8sSecretsMapWithConjurSecrets(k8sSecretsMap *K8sSecretsMap) (*K8sSecretsMap, error) {
	var err error

	// Read the Conjur access token created by the authenticator
	accessToken, err := ioutil.ReadFile(secrets.Config.TokenFilePath)
	if err != nil {
		return nil, fmt.Errorf("Error reading access token: %s", err)
	}

	pathMap := k8sSecretsMap.PathMap
	variableIDs := GetVariableIDsToRetrieve(pathMap)
	retrievedSecrets, err := RetrieveConjurSecrets(accessToken, variableIDs)

	// Update K8s map by replacing variable IDs with their corresponding secret values
	for variableId, secret := range retrievedSecrets {
		locationInK8sSecretsMap := strings.Split(pathMap[variableId], ":")
		k8sSecretName := locationInK8sSecretsMap[0]
		k8sSecretDataEntryKey := locationInK8sSecretsMap[1]
		k8sSecretsMap.K8sSecrets[k8sSecretName][k8sSecretDataEntryKey] = secret

		// Clear secret from memory
		secret = nil
	}

	return k8sSecretsMap, nil
}

func (secrets *Secrets) PatchK8sSecrets(k8sSecretsMap *K8sSecretsMap) error {
	namespace := secrets.Config.PodNamespace

	for secretName, dataEntryMap := range k8sSecretsMap.K8sSecrets {
		err := patchK8sSecret(namespace, secretName, dataEntryMap)
		if err != nil {
			return fmt.Errorf("failed to patch k8s secret: %s", err)
		}
	}

	return nil
}
