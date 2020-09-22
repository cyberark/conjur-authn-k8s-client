package file

import (
	"os"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

type ProxyHandlerTokenFile struct {
	AccessToken access_token.AccessToken
}

func TestAccessTokenFile(t *testing.T) {
	var accessToken, _ = NewAccessToken("/tmp/accessTokenFile1")

	Convey("Read", t, func() {

		Convey("Given an access token with data", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			err := accessToken.Write(dataActual)

			Convey("Finishes without raising an error", func() {
				So(err, ShouldEqual, nil)
			})

			Convey("When running Read method", func() {
				dataExpected, err := accessToken.Read()

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
				})

				Convey("Returns the data that was written", func() {
					eq := reflect.DeepEqual(dataActual, dataExpected)
					So(eq, ShouldEqual, true)
				})
			})
		})

		Convey("Given an access token's data is empty", func() {
			accessToken.Data = nil

			Convey("When running the Read method", func() {
				_, err := accessToken.Read()

				Convey("Raises an error that the data is empty", func() {
					So(err.Error(), ShouldEqual, log.CAKC006)
				})
			})
		})
	})

	Convey("Write", t, func() {

		Convey("Given an access token with data and a defined file location", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			accessToken.FilePath = "/tmp/accessTokenFileWrite1"

			Convey("When running the Write method", func() {
				err := accessToken.Write(dataActual)

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
				})

				Convey("Checks that the file exists in the path defined", func() {
					_, err = os.Stat("/tmp/accessTokenFileWrite1")
					So(err, ShouldEqual, nil)
				})

				Convey("And the data was read successfully", func() {
					dataExpected, _ := accessToken.Read()

					// Confirm data was written
					Convey("Returns the data the was written to the file", func() {
						eq := reflect.DeepEqual(dataActual, dataExpected)
						So(eq, ShouldEqual, true)
					})
				})

				Convey("When running the Write method a second time", func() {
					dataActual = []byte{'t', 'e', 's', 't', '2'}
					err := accessToken.Write(dataActual)

					Convey("The file exists without raising an error", func() {
						_, err = os.Stat("/tmp/accessTokenFileWrite1")
						So(err, ShouldEqual, nil)
					})

					Convey("Writes the data to the file", func() {
						// TODO: read the content with `os` methods (not with `accessToken`)
						dataExpected, _ := accessToken.Read()
						eq := reflect.DeepEqual(dataActual, dataExpected)
						So(eq, ShouldEqual, true)
					})
				})
			})
		})

		Convey("Given an access token without data", func() {

			Convey("When running the Write method", func() {
				err := accessToken.Write(nil)

				Convey("Raises an error that the access token data is empty", func() {
					So(err.Error(), ShouldEqual, log.CAKC005)
				})
			})
		})
	})

	Convey("Delete", t, func() {

		Convey("Given an access token with data", func() {
			dataActual := []byte{'t', 'e', 's', 't'}

			Convey("And the data was written successfully", func() {
				accessToken.FilePath = "/tmp/accessTokenFileDel1"
				err := accessToken.Write(dataActual)
				So(err, ShouldEqual, nil)

				// Read is added here because we want to check later that the contents were deleted from memory successfully
				Convey("And the data was read successfully", func() {
					dataFromRead, err := accessToken.Read()

					Convey("Finishes without raising an error", func() {
						So(err, ShouldEqual, nil)
					})

					Convey("When running the Delete method", func() {
						err = accessToken.Delete()

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
			accessToken.Data = nil
			os.Create("/tmp/accessTokenFileDel2")
			accessToken.FilePath = "/tmp/accessTokenFileDel2"

			Convey("When running the Delete method", func() {
				err := accessToken.Delete()

				Convey("Deletes the file and no errors are raised", func() {
					So(err, ShouldEqual, nil)

					// Check that file does not exist
					_, err = os.Stat("/tmp/accessTokenFileDel2")
					So(err, ShouldNotEqual, nil)
				})

				Convey("When running the Delete method again on the same file", func() {
					err = accessToken.Delete()

					Convey("Finishes with proper error", func() {
						So(err.Error(), ShouldContainSubstring, log.CAKC002)
					})
				})
			})
		})

		Convey("Given two instances of the accessToken interface", func() {
			// Write Data to source interface
			dataActual := []byte{'t', 'e', 's', 't'}
			accessToken.Write(dataActual)

			Convey("When setting token file location in proxy struct", func() {
				// Set proxy struct with source interface
				var proxyStruct ProxyHandlerTokenFile
				proxyStruct.AccessToken = accessToken

				Convey("When running the Delete method", func() {
					// Delete access token from proxy
					err := proxyStruct.AccessToken.Delete()

					Convey("Deletes the accessToken file of proxyStruct", func() {
						So(err, ShouldEqual, nil)
					})

					Convey("When running the Read method", func() {
						dataExpected, err := accessToken.Read()

						Convey("Returns no data because data in source interface was cleared", func() {
							So(dataExpected, ShouldEqual, nil)
						})

						Convey("Raises the proper error", func() {
							So(err.Error(), ShouldEqual, log.CAKC006)
						})
					})
				})
			})
		})
	})
}
