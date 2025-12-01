package common

import (
	"errors"
	"testing"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadSSLCert(t *testing.T) {
	testCases := []struct {
		name          string
		settings      map[string]string
		mockReadFile  func(string) ([]byte, error)
		expectedCert  []byte
		expectedError string
	}{
		{
			name: "Returns error when both SSL cert and cert path are empty",
			settings: map[string]string{
				"CONJUR_SSL_CERTIFICATE": "",
				"CONJUR_CERT_FILE":       "",
			},
			mockReadFile:  nil,
			expectedCert:  nil,
			expectedError: log.CAKC007,
		},
		{
			name: "Returns SSL certificate when present",
			settings: map[string]string{
				"CONJUR_SSL_CERTIFICATE": "test-certificate-data",
				"CONJUR_CERT_FILE":       "/path/to/cert",
			},
			mockReadFile:  nil,
			expectedCert:  []byte("test-certificate-data"),
			expectedError: "",
		},
		{
			name: "Reads from file when SSL cert is empty",
			settings: map[string]string{
				"CONJUR_SSL_CERTIFICATE": "",
				"CONJUR_CERT_FILE":       "/path/to/cert.pem",
			},
			mockReadFile: func(filename string) ([]byte, error) {
				return []byte("file-certificate-data"), nil
			},
			expectedCert:  []byte("file-certificate-data"),
			expectedError: "",
		},
		{
			name: "Returns error when file reading fails",
			settings: map[string]string{
				"CONJUR_SSL_CERTIFICATE": "",
				"CONJUR_CERT_FILE":       "/invalid/path",
			},
			mockReadFile: func(filename string) ([]byte, error) {
				return nil, errors.New("file read error")
			},
			expectedCert:  nil,
			expectedError: "file read error",
		},
		{
			name: "Prefers SSL certificate over file path when both are present",
			settings: map[string]string{
				"CONJUR_SSL_CERTIFICATE": "direct-certificate",
				"CONJUR_CERT_FILE":       "/path/to/cert",
			},
			mockReadFile:  nil,
			expectedCert:  []byte("direct-certificate"),
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cert, err := ReadSSLCert(tc.settings, tc.mockReadFile)

			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCert, cert)
			}
		})
	}
}
