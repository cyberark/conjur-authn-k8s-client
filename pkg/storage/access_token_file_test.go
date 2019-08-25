package storage

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
		Convey("Returns true if Data output as expected", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			err := tokenInFile.Write(dataActual)
			So(err, ShouldEqual, nil)

			dataExpected, err := tokenInFile.Read()
			eq := reflect.DeepEqual(dataActual, dataExpected)

			So(err, ShouldEqual, nil)
			So(eq, ShouldEqual, true)
		})

		Convey("Returns error if Data is nil", func() {
			tokenInFile.Data = nil
			_, err := tokenInFile.Read()

			So(err.Error(), ShouldEqual, "error reading access token, reason: data is empty")
		})
	})

	Convey("Write", t, func() {
		Convey("Returns no error if Data input is not nil", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			tokenInFile.TokenFilePath = "/tmp/accessTokenFileWrite1"
			err := tokenInFile.Write(dataActual)

			So(err, ShouldEqual, nil)
			// Check if file exits
			_, err = os.Stat("/tmp/accessTokenFileWrite1")
			So(err, ShouldEqual, nil)
		})

		Convey("Returns error if Data input is nil", func() {
			err := tokenInFile.Write(nil)

			So(err.Error(), ShouldEqual, "error writing access token, reason: data is empty")
		})

		Convey("Returns no error if file already exists", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			tokenInFile.TokenFilePath = "/tmp/accessTokenFileWrite2"
			err := tokenInFile.Write(dataActual)
			So(err, ShouldEqual, nil)

			// Check if file exits
			_, err = os.Stat("/tmp/accessTokenFileWrite1")
			So(err, ShouldEqual, nil)

			// Second time
			dataActual = []byte{'t', 'e', 's', 't', '2'}
			err = tokenInFile.Write(dataActual)
			So(err, ShouldEqual, nil)

			// Check if file exits
			_, err = os.Stat("/tmp/accessTokenFileWrite1")
			So(err, ShouldEqual, nil)

			// Test we are reading the new Data
			dataExpected, err := tokenInFile.Read()
			eq := reflect.DeepEqual(dataActual, dataExpected)

			So(err, ShouldEqual, nil)
			So(eq, ShouldEqual, true)
		})
	})

	Convey("Delete", t, func() {
		Convey("Returns no error after read write and delete", func() {
			dataActual := []byte{'t', 'e', 's', 't'}
			tokenInFile.TokenFilePath = "/tmp/accessTokenFileDel1"
			err := tokenInFile.Write(dataActual)
			So(err, ShouldEqual, nil)

			_, err = tokenInFile.Read()
			So(err, ShouldEqual, nil)

			err = tokenInFile.Delete()
			So(err, ShouldEqual, nil)

			// Check that file is not exits
			_, err = os.Stat("/tmp/accessTokenFileDel1")
			So(err, ShouldNotEqual, nil)
		})

		Convey("Returns no error if Data input is nil", func() {
			tokenInFile.Data = nil
			os.Create("/tmp/accessTokenFileDel2")
			tokenInFile.TokenFilePath = "/tmp/accessTokenFileDel2"
			err := tokenInFile.Delete()
			So(err, ShouldEqual, nil)

			// Check that file is not exits
			_, err = os.Stat("/tmp/accessTokenFileDel2")
			So(err, ShouldNotEqual, nil)
		})

		Convey("Returns error if file already deleted", func() {
			tokenInFile.Data = nil
			os.Create("/tmp/accessTokenFileDel3")
			tokenInFile.TokenFilePath = "/tmp/accessTokenFileDel3"
			err := tokenInFile.Delete()
			So(err, ShouldEqual, nil)

			// Check that file is not exits
			err = tokenInFile.Delete()
			So(err.Error(), ShouldEqual, "error deleting access token")
		})

		Convey("Returns no error if delete data from proxy struct is as expected", func() {
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
}
