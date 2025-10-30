package api

import (
	"encoding/json"
	"log/slog"
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
	json.NewEncoder(w).Encode(map[string]string{"error": apiErr.Message})
}
