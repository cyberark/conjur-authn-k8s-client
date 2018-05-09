package main

import (
	"net/http"
	"fmt"
	"bytes"
	"net/url"
	"io/ioutil"
)

func LoginRequest(authnURL string, csrBytes []byte) (*http.Request, error) {
	authenticateURL := fmt.Sprintf("%s/users/login", authnURL)

	req, err := http.NewRequest("POST", authenticateURL, bytes.NewBuffer(csrBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")

	return req, nil
}

func AuthenticateRequest(authnURL, username string) (*http.Request, error) {
	authenticateURL := fmt.Sprintf("%s/users/%s/authenticate", authnURL, url.QueryEscape(username))

	req, err := http.NewRequest("POST", authenticateURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")

	return req, nil
}


func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBytes, err
}

//--------
// DataResponse checks the HTTP status of the response. If it's less than
// 300, it returns the response body as a byte array. Otherwise it returns
// a NewAuthenticatorError.
func DataResponse(resp *http.Response) ([]byte, error) {
	if resp.StatusCode < 300 {
		return readBody(resp)
	}
	return nil, NewAuthenticatorError(resp)
}

// EmptyResponse checks the HTTP status of the response. If it's less than
// 300, it returns without an error. Otherwise it returns
// a NewAuthenticatorError.
func EmptyResponse(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	return NewAuthenticatorError(resp)
}
