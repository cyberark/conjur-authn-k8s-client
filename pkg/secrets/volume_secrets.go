package secrets

import (
	"io/ioutil"
	"os"
	"path"
)

func storeSecretsInVolume(secretsDirPath string, secrets []Secret) error {
	// Create the directory if it doesn't exist
	if _, err := os.Stat(secretsDirPath); os.IsNotExist(err) {
		InfoLogger.Printf("Creating secrets directory: %s", secretsDirPath)
		os.MkdirAll(secretsDirPath, 755)
	}

	for _, secret := range secrets {
		InfoLogger.Printf("Storing secret, %s, in volume '%s'", secret.Key, secretsDirPath)
		err := ioutil.WriteFile(path.Join(secretsDirPath, secret.Key), secret.SecretBytes, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
