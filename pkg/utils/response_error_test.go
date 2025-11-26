package utils

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock ReadCloser for testing that always returns an error on Read
type errorReadCloser struct{}

func (m *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
func (m *errorReadCloser) Close() error {
	return nil
}

func TestResponseError(t *testing.T) {
	testCases := []struct {
		name                 string
		resp                 *http.Response
		expectedErrorMessage string
	}{
		{
			name: "Valid JSON error response",
			resp: &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader(`{"error": {"code": "NotFound", "message": "Resource not found"}}`)),
			},
			expectedErrorMessage: "status code 404, Resource not found",
		},
		{
			name: "Invalid JSON error response",
			resp: &http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(strings.NewReader("Internal Server Error")),
			},
			expectedErrorMessage: "status code 500, Internal Server Error",
		},
		{
			name: "Read error from body",
			resp: &http.Response{
				StatusCode: 400,
				Body:       &errorReadCloser{},
			},
			expectedErrorMessage: "read error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewResponseError(tc.resp)
			assert.Equal(t, tc.expectedErrorMessage, err.Error())
		})
	}
}
