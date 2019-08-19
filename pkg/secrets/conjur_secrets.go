package secrets

import (
	"fmt"
)

type ConjurProvider interface {
	RetrieveSecret(string) ([]byte, error)
	RetrieveBatchSecrets([]string) (map[string][]byte, error)
}

func GetVariableIDsToRetrieve(pathMap map[string]string) ([]string, error) {
	var variableIDs []string

	if len(pathMap) == 0 {
		return nil, fmt.Errorf("Error map should not be empty")
	}

	for key, _ := range pathMap {
		variableIDs = append(variableIDs, key)
	}

	return variableIDs, nil
}

func RetrieveConjurSecrets(accessToken []byte, variableIDs []string) (map[string][]byte, error) {
	var (
		provider ConjurProvider
		err      error
	)

	provider, err = conjurProvider(accessToken)
	if err != nil {
		return nil, fmt.Errorf("Error create Conjur secrets provider: %s", err)
	}

	retrievedSecrets, err := provider.RetrieveBatchSecrets(variableIDs)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Conjur secrets: %s", err)
	}

	return retrievedSecrets, nil
}
