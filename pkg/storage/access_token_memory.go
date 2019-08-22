package storage

import (
	"fmt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
)

type AccessTokenMemory struct {
	Data []byte
}

func NewAccessTokenMemory(config config.Config) (token *AccessTokenMemory, err error) {
	return &AccessTokenMemory{
		Data: nil,
	}, nil
}

func (token AccessTokenMemory) Read() (Data []byte, err error) {
	if token.Data == nil {
		return nil, fmt.Errorf("error reading access token, reason: data is empty")
	}

	return token.Data, nil
}

func (token *AccessTokenMemory) Write(Data []byte) (err error) {
	if Data == nil {
		return fmt.Errorf("error writing access token, reason: data is empty")
	}

	token.Data = Data
	return nil
}

func (token *AccessTokenMemory) Delete() (err error) {
	// Clear Data
	empty := make([]byte, len(token.Data))
	copy(token.Data, empty)
	token.Data = nil

	return nil
}
