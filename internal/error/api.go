package api

import (
	"net/http"
)

type ApiError struct {
	Code    int
	Message string
	Err     error
}

func (e *ApiError) Error() string {
	return e.Message
}

func New(code int, message string, err error) *ApiError {
	return &ApiError{Code: code, Message: message, Err: err}
}

func NotAllowed() *ApiError {
	return &ApiError{Code: http.StatusMethodNotAllowed, Message: "Method is not allowed"}
}

func NotFound(message string) *ApiError {
	return &ApiError{Code: http.StatusNotFound, Message: message}
}

func BadRequest(message string) *ApiError {
	return &ApiError{Code: http.StatusBadRequest, Message: message}
}

func Internal(message string, err error) *ApiError {
	return &ApiError{Code: http.StatusInternalServerError, Message: message, Err: err}
}

func Forbidden(message string) *ApiError {
	return &ApiError{Code: http.StatusForbidden, Message: message}
}

func Unauthorized(message string) *ApiError {
	return &ApiError{Code: http.StatusUnauthorized, Message: message}
}

func RespondError(w http.ResponseWriter, err error) {
	apiErr, ok := err.(*ApiError)
	if !ok {
		apiErr = Internal("Internal server error", err)
	}

	w.WriteHeader(apiErr.Code)
	w.Write([]byte(apiErr.Message))
}
