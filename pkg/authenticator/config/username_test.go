package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUsername(t *testing.T) {
	Convey("NewUsername", t, func() {
		Convey("Given a valid host's username", func() {
			username := "host/path/to/policy/namespace/resource_type/resource_id"
			expectedPrefix := "host.path.to.policy"
			expectedSuffix := "namespace.resource_type.resource_id"

			usernameStruct, err := NewUsername(username)
			Convey("Finishes without raising an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Splits the username as expected", func() {
				So(usernameStruct.Prefix, ShouldEqual, expectedPrefix)
				So(usernameStruct.Suffix, ShouldEqual, expectedSuffix)
			})
		})

		Convey("Given a username that doesn't have the machine identity inside it", func() {
			username := "host/username"

			_, err := NewUsername(username)
			Convey("Raises an invalid username error", func() {
				So(err.Error(), ShouldStartWith, "CAKC032E")
			})

			// Test a username that has a 2-part machine identity instead of 3
			username = "host/namespace/resource_type"

			_, err = NewUsername(username)
			Convey("Raises the same error", func() {
				So(err.Error(), ShouldStartWith, "CAKC032E")
			})
		})

		Convey("Given a username that doesn't start with 'host/'", func() {
			username := "namespace/resource_type/resource_id"

			_, err := NewUsername(username)
			Convey("Raises an invalid username error", func() {
				So(err.Error(), ShouldStartWith, "CAKC032E")
			})
		})
	})
}
