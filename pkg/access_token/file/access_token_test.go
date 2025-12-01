package file

import (
	"os"
	"path/filepath"
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
	tempDir := t.TempDir()
	var accessToken, _ = NewAccessToken(filepath.Join(tempDir, "accessTokenFile1"))

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
		accessToken.FilePath = filepath.Join(tempDir, "accessTokenFileWrite1")

		err := accessToken.Write(dataActual)
		assert.NoError(t, err)

		_, err = os.Stat(accessToken.FilePath)
		assert.NoError(t, err)

		dataExpected, _ := accessToken.Read()

		// Confirm data was written
		eq := reflect.DeepEqual(dataActual, dataExpected)
		assert.True(t, eq)

		dataActual = []byte{'t', 'e', 's', 't', '2'}
		err = accessToken.Write(dataActual)
		assert.NoError(t, err)

		_, err = os.Stat(accessToken.FilePath)
		assert.NoError(t, err)

		// TODO: read the content with `os` methods (not with `accessToken`)
		dataExpected, _ = accessToken.Read()
		eq = reflect.DeepEqual(dataActual, dataExpected)
		assert.True(t, eq)

		t.Run("Creates nonexistent directory", func(t *testing.T) {
			accessToken.FilePath = filepath.Join(tempDir, "nonexistent_dir", "accessTokenFileWrite2")
			dataActual := []byte{'d', 'i', 'r', 'e', 'c', 't', 'o', 'r', 'y'}

			err := accessToken.Write(dataActual)
			assert.NoError(t, err)

			_, err = os.Stat(accessToken.FilePath)
			assert.NoError(t, err)

			dataExpected, _ := accessToken.Read()

			// Confirm data was written
			eq := reflect.DeepEqual(dataActual, dataExpected)
			assert.True(t, eq)
		})

		t.Run("Given an access token without data", func(t *testing.T) {
			err := accessToken.Write(nil)

			assert.EqualError(t, err, log.CAKC005)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		dataActual := []byte{'t', 'e', 's', 't'}
		accessToken.FilePath = filepath.Join(tempDir, "accessTokenFileDel1")
		err := accessToken.Write(dataActual)
		assert.NoError(t, err)

		// Read is added here because we want to check later that the contents were deleted from memory successfully
		dataFromRead, err := accessToken.Read()
		assert.NoError(t, err)

		err = accessToken.Delete()
		assert.NoError(t, err)

		_, err = os.Stat(filepath.Join(tempDir, "accessTokenFileDel1"))
		assert.Error(t, err)
		// Both input & output arrays should be cleared from memory
		empty := make([]byte, len(dataActual))
		eq := reflect.DeepEqual(dataActual, empty)
		assert.True(t, eq)
		eq = reflect.DeepEqual(dataFromRead, empty)
		assert.True(t, eq)

		t.Run("Given an access token with no data saved to a file", func(t *testing.T) {
			accessToken.Data = nil
			delFilePath := filepath.Join(tempDir, "accessTokenFileDel2")
			os.Create(delFilePath)
			accessToken.FilePath = delFilePath

			err := accessToken.Delete()
			assert.NoError(t, err)

			// Check that file does not exist
			_, err = os.Stat(delFilePath)
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
