package access_token

import (
	storageConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"reflect"
	"testing"
)

type ProxyHandlerTokenFile struct {
	AccessToken AccessTokenHandler
}

func TestAccessTokenFile(t *testing.T) {
	var config storageConfig.Config
	config.StoreType = storageConfig.K8S
	config.TokenFilePath = "/tmp/accessTokenFile1"
	var tokenInFile, _ = NewAccessTokenFile(config)

	Convey("Read", t, func() {

		Convey("Given an access token with data", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			err := tokenInFile.Write(dataActual)

			Convey("Finishes without raising an error", func() {
				So(err, ShouldEqual, nil)
			})

			Convey("When running Read method", func() {
				dataExpected, err := tokenInFile.Read()

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
				})

				Convey("Returns the data that was written", func() {
					eq := reflect.DeepEqual(dataActual, dataExpected)
					So(eq, ShouldEqual, true)
				})
			})
		})

		Convey("Given an access token's data is empty", func () {
			tokenInFile.Data = nil

			Convey("When the token data is empty and running the Read method", func() {
				_, err := tokenInFile.Read()

				Convey("Raises an error that the data is empty", func() {
					So(err.Error(), ShouldEqual, "error reading access token, reason: data is empty")
				})
			})
		})
	})

	Convey("Write", t, func() {

		Convey("Given an access token with data", func() {
			dataActual := []byte{'t', 'e', 's', 't'}

			Convey("When running the Write method, writes the access token to a file in the path defined", func() {
				tokenInFile.TokenFilePath = "/tmp/accessTokenFileWrite1"
				err := tokenInFile.Write(dataActual)

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
					// Check if file exists
					_, err = os.Stat("/tmp/accessTokenFileWrite1")
					So(err, ShouldEqual, nil)
				})
			})
		})

		Convey("Given an access token without data", func() {

			Convey("When running the Write method and attempts to write the access token", func() {
				err := tokenInFile.Write(nil)

				Convey("Raises an error that the access token data is empty", func() {
					So(err.Error(), ShouldEqual, "error writing access token, reason: data is empty")
				})
			})
		})

		Convey("Given access token with data and a file location defined", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			tokenInFile.TokenFilePath = "/tmp/accessTokenFileWrite2"

			Convey("When running the Write method", func() {
				err := tokenInFile.Write(dataActual)

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)

					// Check if file exists
					_, err = os.Stat("/tmp/accessTokenFileWrite2")
					So(err, ShouldEqual, nil)
				})
			})

			Convey("When running the Write method a second time", func() {
				dataActual = []byte{'t', 'e', 's', 't', '2'}
				err := tokenInFile.Write(dataActual)
				So(err, ShouldEqual, nil)

				// Check if file exits
				_, err = os.Stat("/tmp/accessTokenFileWrite2")
				So(err, ShouldEqual, nil)

				Convey("Raises no error and the data should be written to the file as expected", func() {
					// Test we are reading the new Data
					dataExpected, err := tokenInFile.Read()
					eq := reflect.DeepEqual(dataActual, dataExpected)

					So(err, ShouldEqual, nil)
					So(eq, ShouldEqual, true)
				})

			})
		})
	})

	Convey("Delete", t, func() {

		Convey("Given access token with data", func() {
			dataActual := []byte{'t', 'e', 's', 't'}

			Convey("When running the Write method", func() {
				tokenInFile.TokenFilePath = "/tmp/accessTokenFileDel1"
				err := tokenInFile.Write(dataActual)

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
				})

				Convey("When running the Read method", func() {
					dataFromRead, err := tokenInFile.Read()

					Convey("Finishes without raising an error", func() {
						So(err, ShouldEqual, nil)
					})

					Convey("When running the Delete method", func() {
						err = tokenInFile.Delete()

						Convey("Finishes without raising an error", func() {
							So(err, ShouldEqual, nil)
						})

						Convey("Properly deletes the file", func() {
							_, err = os.Stat("/tmp/accessTokenFileDel1")
							So(err, ShouldNotEqual, nil)
						})

						Convey("Properly clears all data from memory", func() {
							// Both input & output arrays should be cleared from memory
							empty := make([]byte, len(dataActual))
							eq := reflect.DeepEqual(dataActual, empty)
							So(eq, ShouldEqual, true)
							eq = reflect.DeepEqual(dataFromRead, empty)
							So(eq, ShouldEqual, true)
						})
					})
				})
			})
		})

		Convey("Given an access token with no data saved to a file", func() {
			tokenInFile.Data = nil
			os.Create("/tmp/accessTokenFileDel2")
			tokenInFile.TokenFilePath = "/tmp/accessTokenFileDel2"

			Convey("When running the Delete method", func() {
				err := tokenInFile.Delete()

				Convey("Deletes the file and no errors are raised", func() {
					So(err, ShouldEqual, nil)

					// Check that file does not exist
					_, err = os.Stat("/tmp/accessTokenFileDel2")
					So(err, ShouldNotEqual, nil)
				})

				Convey("When running the Delete method again on the same file", func() {
					err = tokenInFile.Delete()

					Convey("Finishes with an error deleting the access token", func() {
						So(err.Error(), ShouldEqual, "error deleting access token")
					})
				})
			})
		})

		Convey("Given two instances of the accessTokenHandler interface", func() {

			Convey("Finishes without raising errors and data has been deleted", func() {
				// Write Data to source interface
				dataActual := []byte{'t', 'e', 's', 't'}
				tokenInFile.Write(dataActual)

				// Set proxy struct with source interface
				var proxyStruct ProxyHandlerTokenFile
				proxyStruct.AccessToken = tokenInFile

				// Delete access token from proxy
				err := proxyStruct.AccessToken.Delete()
				So(err, ShouldEqual, nil)

				// Data in source interface should be deleted
				dataExpected, err := tokenInFile.Read()
				So(err.Error(), ShouldEqual, "error reading access token, reason: data is empty")
				So(dataExpected, ShouldEqual, nil)
			})
		})
	})
}
