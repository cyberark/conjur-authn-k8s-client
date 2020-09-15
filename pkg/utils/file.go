package utils

import (
	"os"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

var defaultVerifyFileExistsFunc = verifyFileExists

type VerifyFileExistsFunc func(path string) error

// WaitForFile waits for retryCountLimit seconds to see if the file
// exists in the given path. If it's not there by the end of the retry count limit, it returns
// an error.
func WaitForFile(
	path string,
	retryCountLimit int,
	verifyFileExistsFunc VerifyFileExistsFunc,
) error {
	if verifyFileExistsFunc == nil {
		verifyFileExistsFunc = defaultVerifyFileExistsFunc
	}

	limitedBackOff := NewLimitedBackOff(
		time.Second,
		retryCountLimit,
	)

	err := backoff.Retry(func() error {
		if limitedBackOff.RetryCount() > 0 {
			log.Info(log.CAKC017I, path)
		}

		return verifyFileExistsFunc(path)
	}, limitedBackOff)

	if err != nil {
		return log.RecordedError(log.CAKC033E, retryCountLimit, path)
	}

	return nil
}

func verifyFileExists(path string) error {
	info, err := os.Stat(path)
	if !os.IsNotExist(err) && info.Mode().IsRegular() {
		// No error, the certificate exists
		return nil
	}

	return err
}
