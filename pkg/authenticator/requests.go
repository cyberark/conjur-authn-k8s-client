package authenticator

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// LoginRequest sends a login request
func LoginRequest(authnURL string, conjurVersion string, csrBytes []byte, usernamePrefix string) (*http.Request, error) {
	var authenticateURL string

	if conjurVersion == "4" {
		authenticateURL = fmt.Sprintf("%s/users/login", authnURL)
	} else if conjurVersion == "5" {
		authenticateURL = fmt.Sprintf("%s/inject_client_cert", authnURL)
	}

	log.Debug(log.CAKC045, authenticateURL)

	req, err := http.NewRequest("POST", authenticateURL, bytes.NewBuffer(csrBytes))
	if err != nil {
		return nil, log.RecordedError(log.CAKC024, err)
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Host-Id-Prefix", usernamePrefix)

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

	log.Debug(log.CAKC046, authenticateURL)

	if req, err = http.NewRequest("POST", authenticateURL, nil); err != nil {
		return nil, log.RecordedError(log.CAKC023, err)
	}

	req.Header.Set("Content-Type", "text/plain")

	return req, nil
}
