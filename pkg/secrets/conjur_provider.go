package secrets

import (
	"fmt"

	"github.com/cyberark/conjur-api-go/conjurapi"
)

func conjurProvider(tokenData []byte) (ConjurProvider, error) {
	InfoLogger.Printf("Creating Conjur client...")
	config, err := conjurapi.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("Failed to load Conjur config: %s", err)
	}

	client, err := conjurapi.NewClientFromToken(config, string(tokenData))
	if err != nil {
		return nil, fmt.Errorf("Failed to create Conjur client from token: %s", err)
	}

	return client, nil
}
