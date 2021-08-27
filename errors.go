package solus

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// HTTPError represents errors occurred when some action failed due to some problem
// with request.
type HTTPError struct {
	Method   string
	Path     string
	HTTPCode int    `json:"http_code"`
	Message  string `json:"message"`
}

func (e HTTPError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("HTTP %s %s returns %d status code: %s", e.Method, e.Path, e.HTTPCode, e.Message)
	}
	return fmt.Sprintf("HTTP %s %s returns %d status code", e.Method, e.Path, e.HTTPCode)
}

func newHTTPError(method, path string, httpCode int, body []byte) error {
	e := HTTPError{
		Method:   method,
		Path:     path,
		HTTPCode: httpCode,
	}

	if err := json.Unmarshal(body, &e); err != nil {
		e.Message = string(body)
		return e
	}

	return e
}

// IsNotFound returns true if specified error is produced 'cause requested resource
// is not found.
func IsNotFound(err error) bool {
	var httpErr HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}

	return httpErr.HTTPCode == http.StatusNotFound
}
