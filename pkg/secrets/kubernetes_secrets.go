package secrets

import (
	"fmt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sSecret struct {
	Secret *v1.Secret
}

func retrieveK8sSecret(namespace string, secretName string) (*K8sSecret, error) {
	// TODO: extract this to another function
	// Create the Kubernetes client
	InfoLogger.Print("Creating Kubernetes client...")
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster Kubernetes client config: %s", err)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %s", err)
	}

	// TODO: extract until here

	InfoLogger.Printf("Retrieving Kubernetes secret '%s' from namespace '%s'...", secretName, namespace)
	k8sSecret, err := kubeClient.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Kubernetes secret: %s", err)
	}

	return &K8sSecret{
		Secret: k8sSecret,
	}, nil
}

func patchK8sSecret(namespace string, secretName string, stringDataEntriesMap map[string][]byte) error {
	// TODO: extract this to another function

	// Create the Kubernetes client
	InfoLogger.Print("Creating Kubernetes client...")
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to load in-cluster Kubernetes client config: %s", err)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %s", err)
	}

	// TODO: extract until here

	stringDataEntry, err := generateStringDataEntry(stringDataEntriesMap)
	if err != nil {
		return fmt.Errorf("failed to parse Kubernetes secret list: %s", err)
	}

	InfoLogger.Printf("Patching Kubernetes secret '%s' in namespace '%s'...", secretName, namespace)
	_, err = kubeClient.CoreV1().Secrets(namespace).Patch(secretName, types.StrategicMergePatchType, stringDataEntry)
	// Clear secret from memory
	stringDataEntry = nil
	if err != nil {
		return fmt.Errorf("failed to patch Kubernetes secret: %s", err)
	}

	return nil
}

// Convert the data entry map to a stringData entry for the PATCH request.
// for example, the map:
// {
//   "username": "theuser",
//   "password": "supersecret"
// }
// will be parsed to the stringData entry "{\"stringData\":{\"username\": \"theuser\", \"password\": \"supersecret\"}}"
//
// we need the values to always stay as byte arrays so we don't have Conjur secrets stored as string.
func generateStringDataEntry(stringDataEntriesMap map[string][]byte) ([]byte, error) {
	var entry []byte
	index := 0

	if len(stringDataEntriesMap) == 0 {
		return nil, fmt.Errorf("error map should not be empty")
	}

	entries := make([][]byte, len(stringDataEntriesMap))
	// Parse every key-value pair in the map to a "key:value" byte array
	for key, value := range stringDataEntriesMap {
		entry = utils.ByteSlicePrintf(
			`"%v":"%v"`,
			"%v",
			[]byte(key),
			value,
		)
		entries[index] = entry
		index++

		// Clear secret from memory
		value = nil
		entry = nil
	}

	// Add a comma between each pair of entries
	numEntries := len(entries)
	entriesCombined := entries[0]
	for i := range entries {
		if i < numEntries-1 {
			entriesCombined = utils.ByteSlicePrintf(
				`%v,%v`,
				"%v",
				entriesCombined,
				entries[i+1],
			)
		}

		// Clear secret from memory
		entries[i] = nil
	}

	// Wrap all the entries
	stringDataEntry := utils.ByteSlicePrintf(
		`{"stringData":{%v}}`,
		"%v",
		entriesCombined,
	)

	// Clear secret from memory
	entriesCombined = nil

	return stringDataEntry, nil
}
