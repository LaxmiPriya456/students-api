// Package response has helpers to send JSON responses.
package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Response is the JSON shape for errors: {"status": "Error", "error": "..."}
type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// A const block groups related constants (like Java's static final fields).
const (
	StatusOK    = "OK"
	StatusError = "Error"
)

// WriteJson sends any value as JSON with the given status code.
//
// `interface{}` means "any type" (like Object in Java). Newer Go code
// writes the same thing as `any`.
func WriteJson(w http.ResponseWriter, status int, data interface{}) error {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // the status must be written BEFORE the body

	return json.NewEncoder(w).Encode(data)
}

// GeneralError wraps any error into a Response.
func GeneralError(err error) Response {
	return Response{
		Status: StatusError,
		Error:  err.Error(),
	}
}

// ValidationError turns validator errors into one readable message,
// e.g. "field Name is required field, field Age is required field".
func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	// ValidationErrors is a slice: one entry per field that failed.
	for _, err := range errs {
		// switch in Go needs no `break`: only the matching case runs.
		switch err.ActualTag() { // the rule that failed, e.g. "required"
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is required field", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is  invalid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "), // like String.join(", ", list)
	}

}
