package config

import (
	"fmt"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUsername(t *testing.T) {
	Convey("NewUsername", t, func() {
		Convey("Given a valid host's username", func() {
			Convey("that is longer than 4 parts", func() {
				username := "host/path/to/policy/namespace/resource_type/resource_id"
				expectedPrefix := "host.path.to.policy"
				expectedSuffix := "namespace.resource_type.resource_id"

				usernameStruct, err := NewUsername(username)
				Convey("Finishes without raising an error", func() {
					So(err, ShouldBeNil)
				})

				Convey("The suffix include the last 3 parts", func() {
					So(usernameStruct.Prefix, ShouldEqual, expectedPrefix)
					So(usernameStruct.Suffix, ShouldEqual, expectedSuffix)
				})
			})

			Convey("that is shorter than 4 parts", func() {
				username := "host/policy/host_id"
				expectedPrefix := "host.policy"
				expectedSuffix := "host_id"

				usernameStruct, err := NewUsername(username)
				Convey("Finishes without raising an error", func() {
					So(err, ShouldBeNil)
				})

				Convey("The suffix includes only the host id", func() {
					So(usernameStruct.Prefix, ShouldEqual, expectedPrefix)
					So(usernameStruct.Suffix, ShouldEqual, expectedSuffix)
				})
			})
		})

		Convey("Given a username that doesn't start with 'host/'", func() {
			username := "namespace/resource_type/resource_id"

			_, err := NewUsername(username)
			Convey("Raises an invalid username error", func() {
				So(err.Error(), ShouldStartWith, "CAKC032")
			})
		})

		Convey("String representation of username only shows full username", func() {
			username := "host/path/to/policy/namespace/resource_type/resource_id"

			usernameStruct, _ := NewUsername(username)
			usernameStructStr := fmt.Sprintf("%s", usernameStruct)
			So(usernameStructStr, ShouldEqual, username)
		})
	})
}
