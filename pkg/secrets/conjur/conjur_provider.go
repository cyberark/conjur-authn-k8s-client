package conjur

import (
	"fmt"
	"github.com/cyberark/conjur-api-go/conjurapi"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
)

func conjurProvider(tokenData []byte) (ConjurProvider, error) {
	log.InfoLogger.Printf(log.CAKC015I)
	config, err := conjurapi.LoadConfig()
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC018E, err.Error())
	}

	client, err := conjurapi.NewClientFromToken(config, string(tokenData))
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC019E, err.Error())
	}

	return client, nil
}
