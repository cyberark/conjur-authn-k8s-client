package conjur

import (
	"fmt"
	"github.com/cyberark/conjur-api-go/conjurapi"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/sidecar/logging"
)

var InfoLogger = log.InfoLogger

func conjurProvider(tokenData []byte) (ConjurProvider, error) {
	InfoLogger.Printf("Creating Conjur client...")
	config, err := conjurapi.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load Conjur config: %s", err)
	}

	client, err := conjurapi.NewClientFromToken(config, string(tokenData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Conjur client from token: %s", err)
	}

	return client, nil
}
