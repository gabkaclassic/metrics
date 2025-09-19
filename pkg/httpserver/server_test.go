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
		wantType interface{}
	}{
		{
			name:     "default options",
			options:  nil,
			wantAddr: defaultAddress,
			wantType: &http.ServeMux{},
		},
		{
			name:     "custom address",
			options:  []Option{Address("127.0.0.1:9000")},
			wantAddr: "127.0.0.1:9000",
			wantType: &http.ServeMux{},
		},
		{
			name: "custom handler",
			options: []Option{
				Handler(func() *http.Handler {
					h := http.NewServeMux()
					var hh http.Handler = h
					return &hh
				}()),
			},
			wantAddr: defaultAddress,
			wantType: &http.ServeMux{},
		},
		{
			name: "custom address and handler",
			options: []Option{
				Address("127.0.0.1:9001"),
				Handler(func() *http.Handler {
					h := http.NewServeMux()
					var hh http.Handler = h
					return &hh
				}()),
			},
			wantAddr: "127.0.0.1:9001",
			wantType: &http.ServeMux{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New(tt.options...)

			assert.Equal(t, tt.wantAddr, server.address)
			if server.handler != nil {
				assert.IsType(t, tt.wantType, *server.handler)
			} else {
				assert.Nil(t, server.handler)
			}
		})
	}
}
