package secrets

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestConjurSecrets(t *testing.T) {
	Convey("GetVariableIDsToRetrieve", t, func() {
		Convey("Returns true if pathMap output ia as expected", func() {
			m := make(map[string]string)

			m["account/var_path1"] = "secret1:key1"
			m["account/var_path2"] = "secret1:key2"
			variableIDsExpected := []string{"account/var_path1", "account/var_path2"}
			variableIDsActual, err := GetVariableIDsToRetrieve(m)

			eq := reflect.DeepEqual(variableIDsActual, variableIDsExpected)

			So(err, ShouldEqual, nil)
			So(eq, ShouldEqual, true)
		})

		Convey("Returns error if map input is empty", func() {
			m := make(map[string]string)
			_, err := GetVariableIDsToRetrieve(m)

			So(err.Error(), ShouldEqual, "error map should not be empty")
		})
	})
}
