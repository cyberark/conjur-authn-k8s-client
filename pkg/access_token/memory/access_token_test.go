package memory

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

type ProxyHandlerTokenMemory struct {
	AccessToken access_token.AccessToken
}

func TestAccessTokenMemory(t *testing.T) {
	var accessToken, _ = NewAccessToken()

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

		Convey("Given an access token with data", func() {
			dataActual := []byte{'t', 'e', 's', 't'}

			Convey("Writes the access token to memory without raising an error", func() {
				err := accessToken.Write(dataActual)
				So(err, ShouldEqual, nil)
			})

			Convey("When running Read method", func() {
				dataExpected, _ := accessToken.Read()

				// Confirm data was written
				Convey("Returns the data the was written", func() {
					eq := reflect.DeepEqual(dataActual, dataExpected)
					So(eq, ShouldEqual, true)
				})
			})
		})

		Convey("Given an access token without data", func() {
			err := accessToken.Write(nil)

			Convey("Raises an error that the data is empty", func() {
				So(err.Error(), ShouldEqual, log.CAKC005)
			})
		})
	})

	Convey("Delete", t, func() {

		Convey("Given an access token with data", func() {
			dataActual := []byte{'t', 'e', 's', 't'}

			Convey("And the data was written successfully", func() {
				err := accessToken.Write(dataActual)
				So(err, ShouldEqual, nil)

				// Read is added here because we want to check later that the contents were deleted from memory successfully
				Convey("And the data was read successfully", func() {
					dataFromRead, err := accessToken.Read()

					Convey("Finishes without raising an error", func() {
						So(err, ShouldEqual, nil)
					})

					Convey("When running the Delete method", func() {
						err := accessToken.Delete()

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
			accessToken.Data = nil

			Convey("Finishes without raising an error", func() {
				err := accessToken.Delete()
				So(err, ShouldEqual, nil)
			})
		})

		Convey("Given two instances of the accessToken interface", func() {
			// Write Data to source interface
			dataActual := []byte{'t', 'e', 's', 't'}
			accessToken.Write(dataActual)

			Convey("When setting token file location in proxy struct", func() {
				var proxyStruct ProxyHandlerTokenMemory
				proxyStruct.AccessToken = accessToken

				Convey("When running the Delete method", func() {
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
