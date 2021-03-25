package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	ExistingFilePath = "path/to/existing/file"
)

func MockVerifyFileExistsFunc(path string) error {
	if path == ExistingFilePath {
		return nil
	}

	return errors.New("NotExist")
}

func TestFile(t *testing.T) {
	Convey("WaitForFile", t, func() {
		retryCountLimit := 10
		Convey("Returns nil if file exists", func() {
			path := ExistingFilePath

			So(
				WaitForFile(
					path,
					retryCountLimit,
					MockVerifyFileExistsFunc,
				),
				ShouldBeNil,
			)
		})

		Convey("Waits for whole time if file does not exist", func() {
			path := "path/to/non-existing/file"

			expectedOutput := fmt.Errorf(
				"CAKC033 Timed out after waiting for %d seconds for file to exist: %s",
				retryCountLimit,
				path,
			)

			So(
				WaitForFile(
					path,
					retryCountLimit,
					MockVerifyFileExistsFunc,
				),
				ShouldResemble,
				expectedOutput,
			)
		})
	})

	Convey("VerifyFileExists", t, func() {
		Convey("An existing file returns nil", func() {
			path := "/tmp/test_file"
			dataStr := "some\ntext\n"
			err := ioutil.WriteFile(path, []byte(dataStr), 0644)
			if err != nil {
				t.FailNow()
			}

			err = VerifyFileExists(path)

			So(err, ShouldBeNil)
		})

		Convey("A folder at the path returns an error", func() {
			path := "/"
			expectedOutput := fmt.Errorf(
				"CAKC058 Path exists but does not contain regular file: %s",
				path,
			)

			err := VerifyFileExists(path)

			So(err, ShouldResemble, expectedOutput)
		})

		Convey("A non-existing file returns an error", func() {
			path := "non/existing/path"

			err := VerifyFileExists(path)

			So(os.IsNotExist(err), ShouldBeTrue)
		})

		Convey("A non-ErrNotExist error is returned", func() {
			err := VerifyFileExists("\000invalid")

			So(err.Error(), ShouldResemble, "stat \x00invalid: invalid argument")
		})
	})
}
