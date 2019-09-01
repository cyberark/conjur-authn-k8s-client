package access_token

/*
	This interface handles a Conjur access token. Structs implementing this interface should have the ability to read,
	write & delete the access tokenץ

	For example, in the conjur-k8s-secrets-manager we will use AccessTokenMemory. we will create the authenticator
	object with this handler which will not write the data to a file. later on, we will use the Read() method to get the
	token for retrieving secrets from conjur.
*/
type AccessTokenHandler interface {
	Read() ([]byte, error)
	Write(Data []byte) error
	Delete() error
}
