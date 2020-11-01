package utils

import (
	"io/ioutil"
	"net/http"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// ReadResponseBody returns the response body
func ReadResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, log.RecordedError(log.CAKC022, err)
	}

	return responseBytes, err
}

// validateResponse checks the HTTP status of the response. If it's less than
// 300, it returns the response body as a byte array. Otherwise it returns
// a NewResponseError.
func ValidateResponse(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	return NewResponseError(resp)
}
