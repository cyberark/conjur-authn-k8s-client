package access_token

import (
	"os"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type ProxyHandlerTokenFile struct {
	AccessToken AccessToken
}

func TestAccessTokenFile(t *testing.T) {
	var tokenInFile, _ = NewAccessTokenFile("/tmp/accessTokenFile1")

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

		Convey("Given an access token's data is empty", func() {
			tokenInFile.Data = nil

			Convey("When running the Read method", func() {
				_, err := tokenInFile.Read()

				Convey("Raises an error that the data is empty", func() {
					So(err.Error(), ShouldEqual, "error reading access token, reason: data is empty")
				})
			})
		})
	})

	Convey("Write", t, func() {

		Convey("Given an access token with data and a defined file location", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			tokenInFile.TokenFilePath = "/tmp/accessTokenFileWrite1"

			Convey("When running the Write method", func() {
				err := tokenInFile.Write(dataActual)

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
				})

				Convey("Checks that the file exists in the path defined", func() {
					_, err = os.Stat("/tmp/accessTokenFileWrite1")
					So(err, ShouldEqual, nil)
				})

				Convey("And the data was read successfully", func() {
					dataExpected, _ := tokenInFile.Read()

					// Confirm data was written
					Convey("Returns the data the was written to the file", func() {
						eq := reflect.DeepEqual(dataActual, dataExpected)
						So(eq, ShouldEqual, true)
					})
				})

				Convey("When running the Write method a second time", func() {
					dataActual = []byte{'t', 'e', 's', 't', '2'}
					err := tokenInFile.Write(dataActual)

					Convey("The file exists without raising an error", func() {
						_, err = os.Stat("/tmp/accessTokenFileWrite1")
						So(err, ShouldEqual, nil)
					})

					Convey("Writes the data to the file", func() {
						// TODO: read the content with `os` methods (not with `tokenInFile`)
						dataExpected, _ := tokenInFile.Read()
						eq := reflect.DeepEqual(dataActual, dataExpected)
						So(eq, ShouldEqual, true)
					})
				})
			})
		})

		Convey("Given an access token without data", func() {

			Convey("When running the Write method", func() {
				err := tokenInFile.Write(nil)

				Convey("Raises an error that the access token data is empty", func() {
					So(err.Error(), ShouldEqual, "error writing access token, reason: data is empty")
				})
			})
		})
	})

	Convey("Delete", t, func() {

		Convey("Given an access token with data", func() {
			dataActual := []byte{'t', 'e', 's', 't'}

			Convey("And the data was written successfully", func() {
				tokenInFile.TokenFilePath = "/tmp/accessTokenFileDel1"
				err := tokenInFile.Write(dataActual)
				So(err, ShouldEqual, nil)

				// Read is added here because we want to check later that the contents were deleted from memory successfully
				Convey("And the data was read successfully", func() {
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

					Convey("Finishes with proper error", func() {
						So(err.Error(), ShouldEqual, "error deleting access token")
					})
				})
			})
		})

		Convey("Given two instances of the accessTokenHandler interface", func() {
			// Write Data to source interface
			dataActual := []byte{'t', 'e', 's', 't'}
			tokenInFile.Write(dataActual)

			Convey("When setting token file location in proxy struct", func() {
				// Set proxy struct with source interface
				var proxyStruct ProxyHandlerTokenFile
				proxyStruct.AccessToken = tokenInFile

				Convey("When running the Delete method", func() {
					// Delete access token from proxy
					err := proxyStruct.AccessToken.Delete()

					Convey("Deletes the accessToken file of proxyStruct", func() {
						So(err, ShouldEqual, nil)
					})

					Convey("When running the Read method", func() {
						dataExpected, err := tokenInFile.Read()

						Convey("Returns no data because data in source interface was cleared", func() {
							So(dataExpected, ShouldEqual, nil)
						})

						Convey("Raises the proper error", func() {
							So(err.Error(), ShouldEqual, "error reading access token, reason: data is empty")
						})
					})
				})
			})
		})
	})
}
