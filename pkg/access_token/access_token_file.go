package access_token

import (
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
)

type AccessTokenFile struct {
	TokenFilePath string
	Data          []byte
}

func NewAccessTokenFile(config config.Config) (token *AccessTokenFile, err error) {
	return &AccessTokenFile{
		TokenFilePath: config.TokenFilePath,
		Data:          nil,
	}, nil
}

func (token AccessTokenFile) Read() (Data []byte, err error) {
	if token.Data == nil {
		return nil, log.PrintAndReturnError(log.CAKC010E)
	}

	return token.Data, nil
}

func (token *AccessTokenFile) Write(Data []byte) (err error) {
	if Data == nil {
		return log.PrintAndReturnError(log.CAKC009E)
	}

	token.Data = Data
	// Write the data to file
	// Create the directory if it doesn't exist
	tokenDir := filepath.Dir(token.TokenFilePath)
	if _, err := os.Stat(tokenDir); os.IsNotExist(err) {
		err = os.MkdirAll(tokenDir, 755)
		if err != nil {
			// Do not specify the directory in the error message for security reasons
			return log.PrintAndReturnError(log.CAKC008E, err.Error())
		}
	}

	err = ioutil.WriteFile(token.TokenFilePath, token.Data, 0644)
	if err != nil {
		// Do not specify the file path in the error message for security reasons
		return log.PrintAndReturnError(log.CAKC007E, err.Error())
	}

	return nil
}

func (token *AccessTokenFile) Delete() (err error) {
	err = os.Remove(token.TokenFilePath)
	if err != nil {
		// Do not specify the file path in the error message for security reasons
		return log.PrintAndReturnError(log.CAKC006E)
	}

	// Clear Data
	empty := make([]byte, len(token.Data))
	copy(token.Data, empty)
	token.Data = nil

	return nil
}
