package authenticator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRequests(t *testing.T) {
	Convey("LoginRequest", t, func() {
		authnURL := "dummyURL"
		csrBytes := []byte("dummyCSRBytes")

		Convey("Given a host's username prefix", func() {
			usernamePrefix := "host.path.to.policy"

			req, err := LoginRequest(authnURL, csrBytes, usernamePrefix)
			Convey("Finishes without raising an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Includes the username prefix in the 'Host-Id-Prefix' header", func() {
				So(req.Header.Get("Host-Id-Prefix"), ShouldEqual, usernamePrefix)
			})
		})
	})
}
