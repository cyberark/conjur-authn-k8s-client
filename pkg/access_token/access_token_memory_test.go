package access_token

import (
	storageConfig "github.com/cyberark/conjur-authn-k8s-client/pkg/storage/config"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
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

			So(err.Error(), ShouldEqual, log.CAKC010E)
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

			So(err.Error(), ShouldEqual, log.CAKC009E)
		})
	})

	Convey("Delete", t, func() {
		Convey("Returns no error after read write and delete", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			err := tokenInMemory.Write(dataActual)
			So(err, ShouldEqual, nil)

			dataFromRead, err := tokenInMemory.Read()
			So(err, ShouldEqual, nil)

			err = tokenInMemory.Delete()
			So(err, ShouldEqual, nil)

			// Both input & output arrays should be cleared from memory
			empty := make([]byte, len(dataActual))
			eq := reflect.DeepEqual(dataActual, empty)
			So(eq, ShouldEqual, true)
			eq = reflect.DeepEqual(dataFromRead, empty)
			So(eq, ShouldEqual, true)
		})

		Convey("Returns no error if Data input is nil", func() {
			tokenInMemory.Data = nil
			err := tokenInMemory.Delete()

			So(err, ShouldEqual, nil)
		})

		Convey("Returns no error if delete from proxy struct is as expected", func() {
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
			So(err.Error(), ShouldEqual, log.CAKC010E)
			So(dataExpected, ShouldEqual, nil)
		})
	})
}
