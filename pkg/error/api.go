package api

import (
	"net/http"
)

type APIError struct {
	Code    int
	Message string
	Err     error
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

func Internal(message string, err error) *APIError {
	return &APIError{Code: http.StatusInternalServerError, Message: message, Err: err}
}

func Forbidden(message string) *APIError {
	return &APIError{Code: http.StatusForbidden, Message: message}
}

func Unauthorized(message string) *APIError {
	return &APIError{Code: http.StatusUnauthorized, Message: message}
}

func RespondError(w http.ResponseWriter, err error) {
	apiErr, ok := err.(*APIError)
	if !ok {
		apiErr = Internal("Internal server error", err)
	}

	w.WriteHeader(apiErr.Code)
	w.Write([]byte(apiErr.Message))
}
