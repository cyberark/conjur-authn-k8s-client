package conjur

import (
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
)

/*
	This struct holds a ConjurClient and retrieves Conjur secrets with it.

 	For example:

	// the c-tor needs a Conjur access token to initialize the client
	conjurSecretsRetriever, err := conjur.NewConjurSecretsRetriever(accessToken)
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC069E)
	}

	// this method receives a list of Conjur variables to retrieve
	retrievedConjurSecrets, err := conjurSecretsRetriever.RetrieveConjurSecrets(variableIDs)
	if err != nil {
		return log.PrintAndReturnError(log.CAKC026E)
	}
*/
type ConjurSecretsRetriever struct {
	conjurClient ConjurClient
}

func NewConjurSecretsRetriever(accessToken []byte) (*ConjurSecretsRetriever, error) {
	conjurClient, err := NewConjurClient(accessToken)
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC020E)
	}

	return &ConjurSecretsRetriever{
		conjurClient: conjurClient,
	}, nil
}

func (conjurSecretsRetriever ConjurSecretsRetriever) RetrieveConjurSecrets(variableIDs []string) (map[string][]byte, error) {
	log.InfoLogger.Println(log.CAKC018I, variableIDs)

	retrievedSecrets, err := conjurSecretsRetriever.conjurClient.RetrieveBatchSecrets(variableIDs)
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC021E, err.Error())
	}

	return retrievedSecrets, nil
}
