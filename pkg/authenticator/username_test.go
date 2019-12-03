package authenticator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUsername(t *testing.T) {
	Convey("NewUsername", t, func() {
		Convey("Given a host's username", func() {
			username := "host/path/to/policy/namespace/service_type/service_id"
			expectedPrefix := "host.path.to.policy"
			expectedSuffix := "namespace.service_type.service_id"

			usernameStruct := NewUsername(username)

			Convey("Splits the username as expected", func() {
				So(usernameStruct.Prefix, ShouldEqual, expectedPrefix)
				So(usernameStruct.Suffix, ShouldEqual, expectedSuffix)
			})
		})
	})
}
