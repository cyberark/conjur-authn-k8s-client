package log

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticator(t *testing.T) {
	t.Run("Logger", func(t *testing.T) {
		t.Run("Calling RecordedError logs the message and return error object with that message", func(t *testing.T) {
			validateLog(t, func(message string, params ...interface{}) {
				err := RecordedError(message, params...)
				assert.Contains(t, err.Error(), fmt.Sprintf(message, params...))
			}, "ERROR", "log message with param: <%s>", "param value")
		})

		t.Run("Calling Error logs the message", func(t *testing.T) {
			validateLog(t, Error, "ERROR", "log message with param: <%s>", "param value")
		})

		t.Run("Calling Warn logs the message", func(t *testing.T) {
			validateLog(t, Warn, "WARN", "log message with param: <%s>", "param value")
		})

		t.Run("Calling Info logs the message", func(t *testing.T) {
			validateLog(t, Info, "INFO", "log message with param: <%s>", "param value")
		})

		t.Run("Calling Debug does nothing before Calling EnableDebugMode", func(t *testing.T) {
			var logBuffer bytes.Buffer
			InfoLogger = log.New(&logBuffer, "", 0)

			Debug("message")

			assert.Equal(t, logBuffer.Len(), 0)
		})

		t.Run("Calling Debug logs the message after Calling EnableDebugMode", func(t *testing.T) {
			EnableDebugMode()
			validateLog(t, Debug, "DEBUG", "log message with param: <%s>", "param value")
		})
	})
}

func validateLog(t *testing.T, logFunc func(string, ...interface{}), logLevel, messageFormat, param string) {
	// Replace logger with buffer to test its value
	var logBuffer bytes.Buffer
	ErrorLogger = log.New(&logBuffer, "", 0)
	InfoLogger = log.New(&logBuffer, "", 0)

	logFunc(messageFormat, param)

	logMessages := logBuffer.String()
	assert.Contains(t, logMessages, logLevel)
	assert.Contains(t, logMessages, fmt.Sprintf(messageFormat, param))
}
