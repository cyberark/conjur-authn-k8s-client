package file

import (
	"os"
	"reflect"
	"testing"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/stretchr/testify/assert"
)

type ProxyHandlerTokenFile struct {
	AccessToken access_token.AccessToken
}

func TestAccessTokenFile(t *testing.T) {
	var accessToken, _ = NewAccessToken("/tmp/accessTokenFile1")

	t.Run("Read", func(t *testing.T) {
		dataActual := []byte{'t', 'e', 's', 't'}
		err := accessToken.Write(dataActual)
		assert.NoError(t, err)

		dataExpected, err := accessToken.Read()
		assert.NoError(t, err)

		eq := reflect.DeepEqual(dataActual, dataExpected)
		assert.True(t, eq)

		t.Run("Given an access token's data is empty", func(t *testing.T) {
			accessToken.Data = nil

			_, err := accessToken.Read()
			assert.EqualError(t, err, log.CAKC006)
		})
	})

	t.Run("Write", func(t *testing.T) {

		dataActual := []byte{'t', 'e', 's', 't'}
		accessToken.FilePath = "/tmp/accessTokenFileWrite1"

		err := accessToken.Write(dataActual)
		assert.NoError(t, err)

		_, err = os.Stat("/tmp/accessTokenFileWrite1")
		assert.NoError(t, err)

		dataExpected, _ := accessToken.Read()

		// Confirm data was written
		eq := reflect.DeepEqual(dataActual, dataExpected)
		assert.True(t, eq)

		dataActual = []byte{'t', 'e', 's', 't', '2'}
		err = accessToken.Write(dataActual)
		assert.NoError(t, err)

		_, err = os.Stat("/tmp/accessTokenFileWrite1")
		assert.NoError(t, err)

		// TODO: read the content with `os` methods (not with `accessToken`)
		dataExpected, _ = accessToken.Read()
		eq = reflect.DeepEqual(dataActual, dataExpected)
		assert.True(t, eq)

		t.Run("Given an access token without data", func(t *testing.T) {
			err := accessToken.Write(nil)

			assert.EqualError(t, err, log.CAKC005)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		dataActual := []byte{'t', 'e', 's', 't'}
		accessToken.FilePath = "/tmp/accessTokenFileDel1"
		err := accessToken.Write(dataActual)
		assert.NoError(t, err)

		// Read is added here because we want to check later that the contents were deleted from memory successfully
		dataFromRead, err := accessToken.Read()
		assert.NoError(t, err)

		err = accessToken.Delete()
		assert.NoError(t, err)

		_, err = os.Stat("/tmp/accessTokenFileDel1")
		assert.Error(t, err)
		// Both input & output arrays should be cleared from memory
		empty := make([]byte, len(dataActual))
		eq := reflect.DeepEqual(dataActual, empty)
		assert.True(t, eq)
		eq = reflect.DeepEqual(dataFromRead, empty)
		assert.True(t, eq)

		t.Run("Given an access token with no data saved to a file", func(t *testing.T) {
			accessToken.Data = nil
			os.Create("/tmp/accessTokenFileDel2")
			accessToken.FilePath = "/tmp/accessTokenFileDel2"

			err := accessToken.Delete()
			assert.NoError(t, err)

			// Check that file does not exist
			_, err = os.Stat("/tmp/accessTokenFileDel2")
			assert.Error(t, err)
			err = accessToken.Delete()

			assert.Contains(t, err.Error(), log.CAKC002)
		})

		t.Run("Given two instances of the accessToken interface", func(t *testing.T) {
			// Write Data to source interface
			dataActual := []byte{'t', 'e', 's', 't'}
			accessToken.Write(dataActual)

			t.Run("When setting token file location in proxy struct", func(t *testing.T) {
				// Set proxy struct with source interface
				var proxyStruct ProxyHandlerTokenFile
				proxyStruct.AccessToken = accessToken

				// Delete access token from proxy
				err := proxyStruct.AccessToken.Delete()

				assert.NoError(t, err)

				// Returns no data because data in source interface was cleared
				dataExpected, err := accessToken.Read()
				assert.Nil(t, dataExpected)
				assert.EqualError(t, err, log.CAKC006)
			})
		})
	})
}
