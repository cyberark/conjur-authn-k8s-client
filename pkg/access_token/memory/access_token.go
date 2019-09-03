package memory

import (
	"github.com/cyberark/conjur-authn-k8s-client/pkg/logger"
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
		return nil, logger.PrintAndReturnError(logger.CAKC006E)
	}

	return token.Data, nil
}

func (token *AccessToken) Write(Data []byte) (err error) {
	if Data == nil {
		return logger.PrintAndReturnError(logger.CAKC005E)
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
