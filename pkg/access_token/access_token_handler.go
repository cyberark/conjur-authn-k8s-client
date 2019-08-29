package access_token

/*
	This interface handles a Conjur access token. Classes implementing this interface should have the ability to read,
	write & delete the access tokenץ

	For example, AccessTokenFile implements the AccessTokenHandler interface. It receives a file path in its constructor
	and its methods implementations are:
	Write - creates a file in the specified path and writes a byte array into it
	Read - reads the data from the file into a byte array
	Delete - deletes the file and clears the data from memory
*/
type AccessTokenHandler interface {
	Read() ([]byte, error)
	Write(Data []byte) error
	Delete() error
}
