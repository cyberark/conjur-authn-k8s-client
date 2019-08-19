package storage

import "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"

//type AccessTokenHandler interface {
//	Read()	([]byte, error)
//	Write()	error
//	Delete()	error
//}

type AccessTokenMemory struct {
	Data	[]byte
}

// New returns a new AccessTokenMemory object
// return interface
func NewAccessTokenMemory(config config.Config) (token *AccessTokenMemory, err error) {
	return &AccessTokenMemory {

	}, nil
}

func (token *AccessTokenMemory) Read() (Data []byte, err error) {
	return nil,nil
}

func (token *AccessTokenMemory) Write() (err error) {
	return nil
}

func (token *AccessTokenMemory) Delete() (err error) {
	return nil
}
