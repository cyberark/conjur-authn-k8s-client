package memory

import (
	"fmt"
)

type AccessToken struct {
	Data []byte
}

func NewAccessToken() (token *AccessToken, err error) {
	return &AccessToken{
		Data: nil,
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
	return nil
}

func (token *AccessToken) Delete() (err error) {
	// Clear Data
	empty := make([]byte, len(token.Data))
	copy(token.Data, empty)
	token.Data = nil

	return nil
}
