package conjur

import (
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
)

type ConjurProvider interface {
	RetrieveSecret(string) ([]byte, error)
	RetrieveBatchSecrets([]string) (map[string][]byte, error)
}

type ConjurSecretsFetcherInterface interface {
	// functions will need to be implemented by another struct
	RetrieveConjurSecrets(accessToken []byte, variableIDs []string) (map[string][]byte, error)
}

// We create this empty struct so we have an object with the functions below
type ConjurSecretsFetcher struct{}

func (conjurSecretsFetcher ConjurSecretsFetcher) RetrieveConjurSecrets(accessToken []byte, variableIDs []string) (map[string][]byte, error) {
	var (
		provider ConjurProvider
		err      error
	)

	log.InfoLogger.Println(log.CAKC018I, variableIDs)

	provider, err = conjurProvider(accessToken)
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC020E)
	}

	retrievedSecrets, err := provider.RetrieveBatchSecrets(variableIDs)
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC021E, err.Error())
	}

	return retrievedSecrets, nil
}
