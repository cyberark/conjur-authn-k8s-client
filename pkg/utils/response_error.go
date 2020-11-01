package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Error includes the error info for response errors
type Error struct {
	Code    int
	Message string
	Details *ErrorDetails `json:"error"`
}

// ErrorDetails includes JSON data on Errors
type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewResponseError creates a new instance of Error
func NewResponseError(resp *http.Response) error {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	responseErr := Error{}
	responseErr.Code = resp.StatusCode
	err = json.Unmarshal(body, &responseErr)
	if err != nil {
		responseErr.Message = strings.TrimSpace(string(body))
	}
	return &responseErr
}

// Error returns the error message
func (responseErr *Error) Error() string {
	msg := responseErr.Message
	if responseErr.Details != nil && responseErr.Details.Message != "" {
		msg = responseErr.Details.Message
	}

	return fmt.Sprintf("status code %v, %s", responseErr.Code, msg)
}
