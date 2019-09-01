package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
		return nil, fmt.Errorf("error reading access token, reason: data is empty")
	}

	return token.Data, nil
}

func (token *AccessToken) Write(Data []byte) (err error) {
	if Data == nil {
		return fmt.Errorf("error writing access token, reason: data is empty")
	}

	token.Data = Data
	// Write the data to file
	// Create the directory if it doesn't exist
	tokenDir := filepath.Dir(token.FilePath)
	if _, err := os.Stat(tokenDir); os.IsNotExist(err) {
		err = os.MkdirAll(tokenDir, 755)
		if err != nil {
			// Do not specify the directory in the error message for security reasons
			return fmt.Errorf("error writing access token, reason: failed to create directory")
		}
	}

	err = ioutil.WriteFile(token.FilePath, token.Data, 0644)
	if err != nil {
		// Do not specify the file path in the error message for security reasons
		return fmt.Errorf("error writing access token, reason: failed to write file")
	}

	return nil
}

func (token *AccessToken) Delete() (err error) {
	err = os.Remove(token.FilePath)
	if err != nil {
		// Do not specify the file path in the error message for security reasons
		return fmt.Errorf("error deleting access token")
	}

	// Clear Data
	empty := make([]byte, len(token.Data))
	copy(token.Data, empty)
	token.Data = nil

	return nil
}
