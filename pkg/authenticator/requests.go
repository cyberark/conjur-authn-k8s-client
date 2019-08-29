package authenticator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	log "github.com/cyberark/conjur-authn-k8s-client/pkg/logging"
)

// LoginRequest sends a login request
func LoginRequest(authnURL string, conjurVersion string, csrBytes []byte) (*http.Request, error) {
	var authenticateURL string

	if conjurVersion == "4" {
		authenticateURL = fmt.Sprintf("%s/users/login", authnURL)
	} else if conjurVersion == "5" {
		authenticateURL = fmt.Sprintf("%s/inject_client_cert", authnURL)
	}

	log.InfoLogger.Printf(log.CAKC011I, authenticateURL)

	req, err := http.NewRequest("POST", authenticateURL, bytes.NewBuffer(csrBytes))
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC058E, err.Error())
	}
	req.Header.Set("Content-Type", "text/plain")

	return req, nil
}

// AuthenticateRequest sends an authenticate request
func AuthenticateRequest(authnURL string, conjurVersion string, account string, username string) (*http.Request, error) {
	var authenticateURL string
	var err error
	var req *http.Request

	if conjurVersion == "4" {
		authenticateURL = fmt.Sprintf("%s/users/%s/authenticate", authnURL, url.QueryEscape(username))
	} else if conjurVersion == "5" {
		authenticateURL = fmt.Sprintf("%s/%s/%s/authenticate", authnURL, account, url.QueryEscape(username))
	}

	log.InfoLogger.Printf(log.CAKC012I, authenticateURL)

	if req, err = http.NewRequest("POST", authenticateURL, nil); err != nil {
		return nil, log.PrintAndReturnError(log.CAKC057E, err.Error())
	}

	req.Header.Set("Content-Type", "text/plain")

	return req, nil
}

func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, log.PrintAndReturnError(log.CAKC056E, err.Error())
	}

	return responseBytes, err
}

// DataResponse checks the HTTP status of the response. If it's less than
// 300, it returns the response body as a byte array. Otherwise it returns
// a NewError.
func DataResponse(resp *http.Response) ([]byte, error) {
	if resp.StatusCode < 300 {
		return readBody(resp)
	}
	return nil, NewError(resp)
}

// EmptyResponse checks the HTTP status of the response. If it's less than
// 300, it returns without an error. Otherwise it returns
// a NewError.
func EmptyResponse(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	return NewError(resp)
}
