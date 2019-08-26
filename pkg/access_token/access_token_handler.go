package access_token

type AccessTokenHandler interface {
	Read() ([]byte, error)
	Write(Data []byte) error
	Delete() error
}
