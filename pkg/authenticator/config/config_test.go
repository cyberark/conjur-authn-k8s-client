package config

import (
	"bytes"
	"log"
	"os"
	"testing"

	logger "github.com/cyberark/conjur-authn-k8s-client/pkg/log"

	. "github.com/smartystreets/goconvey/convey"
)

var environmentValues = map[string]string{
	"CONJUR_AUTHN_URL":       "filepath",
	"CONJUR_ACCOUNT":         "testAccount",
	"CONJUR_AUTHN_LOGIN":     "host",
	"CONJUR_CERT_FILE":       "testSSLCertFile.txt",
	"CONJUR_SSL_CERTIFICATE": "testSSLCert",
	"CONJUR_VERSION":         "",
	"MY_POD_NAMESPACE":       "testNameSpace",
	"MY_POD_NAME":            "testPodName",
}

func TestAuthenticator(t *testing.T) {
	// Set default environment variables
	for key, value := range environmentValues {
		err := os.Setenv(key, value)
		if err != nil {
			panic(err)
		}
	}

	// Unset variables when finished
	defer func() {
		for key, _ := range environmentValues {
			err := os.Setenv(key, "")
			if err != nil {
				return
			}
		}
	}()

	Convey("FromEnv", t, func() {
		Convey("Debug logs are enabled if DEBUG env var is 'true'", func() {
			_ = os.Setenv("DEBUG", "true")

			// Replace logger with buffer to test its value
			var logBuffer bytes.Buffer
			logger.InfoLogger = log.New(&logBuffer, "", 0)

			_, err := FromEnv(successfulMockReadFile)

			So(err, ShouldNotBeNil)

			logMessages := string(logBuffer.Bytes())
			So(logMessages, ShouldContainSubstring, "DEBUG")
			So(logMessages, ShouldContainSubstring, "CAKC052")
		})

		Convey("Debug logs are disabled if DEBUG env var is not 'true'", func() {
			_ = os.Setenv("DEBUG", "some invalid value")

			// Replace logger with buffer to test its value
			var logBuffer bytes.Buffer
			logger.InfoLogger = log.New(&logBuffer, "", 0)

			_, err := FromEnv(successfulMockReadFile)

			So(err, ShouldNotBeNil)

			logMessages := string(logBuffer.Bytes())
			So(logMessages, ShouldContainSubstring, "WARN")
			So(logMessages, ShouldContainSubstring, "CAKC034")
		})
	})
}

func successfulMockReadFile(filename string) ([]byte, error) {
	return []byte{}, nil
}
