package storage

import "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"

type AccessTokenFile struct {
	TokenFilePath	string
	Data	[]byte
}

func NewAccessTokenFile(config config.Config) (token *AccessTokenFile, err error) {
	return
}

func (token AccessTokenFile) Read() (Data []byte, err error) {
	return
}

func (token *AccessTokenFile) Write() (err error) {
	return
}

func (token *AccessTokenFile) Delete() (err error) {
	return
}

