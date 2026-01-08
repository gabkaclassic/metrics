// Package api provides common HTTP API primitives used across handlers and services.
//
// The package defines a unified error type (APIError) that encapsulates
// HTTP status codes, client-facing error messages, and optional internal errors.
//
// APIError is used throughout the application to ensure consistent
// error handling, logging, and JSON serialization of error responses.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// APIError represents an internal application error.
//
// Used inside service and handler layers.
// Serialized to client as ErrorResponse.
//
// swagger:model APIError
type APIError struct {
	// HTTP status code
	// example: 400
	Code int `json:"-"`
	// Error message
	// example: invalid metric type
	Message string `json:"error"`
	Err     error  `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

func New(code int, message string, err error) *APIError {
	return &APIError{Code: code, Message: message, Err: err}
}

func NotAllowed() *APIError {
	return &APIError{Code: http.StatusMethodNotAllowed, Message: "Method is not allowed"}
}

func NotFound(message string) *APIError {
	return &APIError{Code: http.StatusNotFound, Message: message}
}

func BadRequest(message string) *APIError {
	return &APIError{Code: http.StatusBadRequest, Message: message}
}

func UnprocessibleEntity(message string) *APIError {
	return &APIError{Code: http.StatusUnprocessableEntity, Message: message}
}

func Internal(message string, err error) *APIError {
	return &APIError{Code: http.StatusInternalServerError, Message: message, Err: err}
}

func Forbidden(message string) *APIError {
	return &APIError{Code: http.StatusForbidden, Message: message}
}

func Unauthorized(message string) *APIError {
	return &APIError{Code: http.StatusUnauthorized, Message: message}
}

// RespondError writes an API error response to the client.
//
// On error, responds with JSON body:
//
//	{ "error": "<message>" }
//
// HTTP status code is taken from APIError.Code.
// Unknown errors are converted to 500 Internal Server Error.
func RespondError(w http.ResponseWriter, err error) {

	if err == nil {
		return
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		apiErr = Internal("Internal server error", err)
	}

	requestID := w.Header().Get("X-Request-ID")
	slog.Info("error request handling", slog.Any("error", apiErr.Err), slog.String("message", apiErr.Message), slog.String("id", requestID))
	w.WriteHeader(apiErr.Code)
	json.NewEncoder(w).Encode(apiErr)
}
