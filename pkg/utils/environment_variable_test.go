package utils

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
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
	Convey("IntFromEnvOrDefault", t, func() {
		Convey("When parsing an int value", func() {
			Convey("That was retrieved from the env", func() {
				mockGetEnv := newMockGetEnv(intValueStr)
				intValue, err := IntFromEnvOrDefault(
					envVarName,
					defaultIntValueStr,
					mockGetEnv.getEnv,
				)

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
				})

				Convey("Parses the value into an int", func() {
					So(
						intValue,
						ShouldResemble,
						expectedIntValue,
					)
				})
			})

			Convey("That was given as a default value", func() {
				mockGetEnv := newMockGetEnv("")
				intValue, err := IntFromEnvOrDefault(
					envVarName,
					defaultIntValueStr,
					mockGetEnv.getEnv,
				)

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
				})

				Convey("Parses the value into an int", func() {
					So(
						intValue,
						ShouldResemble,
						defaultIntValue,
					)
				})
			})
		})

		Convey("When parsing a non-int value", func() {
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

			Convey("Raises a proper error", func() {
				So(err, ShouldResemble, expectedError)
			})
		})
	})

	Convey("DurationFromEnvOrDefault", t, func() {
		Convey("When parsing a duration value", func() {
			Convey("That was retrieved from the env", func() {
				mockGetEnv := newMockGetEnv(durationValueStr)
				durationValue, err := DurationFromEnvOrDefault(
					envVarName,
					defaultDurationValueStr,
					mockGetEnv.getEnv,
				)

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
				})

				Convey("Parses the value into a duration", func() {
					So(
						durationValue,
						ShouldResemble,
						expectedDurationValue,
					)
				})
			})

			Convey("That was given as a default value", func() {
				mockGetEnv := newMockGetEnv("")
				durationValue, err := DurationFromEnvOrDefault(
					envVarName,
					defaultDurationValueStr,
					mockGetEnv.getEnv,
				)

				Convey("Finishes without raising an error", func() {
					So(err, ShouldEqual, nil)
				})

				Convey("Parses the value into a duration", func() {
					So(
						durationValue,
						ShouldResemble,
						defaultDurationValue,
					)
				})
			})
		})

		Convey("When parsing a non-duration value", func() {
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

			Convey("Raises a proper error", func() {
				So(err, ShouldResemble, expectedError)
			})
		})
	})
}
