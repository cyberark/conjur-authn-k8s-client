package access_token

type AccessTokenHandler interface {
	Delete() error
	Read() ([]byte, error)
	Write(Data []byte) error
}
