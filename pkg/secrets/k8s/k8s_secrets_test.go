package k8s

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestKubernetesSecrets(t *testing.T) {
	Convey("generateStringDataEntry", t, func() {
		Convey("Returns true if stringData output as expected", func() {
			m := make(map[string][]byte)
			m["user"] = []byte("dummy_user")
			m["password"] = []byte("dummy_password")
			m["address"] = []byte("dummy_address")
			stringDataEntryExpected := `{"stringData":{"user":"dummy_user","password":"dummy_password","address":"dummy_address"}}`

			DataEntry, err := generateStringDataEntry(m)
			stringDataEntryActual := string(DataEntry)

			// Sort actual and expected, because output order can be change
			re := regexp.MustCompile("\\:{(.*?)\\}")
			// Regex example: {"stringData":{"user":"dummy_user","password":"dummy_password"}} => {"user":"dummy_user","password":"dummy_password"}
			match := re.FindStringSubmatch(stringDataEntryActual)
			stringDataEntryActualSorted := strings.Split(match[1], ",")
			sort.Strings(stringDataEntryActualSorted)
			match = re.FindStringSubmatch(stringDataEntryExpected)
			stringDataEntryExpectedSorted := strings.Split(match[1], ",")
			sort.Strings(stringDataEntryExpectedSorted)

			eq := reflect.DeepEqual(stringDataEntryActualSorted, stringDataEntryExpectedSorted)

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
