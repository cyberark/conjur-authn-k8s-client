package storage

import (
	stoargeConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestAccessTokenMemory(t *testing.T) {
	var config stoargeConfig.Config
	config.StoreType = stoargeConfig.None
	var tokenInMemory, _ = NewAccessTokenMemory(config)

	Convey("Read", t, func() {
		Convey("Returns true if Data output as expected", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			tokenInMemory.Write(dataActual)
			dataExpected, err := tokenInMemory.Read()
			eq := reflect.DeepEqual(dataActual, dataExpected)

			So(err, ShouldEqual, nil)
			So(eq, ShouldEqual, true)
		})

		Convey("Returns error if Data is nil", func() {
			tokenInMemory.Data = nil
			_, err := tokenInMemory.Read()

			So(err.Error(), ShouldEqual, "error reading access token, reason: data is empty")
		})
	})

	Convey("Write", t, func() {
		Convey("Returns no error if Data input is not nil", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			err := tokenInMemory.Write(dataActual)

			So(err, ShouldEqual, nil)
		})

		Convey("Returns error if Data input is nil", func() {
			err := tokenInMemory.Write(nil)

			So(err.Error(), ShouldEqual, "error writing access token, reason: data is empty")
		})
	})

	Convey("Delete", t, func() {
		Convey("Returns no error after read write and delete", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			err := tokenInMemory.Write(dataActual)
			So(err, ShouldEqual, nil)

			_, err = tokenInMemory.Read()
			So(err, ShouldEqual, nil)

			err = tokenInMemory.Delete()
			So(err, ShouldEqual, nil)
		})

		Convey("Returns no error if Data input is nil", func() {
			tokenInMemory.Data = nil
			err := tokenInMemory.Delete()

			So(err, ShouldEqual, nil)
		})
	})
}
