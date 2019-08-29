package handlers

import (
	"fmt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	secretsConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/config"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/conjur"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/k8s"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/sidecar/logging"
	"strings"
)

type SecretsHandlerK8sUseCase struct {
	AccessTokenHandler   access_token.AccessTokenHandler
	ConjurSecretsFetcher conjur.ConjurSecretsFetcherInterface
	K8sSecretsHandler    k8s.K8sSecretsHandler
}

var errLogger = log.ErrorLogger

func NewSecretHandlerK8sUseCase(secretsConfig secretsConfig.Config, AccessTokenHandler access_token.AccessTokenHandler) (SecretsHandler *SecretsHandlerK8sUseCase, err error) {
	k8sSecretsHandler, err := k8s.New(secretsConfig)
	if err != nil {
		errLogger.Printf("Failed to create k8s secrets handler: %s", err.Error())
		return nil, err
	}

	var conjurSecretsFetcher conjur.ConjurSecretsFetcher

	return &SecretsHandlerK8sUseCase{
		AccessTokenHandler:   AccessTokenHandler,
		ConjurSecretsFetcher: conjurSecretsFetcher,
		K8sSecretsHandler:    *k8sSecretsHandler,
	}, nil
}

func (secretsHandlerK8sUseCase SecretsHandlerK8sUseCase) HandleSecrets() error {
	k8sSecretsMap, err := secretsHandlerK8sUseCase.K8sSecretsHandler.RetrieveK8sSecrets()
	if err != nil {
		errLogger.Printf("Failure retrieving k8s secretsHandlerK8sUseCase: %s", err.Error())
		return err
	}

	accessToken, err := secretsHandlerK8sUseCase.AccessTokenHandler.Read()
	if err != nil {
		errLogger.Printf("Failure retrieving access token: %s", err.Error())
		return err
	}

	variableIDs, err := getVariableIDsToRetrieve(k8sSecretsMap.PathMap)
	if err != nil {
		return fmt.Errorf("error parsing Conjur variable ids: %s", err)
	}

	retrievedConjurSecrets, err := secretsHandlerK8sUseCase.ConjurSecretsFetcher.RetrieveConjurSecrets(accessToken, variableIDs)
	if err != nil {
		return fmt.Errorf("error retrieving Conjur k8sSecretsHandler: %s", err)
	}

	err = updateK8sSecretsMapWithConjurSecrets(k8sSecretsMap, retrievedConjurSecrets)
	if err != nil {
		errLogger.Printf("Failure updating K8s K8sSecretsHandler map: %s", err.Error())
		return err
	}

	err = secretsHandlerK8sUseCase.K8sSecretsHandler.PatchK8sSecrets(k8sSecretsMap)
	if err != nil {
		errLogger.Printf("Failure patching K8s K8sSecretsHandler: %s", err.Error())
		return err
	}

	return nil
}

func getVariableIDsToRetrieve(pathMap map[string][]string) ([]string, error) {
	var variableIDs []string

	if len(pathMap) == 0 {
		return nil, fmt.Errorf("error map should not be empty")
	}

	for key, _ := range pathMap {
		variableIDs = append(variableIDs, key)
	}

	return variableIDs, nil
}

func updateK8sSecretsMapWithConjurSecrets(k8sSecretsMap *k8s.K8sSecretsMap, conjurSecrets map[string][]byte) error {
	var err error

	// Update K8s map by replacing variable IDs with their corresponding secret values
	for variableId, secret := range conjurSecrets {
		variableId, err = parseVariableID(variableId)
		if err != nil {
			return fmt.Errorf("failed to update k8s k8sSecretsHandler map: %s", err)
		}

		for _, locationInK8sSecretsMap := range k8sSecretsMap.PathMap[variableId] {
			locationInK8sSecretsMap := strings.Split(locationInK8sSecretsMap, ":")
			k8sSecretName := locationInK8sSecretsMap[0]
			k8sSecretDataEntryKey := locationInK8sSecretsMap[1]
			k8sSecretsMap.K8sSecrets[k8sSecretName][k8sSecretDataEntryKey] = secret
		}

		// Clear secret from memory
		empty := make([]byte, len(secret))
		copy(secret, empty)
		secret = nil
	}

	return nil
}

// The variable ID is in the format "<account>:variable:<variable_id>. we need only the last part.
func parseVariableID(fullVariableId string) (string, error) {
	variableIdParts := strings.Split(fullVariableId, ":")
	if len(variableIdParts) != 3 {
		return "", fmt.Errorf("failed to parse Conjur variable ID: %s", fullVariableId)
	}

	return variableIdParts[2], nil
}
