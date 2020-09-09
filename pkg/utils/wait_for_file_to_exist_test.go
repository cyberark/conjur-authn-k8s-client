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
	Convey("WaitForFileToExist", t, func() {
		retryCountLimit := 10
		Convey("Returns nil if cert is installed", func() {
			certificatePath := ExistingFilePath

			So(
				WaitForFileToExist(
					certificatePath,
					retryCountLimit,
					MockVerifyFileExistsFunc,
				),
				ShouldBeNil,
			)
		})

		Convey("Waits for whole time if file does not exist", func() {
			certificatePath := "path/to/non-existing/file"

			expectedOutput := fmt.Errorf(
				"CAKC033E Timed out after waiting for %d seconds for file to exist: %s",
				retryCountLimit,
				certificatePath,
			)

			So(
				WaitForFileToExist(
					certificatePath,
					retryCountLimit,
					MockVerifyFileExistsFunc,
				),
				ShouldResemble,
				expectedOutput,
			)
		})
	})
}
