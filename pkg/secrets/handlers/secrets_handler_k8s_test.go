package handlers

import (
	. "github.com/smartystreets/goconvey/convey"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
	"reflect"
	"sort"
	"testing"
)

func TestSecretsHandlerK8sUseCase(t *testing.T) {
	Convey("getVariableIDsToRetrieve", t, func() {
		Convey("Returns true if pathMap output ia as expected", func() {
			m := make(map[string]string)

			m["account/var_path1"] = "secret1:key1"
			m["account/var_path2"] = "secret1:key2"
			variableIDsExpected := []string{"account/var_path1", "account/var_path2"}
			variableIDsActual, err := getVariableIDsToRetrieve(m)

			// Sort actual and expected, because output order can change
			sort.Strings(variableIDsActual)
			sort.Strings(variableIDsExpected)
			eq := reflect.DeepEqual(variableIDsActual, variableIDsExpected)

			So(err, ShouldEqual, nil)
			So(eq, ShouldEqual, true)
		})

		Convey("Returns error if map input is empty", func() {
			m := make(map[string]string)
			_, err := getVariableIDsToRetrieve(m)

			So(err.Error(), ShouldEqual, log.CAKC029E)
		})
	})
}
