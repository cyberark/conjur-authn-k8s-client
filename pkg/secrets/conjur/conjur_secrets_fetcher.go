package conjur

import (
	"fmt"
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

	provider, err = conjurProvider(accessToken)
	if err != nil {
		return nil, fmt.Errorf("error create Conjur secrets provider: %s", err)
	}

	retrievedSecrets, err := provider.RetrieveBatchSecrets(variableIDs)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Conjur secrets: %s", err)
	}

	return retrievedSecrets, nil
}
