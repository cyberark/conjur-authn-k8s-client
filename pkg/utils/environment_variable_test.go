package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	envVarName              = "SOME_ENV_VAR"
	intValueStr             = "20"
	expectedIntValue        = 20
	defaultIntValueStr      = "10"
	defaultIntValue         = 10
	durationValueStr        = "20m0s"
	expectedDurationValue   = 20 * time.Minute
	defaultDurationValueStr = "10m0s"
	defaultDurationValue    = 10 * time.Minute
	InvalidValueStr         = "SOME_ENV_VAR_VALUE"
)

// mockGetEnv structure is used to implement a getEnvFunc function that
// returns a canned value that is set when the structure is initialized.
type mockGetEnv struct {
	value string
}

// newMockGetEnv creates and initializes a mockGetEnv structure
func newMockGetEnv(value string) mockGetEnv {
	return mockGetEnv{value}
}

// getEnv implements the getEnvFunc type for a mockGetEnv struct
func (m mockGetEnv) getEnv(name string) string {
	return m.value
}

func TestEnvironmentVariable(t *testing.T) {
	t.Run("IntFromEnvOrDefault", func(t *testing.T) {
		t.Run("When parsing an int value", func(t *testing.T) {
			t.Run("That was retrieved from the env", func(t *testing.T) {
				mockGetEnv := newMockGetEnv(intValueStr)
				intValue, err := IntFromEnvOrDefault(
					envVarName,
					defaultIntValueStr,
					mockGetEnv.getEnv,
				)

				assert.NoError(t, err)
				assert.EqualValues(t, expectedIntValue, intValue)
			})

			t.Run("That was given as a default value", func(t *testing.T) {
				mockGetEnv := newMockGetEnv("")
				intValue, err := IntFromEnvOrDefault(
					envVarName,
					defaultIntValueStr,
					mockGetEnv.getEnv,
				)

				assert.NoError(t, err)
				assert.EqualValues(t, defaultIntValue, intValue)
			})
		})

		t.Run("When parsing a non-int value", func(t *testing.T) {
			mockGetEnv := newMockGetEnv(InvalidValueStr)
			_, err := IntFromEnvOrDefault(
				envVarName,
				defaultIntValueStr,
				mockGetEnv.getEnv,
			)

			expectedError := fmt.Errorf(
				"CAKC010 Failed to parse %s. Reason: %s",
				envVarName,
				"strconv.Atoi: parsing \"SOME_ENV_VAR_VALUE\": invalid syntax",
			)

			assert.EqualError(t, err, expectedError.Error())
		})
	})

	t.Run("DurationFromEnvOrDefault", func(t *testing.T) {
		t.Run("When parsing a duration value", func(t *testing.T) {
			t.Run("That was retrieved from the env", func(t *testing.T) {
				mockGetEnv := newMockGetEnv(durationValueStr)
				durationValue, err := DurationFromEnvOrDefault(
					envVarName,
					defaultDurationValueStr,
					mockGetEnv.getEnv,
				)

				assert.NoError(t, err)
				assert.EqualValues(t, expectedDurationValue, durationValue)
			})

			t.Run("That was given as a default value", func(t *testing.T) {
				mockGetEnv := newMockGetEnv("")
				durationValue, err := DurationFromEnvOrDefault(
					envVarName,
					defaultDurationValueStr,
					mockGetEnv.getEnv,
				)

				assert.NoError(t, err)
				assert.EqualValues(t, defaultDurationValue, durationValue)
			})
		})

		t.Run("When parsing a non-duration value", func(t *testing.T) {
			mockGetEnv := newMockGetEnv(InvalidValueStr)
			_, err := DurationFromEnvOrDefault(
				envVarName,
				defaultDurationValueStr,
				mockGetEnv.getEnv,
			)

			expectedError := fmt.Errorf(
				"CAKC010 Failed to parse %s. Reason: %s",
				envVarName,
				"time: invalid duration \"SOME_ENV_VAR_VALUE\"",
			)

			assert.EqualError(t, err, expectedError.Error())
		})
	})
}
