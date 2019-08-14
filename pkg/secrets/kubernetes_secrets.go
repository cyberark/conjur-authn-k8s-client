package secrets

import (
	"fmt"
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

	// TODO: improve this code
	data := []byte("{\"stringData\":{")
	for key, value := range stringDataEntriesMap {
		entry := []byte(fmt.Sprintf("\"%s\": \"", key))
		entry = append(append(entry, value...), []byte("\"")...)
		entry = append(entry, []byte(",")...)
		data = append(data, entry...)
	}
	// TODO: remove last comma
	data = append(data, []byte("}}")...)

	InfoLogger.Printf("Patching Kubernetes secret '%s' in namespace '%s'...", secretName, namespace)
	_, err = kubeClient.CoreV1().Secrets(namespace).Patch(secretName, types.StrategicMergePatchType, data)
	if err != nil {
		return fmt.Errorf("failed to patch Kubernetes secret: %s", err)
	}

	return nil
}
