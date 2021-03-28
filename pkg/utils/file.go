package utils

import (
	"os"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// statFunc type is defined so that the dependency 'os.Stat()'
// can be mocked for testing.
type statFunc func(string) (interface{}, error)

func osStat(path string) (interface{}, error) {
	return os.Stat(path)
}

// isRegularFunc type is defined so that the dependency
// 'os.FileInfo.Mode().IsRegular()' can be mocked for testing.
type isRegularFunc func(interface{}) bool

func osIsRegular(fileInfo interface{}) bool {
	osFileInfo := fileInfo.(os.FileInfo)
	return osFileInfo.Mode().IsRegular()
}

type fileUtils struct {
	stat      statFunc
	isRegular isRegularFunc
}

var osFileUtils = &fileUtils{
	osStat,
	osIsRegular,
}

// WaitForFile waits for retryCountLimit seconds to see if the file
// exists in the given path. If it's not there by the end of the retry
// count limit, it returns an error.
func WaitForFile(
	path string,
	retryCountLimit int,
	utilities *fileUtils,
) error {
	limitedBackOff := NewLimitedBackOff(
		time.Second,
		retryCountLimit,
	)

	err := backoff.Retry(func() error {
		if limitedBackOff.RetryCount() > 0 {
			log.Debug(log.CAKC051, path)
		}

		return VerifyFileExists(path, utilities)
	}, limitedBackOff)

	if err != nil {
		return log.RecordedError(log.CAKC033, retryCountLimit, path)
	}

	return nil
}

// VerifyFileExists verifies that a file exists at a given path and is a
// regular file.
func VerifyFileExists(path string, utilities *fileUtils) error {
	if utilities == nil {
		utilities = osFileUtils
	}
	info, err := utilities.stat(path)
	if os.IsPermission(err) {
		// Permissions error when checking if file exists
		return log.RecordedError(log.CAKC058, path)
	}
	if err == nil && !utilities.isRegular(info) {
		// Path exists but does not container regular file
		err = log.RecordedError(log.CAKC059, path)
	}
	return err
}
