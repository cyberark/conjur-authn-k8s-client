package mocks

import (
	"errors"
	"os"
)

func MockOsStatFunc(name string) (os.FileInfo, error) {
	if name == "exist" {
		return &MockFileInfo{}, nil
	} else {
		return nil, errors.New("NotExist")
	}
}

func MockOsIsNotExistFunc(err error) bool {
	return err != nil && err.Error() == "NotExist"
}
