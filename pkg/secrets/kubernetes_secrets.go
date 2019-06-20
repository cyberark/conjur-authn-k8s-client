package secrets

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func storeSecretsInKubeSecrets(namespace string, secretName string, secrets []Secret) error {

	// Create a map of secret key to value to place in the
	// Kubernetes secret
	secretsData := make(map[string][]byte)
	for _, secret := range secrets {
		secretsData[secret.Key] = secret.SecretBytes
	}

	// Create the Kubernetes client
	InfoLogger.Print("Creating Kubernetes client...")
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("Failed to load in-cluster Kubernetes client config: %s", err)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return fmt.Errorf("Failed to create Kubernetes client: %s", err)
	}

	// Create the Kubernetes secret
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: secretsData,
	}

	// Store the secret in Kubernetes
	InfoLogger.Printf("Creating the Kubernetes secret '%s' in namespace '%s'...", secretName, namespace)
	_, err = kubeClient.CoreV1().Secrets(namespace).Create(&secret)
	if err != nil {
		return fmt.Errorf("Failed to write Kubernetes secret: %s", err)
	}

	return nil
}
