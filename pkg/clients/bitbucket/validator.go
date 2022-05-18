package bitbucket

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/carlmjohnson/requests"
)

type BitbucketError struct {
	Errors []struct {
		Context       string `json:"context,omitempty"`
		Message       string `json:"message,omitempty"`
		ExceptionName string `json:"exceptionName,omitempty"`
	} `json:"errors,omitempty"`
}

type StatusError struct {
	Code  int
	Inner error
}

func (e StatusError) Error() string {
	if e.Inner != nil {
		return e.Inner.Error()
	}
	return fmt.Sprintf("unexpected status: %d:", e.Code)
}

func (e StatusError) Unwrap() error {
	return e.Inner
}

// ErrorJSON validates the response has an acceptable status
// code and if it's bad, attempts to marshal the JSON
// into the error object provided.
func ErrorHandler(acceptStatuses ...int) requests.ResponseHandler {
	return func(res *http.Response) error {
		for _, code := range acceptStatuses {
			if res.StatusCode == code {
				return nil
			}
		}

		if res.Body == nil {
			return StatusError{Code: res.StatusCode}
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return StatusError{Code: res.StatusCode, Inner: err}
		}

		var ex BitbucketError
		if err = json.Unmarshal(data, &ex); err != nil {
			return StatusError{Code: res.StatusCode, Inner: err}
		}

		if len(ex.Errors) == 0 {
			return StatusError{Code: res.StatusCode}
		}

		return StatusError{Code: res.StatusCode, Inner: fmt.Errorf(ex.Errors[0].Message)}
	}
}
