package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRespondError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantBody string
	}{
		{
			name:     "APIError",
			err:      &APIError{Code: 400, Message: "bad request"},
			wantCode: 400,
			wantBody: `{"error":"bad request"}` + "\n",
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"Internal server error"}` + "\n",
		},
		{
			name:     "nil error",
			err:      nil,
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"Internal server error"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			RespondError(rr, tt.err)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, rr.Body.String())
		})
	}
}
