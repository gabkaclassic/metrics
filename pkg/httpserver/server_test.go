package httpserver

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		wantAddr string
		wantMux  *http.ServeMux
	}{
		{
			name:     "default options",
			options:  nil,
			wantAddr: defaultAddress,
			wantMux:  http.NewServeMux(),
		},
		{
			name:     "custom address",
			options:  []Option{Address("127.0.0.1:9000")},
			wantAddr: "127.0.0.1:9000",
			wantMux:  http.NewServeMux(),
		},
		{
			name:     "custom handler",
			options:  []Option{Handler(http.NewServeMux())},
			wantAddr: defaultAddress,
			wantMux:  http.NewServeMux(),
		},
		{
			name: "custom address and handler",
			options: []Option{
				Address("127.0.0.1:9001"),
				Handler(http.NewServeMux()),
			},
			wantAddr: "127.0.0.1:9001",
			wantMux:  http.NewServeMux(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New(tt.options...)

			assert.Equal(t, tt.wantAddr, server.address)
			assert.IsType(t, tt.wantMux, server.GetHandler())
		})
	}
}
