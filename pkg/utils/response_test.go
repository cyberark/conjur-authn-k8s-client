package utils

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadResponseBody(t *testing.T) {
	testCases := []struct {
		name           string
		responseBody   io.ReadCloser
		expectedResult []byte
		expectError    bool
	}{
		{
			name:           "successful read",
			responseBody:   io.NopCloser(strings.NewReader("test response body")),
			expectedResult: []byte("test response body"),
			expectError:    false,
		},
		{
			name:           "empty response body",
			responseBody:   io.NopCloser(strings.NewReader("")),
			expectedResult: []byte{},
			expectError:    false,
		},
		{
			name:           "json response body",
			responseBody:   io.NopCloser(strings.NewReader(`{"key": "value"}`)),
			expectedResult: []byte(`{"key": "value"}`),
			expectError:    false,
		},
		{
			name:           "error reading body",
			responseBody:   &errorReadCloser{},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				Body: tc.responseBody,
			}

			result, err := ReadResponseBody(resp)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestValidateResponse(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		expectError bool
	}{
		{
			name:        "status 200 - success",
			statusCode:  200,
			expectError: false,
		},
		{
			name:        "status 201 - success",
			statusCode:  201,
			expectError: false,
		},
		{
			name:        "status 300 - error",
			statusCode:  300,
			expectError: true,
		},
		{
			name:        "status 400 - error",
			statusCode:  400,
			expectError: true,
		},
		{
			name:        "status 404 - error",
			statusCode:  404,
			expectError: true,
		},
		{
			name:        "status 500 - error",
			statusCode:  500,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tc.statusCode,
				Body:       io.NopCloser(strings.NewReader("")),
			}

			err := ValidateResponse(resp)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
