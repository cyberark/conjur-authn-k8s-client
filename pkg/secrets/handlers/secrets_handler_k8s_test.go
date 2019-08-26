package handlers

import (
	"github.com/cyberark/conjur-authn-k8s-client/pkg/secrets/k8s"
	. "github.com/smartystreets/goconvey/convey"
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

			So(err.Error(), ShouldEqual, "error map should not be empty")
		})
	})

	Convey("updateK8sSecretsMapWithConjurSecrets", t, func () {
		secret := []byte{'s', 'u', 'p', 'e', 'r'}
		conjurSecrets := make(map[string][]byte)
		conjurSecrets["account:variable:allowed/username"] = secret

		newDataEntriesMap := make(map[string][]byte)
		newDataEntriesMap["username"] = []byte("allowed/username")

		k8sSecretsMap := make(map[string]map[string][]byte)
		k8sSecretsMap["mysecret"] = newDataEntriesMap

		pathMap := make(map[string]string)
		pathMap["allowed/username"] = "mysecret:username"

		k8sSecretsStruct := k8s.K8sSecretsMap{k8sSecretsMap, pathMap}
		err := updateK8sSecretsMapWithConjurSecrets(&k8sSecretsStruct, conjurSecrets)

		eq := reflect.DeepEqual(k8sSecretsStruct.K8sSecrets["mysecret"]["username"], secret)
		Convey("Returns true if K8sSecretsMap's variable ID is replaced with corresponding secret value", func() {
			So(err, ShouldEqual, nil)
			So(eq, ShouldEqual, true)
		})

		Convey("Returns true if Conjur secret byte array is cleared from memory", func() {
			// secret value should be cleared from memory after use
			empty := make([]byte, len(secret))
			eq = reflect.DeepEqual(conjurSecrets["account:variable:allowed/username"], empty)
			So(eq, ShouldEqual, true)
		})
	})
}
