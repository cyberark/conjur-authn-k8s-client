package file

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

type AccessToken struct {
	Data     []byte
	FilePath string
}

func NewAccessToken(filePath string) (*AccessToken, error) {
	return &AccessToken{
		Data:     nil,
		FilePath: filePath,
	}, nil
}

func (token AccessToken) Read() ([]byte, error) {
	if token.Data == nil {
		return nil, log.RecordedError(log.CAKC006E)
	}

	return token.Data, nil
}

func (token *AccessToken) Write(Data []byte) (err error) {
	if Data == nil {
		return log.RecordedError(log.CAKC005E)
	}

	token.Data = Data
	// Write the data to file
	// Create the directory if it doesn't exist
	tokenDir := filepath.Dir(token.FilePath)
	if _, err := os.Stat(tokenDir); os.IsNotExist(err) {
		err = os.MkdirAll(tokenDir, 755)
		if err != nil {
			// Do not specify the directory in the error message for security reasons
			return log.RecordedError(log.CAKC004E)
		}
	}

	err = ioutil.WriteFile(token.FilePath, token.Data, 0644)
	if err != nil {
		// Do not specify the file path in the error message for security reasons
		return log.RecordedError(log.CAKC003E)
	}

	return nil
}

func (token *AccessToken) Delete() (err error) {
	err = os.Remove(token.FilePath)
	if err != nil {
		// Do not specify the file path in the error message for security reasons
		return log.RecordedError(log.CAKC002E)
	}

	// Clear Data
	empty := make([]byte, len(token.Data))
	copy(token.Data, empty)
	token.Data = nil

	return nil
}
