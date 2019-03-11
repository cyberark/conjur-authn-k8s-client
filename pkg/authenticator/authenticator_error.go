package authenticator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Error includes the error info for Authenticator-related errors
type Error struct {
	Code    int
	Message string
	Details *ErrorDetails `json:"error"`
}

// ErrorDetails includes JSON data on authenticator.Errors
type ErrorDetails struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// NewError creates a new instance of authenticator.Error
func NewError(resp *http.Response) error {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	autherr := Error{}
	autherr.Code = resp.StatusCode
	err = json.Unmarshal(body, &autherr)
	if err != nil {
		autherr.Message = strings.TrimSpace(string(body))
	}
	return &autherr
}

// Error returns the error message
func (autherr *Error) Error() string {
	msg := autherr.Message
	if autherr.Details != nil && autherr.Details.Message != "" {
		msg = autherr.Details.Message
	}

	return fmt.Sprintf("status code %v, %s", autherr.Code, msg)
}
