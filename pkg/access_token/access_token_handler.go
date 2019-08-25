package access_token

type AccessTokenHandler interface {
	Read() ([]byte, error)
	Write() error
	Delete() error
}
