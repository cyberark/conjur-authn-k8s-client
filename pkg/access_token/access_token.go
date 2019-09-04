package access_token

/*
	This interface represents a Conjur access token residing in some arbitrary storage medium (e.g. in-app memory or
	file system). Structs implementing this interface should have the ability to read, write & delete the access token contents.
	For example, cyberark-secrets-provider-for-k8s uses memory.AccessToken. The authenticator object is created with an empty
	access token which is populated with the Write method when access token contents become available. In this case the
	Write method writes the data to in-app memory. Later, cyberark-secrets-provider-for-k8s can call the Read() method to access
	the token-contents for retrieving secrets from Conjur.
*/
type AccessToken interface {
	Read() ([]byte, error)
	Write(Data []byte) error
	Delete() error
}
