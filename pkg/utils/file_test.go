package utils

import (
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// Different types of file scenarios to test. (For file scenarios that use
// mock file utilities as configured in the testConfigMap below, the string
// values are arbitrary but must be unique.)
const (
	existingFilePath    = "path/to/existing/file"
	nonexistentFilePath = "path/to/nonexistent/file"
	nonRegularFilePath  = "/tmp"
	unpermittedFilePath = "path/to/unpermitted/file"
	invalidArgsFilePath = "invalid/args/in/file/access"
)

// Test config for a given file scenario
type testConfig struct {
	useOSUtils bool
	statError  error // This is a don't care if useOSUtils is 'true'
	isRegular  bool  // This is a don't care if useOSUtils is 'true'
}

// Map of test configuration based on file scenario
var testConfigMap = map[string]testConfig{
	existingFilePath: testConfig{
		useOSUtils: false,
		statError:  nil,
		isRegular:  true},
	nonexistentFilePath: testConfig{
		useOSUtils: false,
		statError:  os.ErrNotExist,
		isRegular:  false},
	nonRegularFilePath: testConfig{
		useOSUtils: true,
		statError:  nil,
		isRegular:  false},
	unpermittedFilePath: testConfig{
		useOSUtils: false,
		statError:  os.ErrPermission,
		isRegular:  false},
	invalidArgsFilePath: testConfig{
		useOSUtils: false,
		statError:  os.ErrInvalid,
		isRegular:  false},
}

func testCaseFileUtils(path string) *fileUtils {
	config := testConfigMap[path]

	if config.useOSUtils {
		return osFileUtils
	}

	return &fileUtils{
		func(s string) (os.FileInfo, error) {
			return nil, config.statError
		},
		func(info os.FileInfo) bool {
			return config.isRegular
		},
	}
}

func TestFile(t *testing.T) {
	Convey("waitForFile", t, func() {
		retryCountLimit := 10
		Convey("Returns nil if file exists", func() {
			path := existingFilePath
			utilities := testCaseFileUtils(path)

			So(
				waitForFile(
					path,
					retryCountLimit,
					utilities,
				),
				ShouldBeNil,
			)
		})

		Convey("Waits for whole time if file does not exist", func() {
			path := nonexistentFilePath
			utilities := testCaseFileUtils(path)
			expectedOutput := fmt.Errorf(
				"CAKC033 Timed out after waiting for %d seconds for file to exist: %s",
				retryCountLimit,
				path,
			)

			So(
				waitForFile(
					path,
					retryCountLimit,
					utilities,
				),
				ShouldResemble,
				expectedOutput,
			)
		})
	})

	Convey("verifyFileExists", t, func() {
		Convey("An existing file returns nil error", func() {
			path := existingFilePath
			utilities := testCaseFileUtils(path)

			err := verifyFileExists(path, utilities)

			So(err, ShouldBeNil)
		})

		Convey("A path to an unpermitted file returns a logged error", func() {
			path := unpermittedFilePath
			utilities := testCaseFileUtils(path)
			expectedOutput := fmt.Errorf(
				"CAKC058 Permissions error occured when checking if file exists: %s",
				path,
			)

			err := verifyFileExists(path, utilities)

			So(err, ShouldResemble, expectedOutput)
		})

		Convey("A path to a non-regular file returns a logged error", func() {
			path := nonRegularFilePath
			utilities := testCaseFileUtils(path)
			expectedOutput := fmt.Errorf(
				"CAKC059 Path exists but does not contain regular file: %s",
				path,
			)

			err := verifyFileExists(path, utilities)

			So(err, ShouldResemble, expectedOutput)
		})

		Convey("A non-existing file returns an error without logging", func() {
			path := nonexistentFilePath
			utilities := testCaseFileUtils(path)

			err := verifyFileExists(path, utilities)

			So(os.IsNotExist(err), ShouldBeTrue)
		})

		Convey("A non-ErrNotExist error is returned without logging", func() {
			path := invalidArgsFilePath
			utilities := testCaseFileUtils(path)

			err := verifyFileExists(path, utilities)

			So(err.Error(), ShouldResemble, "invalid argument")
		})
	})
}
