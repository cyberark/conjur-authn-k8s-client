package authenticator

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"strings"
)

type AuthenticatorError struct {
	Code      int
	Message   string
	Details   *AuthenticatorErrorDetails  `json:"error"`
}

type AuthenticatorErrorDetails struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func NewAuthenticatorError(resp *http.Response) error {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	autherr := AuthenticatorError{}
	autherr.Code = resp.StatusCode
	err = json.Unmarshal(body, &autherr)
	if err != nil {
		autherr.Message = strings.TrimSpace(string(body))
	}
	return &autherr
}

func (autherr *AuthenticatorError) Error() string {
	if autherr.Details != nil && autherr.Details.Message != "" {
		return autherr.Details.Message
	} else {
		return autherr.Message
	}
}

func (autherr *AuthenticatorError) CertExpired() bool {
	return autherr.Details != nil && autherr.Details.Code == "cert_expired"
}
