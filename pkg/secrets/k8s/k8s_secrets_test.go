package k8s

import (
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestKubernetesSecrets(t *testing.T) {
	Convey("generateStringDataEntry", t, func() {

		Convey("Given a map of data entries", func() {
			m := make(map[string][]byte)
			m["user"] = []byte("dummy_user")
			m["password"] = []byte("dummy_password")
			m["address"] = []byte("dummy_address")
			DataEntry, err := generateStringDataEntry(m)

			Convey("Finishes without raising an error", func() {
				So(err, ShouldEqual, nil)
			})

			Convey("Convert the data entry map to a stringData entry in the form of a comma separate byte array", func() {
				stringDataEntryExpected := `{"stringData":{"user":"dummy_user","password":"dummy_password","address":"dummy_address"}}`
				stringDataEntryActual := string(DataEntry)
				// Sort actual and expected, because output order can change
				re := regexp.MustCompile("\\:{(.*?)\\}")
				// Regex example: {"stringData":{"user":"dummy_user","password":"dummy_password"}} => {"user":"dummy_user","password":"dummy_password"}
				match := re.FindStringSubmatch(stringDataEntryActual)
				stringDataEntryActualSorted := strings.Split(match[1], ",")
				sort.Strings(stringDataEntryActualSorted)
				match = re.FindStringSubmatch(stringDataEntryExpected)
				stringDataEntryExpectedSorted := strings.Split(match[1], ",")
				sort.Strings(stringDataEntryExpectedSorted)
				eq := reflect.DeepEqual(stringDataEntryActualSorted, stringDataEntryExpectedSorted)
				So(eq, ShouldEqual, true)
			})
		})

		Convey("Given a map of data entries and a secret with backslashes", func() {
			m := make(map[string][]byte)
			// need to simulate a conjur secret, with 1 backslash, so we use unicode to simulate this
			m["user"] = []byte("super\u005csecret");
			DataEntry, err := generateStringDataEntry(m);

			Convey("Finishes without raising an error", func() {
				So(err, ShouldEqual, nil)
			})

			Convey("Returns proper password with escaped characters", func() {
				// before sending secret to k8s there are two backslashes
				// k8s cuts off the second backslash so here we are expecting two
				stringDataEntryExpected := `{"stringData":{"user":"super\\secret"}}`
				stringDataEntryActual := string(DataEntry)
				eq := reflect.DeepEqual(stringDataEntryActual, stringDataEntryExpected)
				So(eq, ShouldEqual, true)
			})
		})

		Convey("Given an empty map of data entries", func() {
			m := make(map[string][]byte)

			Convey("Raises an error that the map input should not be empty", func() {
				_, err := generateStringDataEntry(m)
				So(err.Error(), ShouldEqual, log.CAKC039E)
			})
		})
	})
}
