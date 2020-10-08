package log

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthenticator(t *testing.T) {
	Convey("Logger", t, func() {
		Convey("Calling RecordedError logs the message and return error object with that message", func() {
			validateLog(func(message string, params ...interface{}) {
				err := RecordedError(message, params...)
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf(message, params...))
			}, "ERROR", "log message with param: <%s>", "param value")
		})

		Convey("Calling Error logs the message", func() {
			validateLog(Error, "ERROR", "log message with param: <%s>", "param value")
		})

		Convey("Calling Warn logs the message", func() {
			validateLog(Warn, "WARN", "log message with param: <%s>", "param value")
		})

		Convey("Calling Info logs the message", func() {
			validateLog(Info, "INFO", "log message with param: <%s>", "param value")
		})

		Convey("Calling Debug does nothing before Calling EnableDebugMode", func() {
			var logBuffer bytes.Buffer
			InfoLogger = log.New(&logBuffer, "", 0)

			Debug("message")

			So(logBuffer.Len(), ShouldEqual, 0)
		})

		Convey("Calling Debug logs the message after Calling EnableDebugMode", func() {
			EnableDebugMode()
			validateLog(Debug, "DEBUG", "log message with param: <%s>", "param value")
		})
	})
}

func validateLog(logFunc func(string, ...interface{}), logLevel, messageFormat, param string) {
	// Replace logger with buffer to test its value
	var logBuffer bytes.Buffer
	ErrorLogger = log.New(&logBuffer, "", 0)
	InfoLogger = log.New(&logBuffer, "", 0)

	logFunc(messageFormat, param)

	logMessages := string(logBuffer.Bytes())
	So(logMessages, ShouldContainSubstring, logLevel)
	So(logMessages, ShouldContainSubstring, fmt.Sprintf(messageFormat, param))
}
