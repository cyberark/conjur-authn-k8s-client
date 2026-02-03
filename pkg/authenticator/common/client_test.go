package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Note: NewHTTPSClient is tested indirectly through integration tests in k8s/authenticator
// and jwt/authenticator packages, which test the timeout configuration end-to-end.
// This test verifies the timeout parameter is properly passed through the configuration flow.

func TestDurationFromStringFunction(t *testing.T) {
	// Test that verifies timeout duration parsing and configuration
	testCases := []struct {
		description  string
		timeoutStr   string
		expectedMs   int64
		expectedErr  bool
	}{
		{
			description: "parses seconds",
			timeoutStr:  "10s",
			expectedMs:  10000,
			expectedErr: false,
		},
		{
			description: "parses minutes",
			timeoutStr:  "5m",
			expectedMs:  300000,
			expectedErr: false,
		},
		{
			description: "parses hours",
			timeoutStr:  "1h",
			expectedMs:  3600000,
			expectedErr: false,
		},
		{
			description: "parses combined duration",
			timeoutStr:  "2h30m45s",
			expectedMs:  9045000,
			expectedErr: false,
		},
		{
			description: "parses milliseconds",
			timeoutStr:  "500ms",
			expectedMs:  500,
			expectedErr: false,
		},
		{
			description: "returns error on invalid format",
			timeoutStr:  "invalid",
			expectedMs:  0,
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// The durationFromString function is the internal function used to parse timeouts
			duration, err := durationFromString("CONJUR_HTTP_TIMEOUT", tc.timeoutStr)

			if tc.expectedErr {
				assert.NotNil(t, err, "expected error but got nil")
				assert.Equal(t, time.Duration(0), duration)
			} else {
				assert.Nil(t, err, "unexpected error: %v", err)
				assert.Equal(t, tc.expectedMs, duration.Milliseconds())
			}
		})
	}
}

func TestValidTimeoutFunction(t *testing.T) {
	// Test the validTimeout validation function used in ValidateSetting
	testCases := []struct {
		description string
		timeout     string
		shouldErr   bool
	}{
		{
			description: "valid timeout",
			timeout:     "30s",
			shouldErr:   false,
		},
		{
			description: "valid timeout with hours",
			timeout:     "2h",
			shouldErr:   false,
		},
		{
			description: "invalid timeout format",
			timeout:     "not-a-duration",
			shouldErr:   true,
		},
		{
			description: "empty timeout",
			timeout:     "",
			shouldErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := validTimeout("CONJUR_HTTP_TIMEOUT", tc.timeout)

			if tc.shouldErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
