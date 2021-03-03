package gcp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// AuthenticateRequest sends an authenticate request
func AuthenticateRequest(applianceURL string, account string, sessionToken []byte, base64Token bool) (*http.Request, error) {
	var authenticateURL string
	var err error
	var req *http.Request

	authenticateURL = fmt.Sprintf("%s/authn-gcp/%s/authenticate", applianceURL, account)
	log.Debug(log.CAKC046, authenticateURL)

	body := strings.NewReader(fmt.Sprintf("jwt=%s", string(sessionToken)))

	if req, err = http.NewRequest("POST", authenticateURL, body); err != nil {
		return nil, log.RecordedError(log.CAKC023, err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if base64Token {
		req.Header.Set("Accept-Encoding", "base64")
	}

	return req, nil
}

// MetadataRequest sends a request to the google metadata
// endpoint to get a service account bearer token
func MetadataRequest(account string, username string) (*http.Request, error) {
	var err error
	var req *http.Request

	metadataIdentityURL := "http://metadata/computeMetadata/v1/instance/service-accounts/default/identity"
	audience := url.QueryEscape(fmt.Sprintf("conjur/%s/%s", account, username))
	metadataURL := fmt.Sprintf("%s?audience=%s&format=full", metadataIdentityURL, audience)
	log.Debug(log.CAKC046, metadataURL)

	if req, err = http.NewRequest("GET", metadataURL, nil); err != nil {
		return nil, log.RecordedError(log.CAKC023, err)
	}

	req.Header.Set("Metadata-Flavor", "Google")

	return req, nil
}
