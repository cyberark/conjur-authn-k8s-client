package access_token

import (
	storageConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

type ProxyHandlerTokenMemory struct {
	AccessToken AccessTokenHandler
}

func TestAccessTokenMemory(t *testing.T) {
	var config storageConfig.Config
	config.StoreType = storageConfig.None
	var tokenInMemory, _ = NewAccessTokenMemory()


	Convey("Read", t, func() {

		Convey("Given an access token with data saved in memory", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			tokenInMemory.Write(dataActual)

			Convey("When running Read method", func() {
				dataExpected, err := tokenInMemory.Read()

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
				})

				Convey("Returns the data the was written", func() {
					eq := reflect.DeepEqual(dataActual, dataExpected)
					So(eq, ShouldEqual, true)
				})
			})
		})

		Convey("Given an access token's data is empty", func() {
			tokenInMemory.Data = nil

			Convey("Raises an error when reading the access token that the data is empty", func() {
				_, err := tokenInMemory.Read()
				So(err.Error(), ShouldEqual, "error reading access token, reason: data is empty")
			})
		})
	})

	Convey("Write", t, func() {

		Convey("Given an access token with data", func() {
			dataActual := []byte{'t', 'e', 's', 't'}

			Convey("Writes the access token to memory without raising an error", func() {
				err := tokenInMemory.Write(dataActual)
				So(err, ShouldEqual, nil)
			})
		})

		Convey("Given an access token without data", func() {
			err := tokenInMemory.Write(nil)

			Convey("Raises an error that the data is empty", func() {
				So(err.Error(), ShouldEqual, "error writing access token, reason: data is empty")
			})
		})
	})

	Convey("Delete", t, func() {

		Convey("Given an access token with data saved in memory", func() {
			dataActual := []byte{'t', 'e', 's', 't'}

			Convey("When running the Write method", func() {
				err := tokenInMemory.Write(dataActual)

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
				})

				Convey("When running the Read method", func() {
					dataFromRead , err := tokenInMemory.Read()

					Convey("Finishes without raising an error", func() {
						So(err, ShouldEqual, nil)
					})

					Convey("When running the Delete method", func() {
						err := tokenInMemory.Delete()

						Convey("Finishes without raising an error", func() {
							So(err, ShouldEqual, nil)
						})

						Convey("Properly clears all data from memory", func() {
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

		Convey("Given an access token with no data", func() {
			tokenInMemory.Data = nil

			Convey("Finishes without raising an error", func() {
				err := tokenInMemory.Delete()
				So(err, ShouldEqual, nil)
			})
		})

		Convey("Given two instances of the accessTokenHandler interface", func() {

			Convey("Finishes without raising errors and data has been deleted", func() {
				// Write Data to source interface
				dataActual := []byte{'t', 'e', 's', 't'}
				tokenInMemory.Write(dataActual)

				// Set proxy struct with source interface
				var proxyStruct ProxyHandlerTokenMemory
				proxyStruct.AccessToken = tokenInMemory

				// Delete access token from proxy
				err := proxyStruct.AccessToken.Delete()
				So(err, ShouldEqual, nil)

				// Data in source interface should be deleted
				dataExpected, err := tokenInMemory.Read()
				So(err.Error(), ShouldEqual, "error reading access token, reason: data is empty")
				So(dataExpected, ShouldEqual, nil)
			})
		})
	})
}
