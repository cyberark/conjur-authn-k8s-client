package memory

import (
	"reflect"
	"testing"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/stretchr/testify/assert"
)

type ProxyHandlerTokenMemory struct {
	AccessToken access_token.AccessToken
}

func TestAccessTokenMemory(t *testing.T) {
	var accessToken, _ = NewAccessToken()

	t.Run("Read", func(t *testing.T) {

		t.Run("Given an access token with data", func(t *testing.T) {
			dataActual := []byte{'t', 'e', 's', 't'}
			err := accessToken.Write(dataActual)

			t.Run("Finishes without raising an error", func(t *testing.T) {
				assert.NoError(t, err)
			})

			t.Run("When running Read method", func(t *testing.T) {
				dataExpected, err := accessToken.Read()

				t.Run("Finishes without raising an error", func(t *testing.T) {
					assert.NoError(t, err)
				})

				t.Run("Returns the data that was written", func(t *testing.T) {
					eq := reflect.DeepEqual(dataActual, dataExpected)
					assert.True(t, eq)
				})
			})
		})

		t.Run("Given an access token's data is empty", func(t *testing.T) {
			accessToken.Data = nil

			t.Run("When running the Read method", func(t *testing.T) {
				_, err := accessToken.Read()

				t.Run("Raises an error that the data is empty", func(t *testing.T) {
					assert.EqualError(t, err, log.CAKC006)
				})
			})
		})
	})

	t.Run("Write", func(t *testing.T) {

		t.Run("Given an access token with data", func(t *testing.T) {
			dataActual := []byte{'t', 'e', 's', 't'}

			t.Run("Writes the access token to memory without raising an error", func(t *testing.T) {
				err := accessToken.Write(dataActual)
				assert.NoError(t, err)
			})

			t.Run("When running Read method", func(t *testing.T) {
				dataExpected, _ := accessToken.Read()

				// Confirm data was written
				t.Run("Returns the data the was written", func(t *testing.T) {
					eq := reflect.DeepEqual(dataActual, dataExpected)
					assert.True(t, eq)
				})
			})
		})

		t.Run("Given an access token without data", func(t *testing.T) {
			err := accessToken.Write(nil)

			t.Run("Raises an error that the data is empty", func(t *testing.T) {
				assert.EqualError(t, err, log.CAKC005)
			})
		})
	})

	t.Run("Delete", func(t *testing.T) {

		t.Run("Given an access token with data", func(t *testing.T) {
			dataActual := []byte{'t', 'e', 's', 't'}

			t.Run("And the data was written successfully", func(t *testing.T) {
				err := accessToken.Write(dataActual)
				assert.NoError(t, err)

				// Read is added here because we want to check later that the contents were deleted from memory successfully
				t.Run("And the data was read successfully", func(t *testing.T) {
					dataFromRead, err := accessToken.Read()

					t.Run("Finishes without raising an error", func(t *testing.T) {
						assert.NoError(t, err)
					})

					t.Run("When running the Delete method", func(t *testing.T) {
						err := accessToken.Delete()

						t.Run("Finishes without raising an error", func(t *testing.T) {
							assert.NoError(t, err)
						})

						t.Run("Properly clears all data from memory", func(t *testing.T) {
							empty := make([]byte, len(dataActual))
							eq := reflect.DeepEqual(dataActual, empty)
							assert.True(t, eq)
							eq = reflect.DeepEqual(dataFromRead, empty)
							assert.True(t, eq)
						})
					})
				})
			})
		})

		t.Run("Given an access token with no data", func(t *testing.T) {
			accessToken.Data = nil

			t.Run("Finishes without raising an error", func(t *testing.T) {
				err := accessToken.Delete()
				assert.NoError(t, err)
			})
		})

		t.Run("Given two instances of the accessToken interface", func(t *testing.T) {
			// Write Data to source interface
			dataActual := []byte{'t', 'e', 's', 't'}
			accessToken.Write(dataActual)

			t.Run("When setting token file location in proxy struct", func(t *testing.T) {
				var proxyStruct ProxyHandlerTokenMemory
				proxyStruct.AccessToken = accessToken

				t.Run("When running the Delete method", func(t *testing.T) {
					err := proxyStruct.AccessToken.Delete()

					t.Run("Deletes the accessToken file of proxyStruct", func(t *testing.T) {
						assert.NoError(t, err)
					})

					t.Run("When running the Read method", func(t *testing.T) {
						dataExpected, err := accessToken.Read()

						t.Run("Returns no data because data in source interface was cleared", func(t *testing.T) {
							assert.Nil(t, dataExpected)
						})

						t.Run("Raises the proper error", func(t *testing.T) {
							assert.EqualError(t, err, log.CAKC006)
						})
					})
				})
			})
		})
	})
}
