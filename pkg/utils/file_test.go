package utils

import (
	"errors"
	"fmt"
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
				"CAKC033E Timed out after waiting for %d seconds for file to exist: %s",
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
}
