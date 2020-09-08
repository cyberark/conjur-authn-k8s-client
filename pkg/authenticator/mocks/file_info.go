package mocks

import (
	"os"
	"time"
)

type MockFileInfo struct{}

func (fileInfo MockFileInfo) Name() string {
	return "MockName"
}

func (fileInfo MockFileInfo) Size() int64 {
	return 0
}

func (fileInfo MockFileInfo) Mode() os.FileMode {
	return 0000
}

func (fileInfo MockFileInfo) ModTime() time.Time {
	return time.Now()
}

// This one is important as we are verifying this in the code
func (fileInfo MockFileInfo) IsDir() bool {
	return false
}

func (fileInfo MockFileInfo) Sys() interface{} {
	return nil
}
