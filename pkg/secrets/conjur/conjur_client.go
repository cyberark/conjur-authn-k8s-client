package conjur

import (
	"github.com/cyberark/conjur-api-go/conjurapi"

	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
)

/*
	Client for communication with Conjur. In this project it is used only for
    batch secrets retrieval so we expose only this method of the client.

	The name ConjurClient also improves readability as Client can be ambiguous.
*/
type ConjurClient interface {
	RetrieveBatchSecrets([]string) (map[string][]byte, error)
}

func NewConjurClient(tokenData []byte) (ConjurClient, error) {
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
