package compress

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGzipWriter(t *testing.T) {
	tests := []struct {
		name        string
		responseW   http.ResponseWriter
		expectError bool
	}{
		{
			name:        "valid writer",
			responseW:   httptest.NewRecorder(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cw, err := NewGzipWriter(tt.responseW)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cw)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cw)
			}
		})
	}
}

func TestCompressWriter_Write(t *testing.T) {
	tests := []struct {
		name        string
		writeData   []byte
		mockWrite   func(m *mockwriter)
		expectError bool
	}{
		{
			name:      "successful write",
			writeData: []byte("hello"),
			mockWrite: func(m *mockwriter) {
				m.EXPECT().Write([]byte("hello")).Return(5, nil)
			},
			expectError: false,
		},
		{
			name:      "write returns error",
			writeData: []byte("fail"),
			mockWrite: func(m *mockwriter) {
				m.EXPECT().Write([]byte("fail")).Return(0, errors.New("write error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockW := newMockwriter(t)
			tt.mockWrite(mockW)

			cw := &CompressWriter{
				ResponseWriter: httptest.NewRecorder(),
				writer:         mockW,
			}

			n, err := cw.Write(tt.writeData)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, 0, n)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.writeData), n)
			}
		})
	}
}

func TestCompressWriter_Close(t *testing.T) {
	tests := []struct {
		name        string
		mockClose   func(m *mockwriter)
		expectError bool
	}{
		{
			name: "successful close",
			mockClose: func(m *mockwriter) {
				m.EXPECT().Close().Return(nil)
			},
			expectError: false,
		},
		{
			name: "close returns error",
			mockClose: func(m *mockwriter) {
				m.EXPECT().Close().Return(errors.New("close error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockW := newMockwriter(t)
			tt.mockClose(mockW)

			cw := &CompressWriter{
				ResponseWriter: httptest.NewRecorder(),
				writer:         mockW,
			}

			err := cw.Close()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCompressWriter_Integration(t *testing.T) {
	rec := httptest.NewRecorder()
	cw, err := NewGzipWriter(rec)
	assert.NoError(t, err)

	data := []byte("test data")
	_, err = cw.Write(data)
	assert.NoError(t, err)

	err = cw.Close()
	assert.NoError(t, err)

	r := rec.Result()
	gr, err := gzip.NewReader(r.Body)
	assert.NoError(t, err)
	defer gr.Close()
	defer r.Body.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gr)
	assert.NoError(t, err)
	assert.Equal(t, string(data), buf.String())
}
