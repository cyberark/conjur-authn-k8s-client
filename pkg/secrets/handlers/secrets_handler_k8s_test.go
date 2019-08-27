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

		Convey("Given a non-empty pathMap", func() {
			m := make(map[string]string)
			m["account/var_path1"] = "secret1:key1"
			m["account/var_path2"] = "secret1:key2"
			variableIDsExpected := []string{"account/var_path1", "account/var_path2"}
			variableIDsActual, err := getVariableIDsToRetrieve(m)

			Convey("Finishes without raising an error", func() {
				So(err, ShouldEqual, nil)
			})

			Convey("Returns variable IDs in the pathMap as expected", func() {
				// Sort actual and expected, because output order can change
				sort.Strings(variableIDsActual)
				sort.Strings(variableIDsExpected)
				eq := reflect.DeepEqual(variableIDsActual, variableIDsExpected)
				So(eq, ShouldEqual, true)
			})
		})

		Convey("Given an empty pathMap", func() {
			m := make(map[string]string)

			Convey("Raises an error that the map input is empty", func() {
				_, err := getVariableIDsToRetrieve(m)
				So(err.Error(), ShouldEqual, "error map should not be empty")
			})
		})
	})

	Convey("updateK8sSecretsMapWithConjurSecrets", t, func () {
		Convey("Given one K8s secret with one Conjur secret", func() {
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

			Convey("Finishes without raising an error and the secret variable IDs are replaced in k8sSecretsMap with their corresponding secret value", func() {
				So(err, ShouldEqual, nil)
			})

			Convey("Replaces the secret variable IDs in k8sSecretsMap with their corresponding secret value", func() {
				eq := reflect.DeepEqual(k8sSecretsStruct.K8sSecrets["mysecret"]["username"], secret)
				So(eq, ShouldEqual, true)
			})

			Convey("Clears the Conjur secret byte array from memory after use", func() {
				empty := make([]byte, len(secret))
				eq := reflect.DeepEqual(k8sSecretsStruct.K8sSecrets["mysecret"]["username"], secret)
				eq = reflect.DeepEqual(conjurSecrets["account:variable:allowed/username"], empty)
				So(eq, ShouldEqual, true)
			})
		})
	})
}
