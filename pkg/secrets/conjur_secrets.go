package secrets

import (
	"fmt"
)

type ConjurProvider interface {
	RetrieveSecret(string) ([]byte, error)
	RetrieveBatchSecrets([]string) (map[string][]byte, error)
}

func GetVariableIDsToRetrieve(pathMap map[string]string) []string {
	variableIDs := make([]string, 0, len(pathMap))
	for key := range pathMap {
		variableIDs = append(variableIDs, key)
	}

	return variableIDs
}

func RetrieveConjurSecrets(accessToken []byte, variableIDs []string) (map[string][]byte, error) {
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
