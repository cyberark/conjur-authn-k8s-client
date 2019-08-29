package handlers

import (
	"fmt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	secretsConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/config"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/conjur"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/k8s"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
	"strings"
)

type SecretsHandlerK8sUseCase struct {
	AccessTokenHandler   access_token.AccessTokenHandler
	ConjurSecretsFetcher conjur.ConjurSecretsFetcherInterface
	K8sSecretsHandler    k8s.K8sSecretsHandler
}

func NewSecretHandlerK8sUseCase(secretsConfig secretsConfig.Config, AccessTokenHandler access_token.AccessTokenHandler) (SecretsHandler *SecretsHandlerK8sUseCase, err error) {
	k8sSecretsHandler, err := k8s.New(secretsConfig)
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC022E, err, false)
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
		return log.PrintAndReturnError(log.CAKC023E, err, false)
	}

	accessToken, err := secretsHandlerK8sUseCase.AccessTokenHandler.Read()
	if err != nil {
		return log.PrintAndReturnError(log.CAKC024E, err, false)
	}

	variableIDs, err := getVariableIDsToRetrieve(k8sSecretsMap.PathMap)
	if err != nil {
		return log.PrintAndReturnError(log.CAKC025E, err, false)
	}

	retrievedConjurSecrets, err := secretsHandlerK8sUseCase.ConjurSecretsFetcher.RetrieveConjurSecrets(accessToken, variableIDs)
	if err != nil {
		return log.PrintAndReturnError(log.CAKC026E, err, false)
	}

	k8sSecretsMap, err = updateK8sSecretsMapWithConjurSecrets(k8sSecretsMap, retrievedConjurSecrets)
	if err != nil {
		return log.PrintAndReturnError(log.CAKC027E, err, false)
	}

	err = secretsHandlerK8sUseCase.K8sSecretsHandler.PatchK8sSecrets(k8sSecretsMap)
	if err != nil {
		return log.PrintAndReturnError(log.CAKC028E, err, false)
	}

	return nil
}

func getVariableIDsToRetrieve(pathMap map[string]string) ([]string, error) {
	var variableIDs []string

	if len(pathMap) == 0 {
		return nil, log.PrintAndReturnError(log.CAKC029E, nil, false)
	}

	for key, _ := range pathMap {
		variableIDs = append(variableIDs, key)
	}

	return variableIDs, nil
}

func updateK8sSecretsMapWithConjurSecrets(k8sSecretsMap *k8s.K8sSecretsMap, conjurSecrets map[string][]byte) (*k8s.K8sSecretsMap, error) {
	var err error

	// Update K8s map by replacing variable IDs with their corresponding secret values
	for variableId, secret := range conjurSecrets {
		variableId, err = parseVariableID(variableId)
		if err != nil {
			return nil, log.PrintAndReturnError(log.CAKC030E, err, false)
		}

		locationInK8sSecretsMap := strings.Split(k8sSecretsMap.PathMap[variableId], ":")
		k8sSecretName := locationInK8sSecretsMap[0]
		k8sSecretDataEntryKey := locationInK8sSecretsMap[1]
		k8sSecretsMap.K8sSecrets[k8sSecretName][k8sSecretDataEntryKey] = secret

		// Clear secret from memory
		secret = nil
	}

	return k8sSecretsMap, nil
}

// The variable ID is in the format "<account>:variable:<variable_id>. we need only the last part.
func parseVariableID(fullVariableId string) (string, error) {
	variableIdParts := strings.Split(fullVariableId, ":")
	if len(variableIdParts) != 3 {
		return "", log.PrintAndReturnError(fmt.Sprintf(log.CAKC031E, fullVariableId), nil, false)
	}

	return variableIdParts[2], nil
}
