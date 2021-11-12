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

		t.Run("Given an access token with data and a defined file location", func(t *testing.T) {
			dataActual := []byte{'t', 'e', 's', 't'}
			accessToken.FilePath = "/tmp/accessTokenFileWrite1"

			t.Run("When running the Write method", func(t *testing.T) {
				err := accessToken.Write(dataActual)

				t.Run("Finishes without raising an error", func(t *testing.T) {
					assert.NoError(t, err)
				})

				t.Run("Checks that the file exists in the path defined", func(t *testing.T) {
					_, err = os.Stat("/tmp/accessTokenFileWrite1")
					assert.NoError(t, err)
				})

				t.Run("And the data was read successfully", func(t *testing.T) {
					dataExpected, _ := accessToken.Read()

					// Confirm data was written
					t.Run("Returns the data the was written to the file", func(t *testing.T) {
						eq := reflect.DeepEqual(dataActual, dataExpected)
						assert.True(t, eq)
					})
				})

				t.Run("When running the Write method a second time", func(t *testing.T) {
					dataActual = []byte{'t', 'e', 's', 't', '2'}
					err := accessToken.Write(dataActual)

					t.Run("The file exists without raising an error", func(t *testing.T) {
						_, err = os.Stat("/tmp/accessTokenFileWrite1")
						assert.NoError(t, err)
					})

					t.Run("Writes the data to the file", func(t *testing.T) {
						// TODO: read the content with `os` methods (not with `accessToken`)
						dataExpected, _ := accessToken.Read()
						eq := reflect.DeepEqual(dataActual, dataExpected)
						assert.True(t, eq)
					})
				})
			})
		})

		t.Run("Given an access token without data", func(t *testing.T) {

			t.Run("When running the Write method", func(t *testing.T) {
				err := accessToken.Write(nil)

				t.Run("Raises an error that the access token data is empty", func(t *testing.T) {
					assert.EqualError(t, err, log.CAKC005)
				})
			})
		})
	})

	t.Run("Delete", func(t *testing.T) {

		t.Run("Given an access token with data", func(t *testing.T) {
			dataActual := []byte{'t', 'e', 's', 't'}

			t.Run("And the data was written successfully", func(t *testing.T) {
				accessToken.FilePath = "/tmp/accessTokenFileDel1"
				err := accessToken.Write(dataActual)
				assert.NoError(t, err)

				// Read is added here because we want to check later that the contents were deleted from memory successfully
				t.Run("And the data was read successfully", func(t *testing.T) {
					dataFromRead, err := accessToken.Read()

					t.Run("Finishes without raising an error", func(t *testing.T) {
						assert.NoError(t, err)
					})

					t.Run("When running the Delete method", func(t *testing.T) {
						err = accessToken.Delete()

						t.Run("Finishes without raising an error", func(t *testing.T) {
							assert.NoError(t, err)
						})

						t.Run("Properly deletes the file", func(t *testing.T) {
							_, err = os.Stat("/tmp/accessTokenFileDel1")
							assert.Error(t, err)
						})

						t.Run("Properly clears all data from memory", func(t *testing.T) {
							// Both input & output arrays should be cleared from memory
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

		t.Run("Given an access token with no data saved to a file", func(t *testing.T) {
			accessToken.Data = nil
			os.Create("/tmp/accessTokenFileDel2")
			accessToken.FilePath = "/tmp/accessTokenFileDel2"

			t.Run("When running the Delete method", func(t *testing.T) {
				err := accessToken.Delete()

				t.Run("Deletes the file and no errors are raised", func(t *testing.T) {
					assert.NoError(t, err)

					// Check that file does not exist
					_, err = os.Stat("/tmp/accessTokenFileDel2")
					assert.Error(t, err)
				})

				t.Run("When running the Delete method again on the same file", func(t *testing.T) {
					err = accessToken.Delete()

					t.Run("Finishes with proper error", func(t *testing.T) {
						assert.Contains(t, err.Error(), log.CAKC002)
					})
				})
			})
		})

		t.Run("Given two instances of the accessToken interface", func(t *testing.T) {
			// Write Data to source interface
			dataActual := []byte{'t', 'e', 's', 't'}
			accessToken.Write(dataActual)

			t.Run("When setting token file location in proxy struct", func(t *testing.T) {
				// Set proxy struct with source interface
				var proxyStruct ProxyHandlerTokenFile
				proxyStruct.AccessToken = accessToken

				t.Run("When running the Delete method", func(t *testing.T) {
					// Delete access token from proxy
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
