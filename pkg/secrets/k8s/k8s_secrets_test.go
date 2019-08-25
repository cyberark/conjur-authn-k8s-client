package k8s

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestKubernetesSecrets(t *testing.T) {
	Convey("generateStringDataEntry", t, func() {
		Convey("Returns true if stringData output as expected", func() {
			m := make(map[string][]byte)
			m["user"] = []byte("dummy_user")
			m["password"] = []byte("dummy_password")
			stringDataEntryExpected := `{"stringData":{"user":"dummy_user","password":"dummy_password"}}`

			DataEntry, err := generateStringDataEntry(m)
			stringDataEntryActual := string(DataEntry)
			eq := reflect.DeepEqual(stringDataEntryActual, stringDataEntryExpected)

			So(err, ShouldEqual, nil)
			So(eq, ShouldEqual, true)
		})

		Convey("Returns error if map input is empty", func() {
			m := make(map[string][]byte)
			_, err := generateStringDataEntry(m)

			So(err.Error(), ShouldEqual, "error map should not be empty")
		})
	})
}
