package jwt

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// AuthenticateRequest sends an authenticate request
func AuthenticateRequest(authnURL string, account string, username string, jwtToken string) (*http.Request, error) {
	var err error
	var req *http.Request

	var authenticateURL = createUrl(authnURL, account, username)

	log.Debug(log.CAKC046, authenticateURL)

	formattedJwt := fmt.Sprintf("jwt=%s", jwtToken)
	requestBody := strings.NewReader(formattedJwt)

	if req, err = http.NewRequest("POST", authenticateURL, requestBody); err != nil {
		return nil, log.RecordedError(log.CAKC023, err)
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Content-Length", strconv.Itoa(len(formattedJwt)))
	req.Header.Set("User-Agent", "k8s")

	return req, nil
}

func createUrl(authnURL string, account string, username string) string {
	if len(username) > 0 {
		return fmt.Sprintf("%s/%s/%s/authenticate", authnURL, account, url.QueryEscape(username))
	}
	return fmt.Sprintf("%s/%s/authenticate", authnURL, account)
}
