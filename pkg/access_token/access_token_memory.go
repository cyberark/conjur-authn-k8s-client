package access_token

import (
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
)

type AccessTokenMemory struct {
	Data []byte
}

func NewAccessTokenMemory() (token *AccessTokenMemory, err error) {
	return &AccessTokenMemory{
		Data: nil,
	}, nil
}

func (token AccessTokenMemory) Read() (Data []byte, err error) {
	if token.Data == nil {
		return nil, log.PrintAndReturnError(log.CAKC010E)
	}

	return token.Data, nil
}

func (token *AccessTokenMemory) Write(Data []byte) (err error) {
	if Data == nil {
		return log.PrintAndReturnError(log.CAKC009E)
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
