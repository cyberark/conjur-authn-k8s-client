package storage

import "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"

type AccessTokenMemory struct {
	Data	[]byte
}

func NewAccessTokenMemory(config config.Config) (token *AccessTokenMemory, err error) {
	return &AccessTokenMemory {

	}, nil
}

func (token AccessTokenMemory) Read() (Data []byte, err error) {
	return nil,nil
}

func (token *AccessTokenMemory) Write() (err error) {
	return nil
}

func (token *AccessTokenMemory) Delete() (err error) {
	return nil
}
