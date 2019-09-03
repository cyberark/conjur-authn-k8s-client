package conjur

import (
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
)

// We have this interface so we can mock it in UTs
type ConjurSecretsFetcherInterface interface {
	// functions will need to be implemented by another struct
	RetrieveConjurSecrets(accessToken []byte, variableIDs []string) (map[string][]byte, error)
}

// We create this empty struct so we have an object with the functions below
type ConjurSecretsFetcher struct{}

func (conjurSecretsFetcher ConjurSecretsFetcher) RetrieveConjurSecrets(accessToken []byte, variableIDs []string) (map[string][]byte, error) {
	var (
		conjurClient ConjurClient
		err          error
	)

	log.InfoLogger.Println(log.CAKC018I, variableIDs)

	conjurClient, err = NewConjurClient(accessToken)
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC020E)
	}

	retrievedSecrets, err := conjurClient.RetrieveBatchSecrets(variableIDs)
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC021E, err.Error())
	}

	return retrievedSecrets, nil
}
