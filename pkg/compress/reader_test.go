package compress

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewGzipReader(t *testing.T) {
	data := []byte("test gzip data")
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(data)
	assert.NoError(t, err)
	err = gw.Close()
	assert.NoError(t, err)

	tests := []struct {
		name        string
		input       io.ReadCloser
		expectError bool
	}{
		{
			name:        "valid gzip reader",
			input:       io.NopCloser(&buf),
			expectError: false,
		},
		{
			name:        "invalid gzip data",
			input:       io.NopCloser(bytes.NewReader([]byte("bad data"))),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr, err := NewGzipReader(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cr)
			}
		})
	}
}

func TestCompressReader_Read(t *testing.T) {
	tests := []struct {
		name        string
		mockRead    func(m *mockreader)
		expectData  []byte
		expectError bool
	}{
		{
			name: "successful read",
			mockRead: func(m *mockreader) {
				m.EXPECT().Read(mock.Anything).
					RunAndReturn(func(p []byte) (int, error) {
						copy(p, []byte("hello"))
						return 5, nil
					})
			},
			expectData:  []byte("hello"),
			expectError: false,
		},
		{
			name: "read returns error",
			mockRead: func(m *mockreader) {
				m.EXPECT().Read(mock.Anything).
					Return(0, errors.New("read error"))
			},
			expectData:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockR := newMockreader(t)
			tt.mockRead(mockR)

			cr := &CompressReader{
				ReadCloser: io.NopCloser(bytes.NewReader(nil)),
				reader:     mockR,
			}

			buf := make([]byte, 5)
			n, err := cr.Read(buf)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, 0, n)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectData), n)
				assert.Equal(t, tt.expectData, buf)
			}
		})
	}
}

func TestCompressReader_Close(t *testing.T) {
	tests := []struct {
		name        string
		mockClose   func(m *mockreader)
		expectError bool
	}{
		{
			name: "successful close",
			mockClose: func(m *mockreader) {
				m.EXPECT().Close().Return(nil)
			},
			expectError: false,
		},
		{
			name: "close returns error",
			mockClose: func(m *mockreader) {
				m.EXPECT().Close().Return(errors.New("close error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockR := newMockreader(t)
			tt.mockClose(mockR)

			cr := &CompressReader{
				ReadCloser: io.NopCloser(bytes.NewReader(nil)),
				reader:     mockR,
			}

			err := cr.Close()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCompressReader_Integration(t *testing.T) {
	data := []byte("integration test data")
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(data)
	assert.NoError(t, err)
	err = gw.Close()
	assert.NoError(t, err)

	cr, err := NewGzipReader(io.NopCloser(&buf))
	assert.NoError(t, err)

	readData, err := io.ReadAll(cr)
	assert.NoError(t, err)
	assert.Equal(t, data, readData)

	err = cr.Close()
	assert.NoError(t, err)
}
