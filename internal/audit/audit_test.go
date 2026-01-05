package audit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gabkaclassic/metrics/internal/config"
	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/pkg/httpclient"
	"github.com/stretchr/testify/assert"
)

func TestNewAudior(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "audit.log")

	tests := []struct {
		name        string
		cfg         config.Audit
		expectErr   bool
		expectCount int
	}{
		{
			name:        "empty config creates auditor with no handlers",
			cfg:         config.Audit{},
			expectCount: 0,
		},
		{
			name: "file handler only",
			cfg: config.Audit{
				File: tmpFile,
			},
			expectCount: 1,
		},
		{
			name: "url handler only",
			cfg: config.Audit{
				URL: "http://example.com",
			},
			expectCount: 1,
		},
		{
			name: "file and url handlers",
			cfg: config.Audit{
				File: tmpFile,
				URL:  "http://example.com",
			},
			expectCount: 2,
		},
		{
			name: "invalid file path returns error",
			cfg: config.Audit{
				File: "/root/definitely/not/allowed/audit.log",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aud, err := NewAudior(tt.cfg)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, aud)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, aud)

			a := aud.(*auditor)
			assert.Len(t, a.handlers, tt.expectCount)
		})
	}
}

func TestNewURLHandler(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "valid URL creates handler",
			url:         "http://example.com",
			expectError: false,
		},
		{
			name:        "valid HTTPS URL creates handler",
			url:         "https://api.example.com",
			expectError: false,
		},
		{
			name:        "empty URL creates handler",
			url:         "",
			expectError: false,
		},
		{
			name:        "URL with path creates handler",
			url:         "http://localhost:8080/api/v1",
			expectError: false,
		},
		{
			name:        "invalid URL format creates handler (client may validate later)",
			url:         "://invalid",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := newURLHandler(tt.url)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, handler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
				assert.NotNil(t, handler.client)
			}
		})
	}
}

func TestNewFileHandler(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		filePath    string
		prepareFile bool
		expectError bool
	}{
		{
			name:        "existing file opens successfully",
			filePath:    filepath.Join(tempDir, "existing.txt"),
			prepareFile: true,
			expectError: false,
		},
		{
			name:        "non-existent file creates new file",
			filePath:    filepath.Join(tempDir, "newfile.txt"),
			prepareFile: false,
			expectError: false,
		},
		{
			name:        "file in non-existent directory returns error",
			filePath:    filepath.Join(tempDir, "nonexistent", "file.txt"),
			prepareFile: false,
			expectError: true,
		},
		{
			name:        "empty file path returns error",
			filePath:    "",
			prepareFile: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepareFile {
				f, err := os.Create(tt.filePath)
				assert.NoError(t, err)
				f.Close()

			}

			handler, err := newFileHandler(tt.filePath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, handler, fileHandler{})
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
				assert.NotNil(t, handler.file)

				if handler.file != nil {
					handler.file.Close()
				}
			}
		})
	}
}

func TestGetMetricsNames(t *testing.T) {
	tests := []struct {
		name     string
		metrics  []models.Metrics
		expected []string
	}{
		{
			name: "multiple metrics return correct names",
			metrics: []models.Metrics{
				{ID: "cpu_usage"},
				{ID: "memory_usage"},
				{ID: "disk_io"},
			},
			expected: []string{"cpu_usage", "memory_usage", "disk_io"},
		},
		{
			name:     "single metric returns single name",
			metrics:  []models.Metrics{{ID: "single_metric"}},
			expected: []string{"single_metric"},
		},
		{
			name:     "empty slice returns empty slice",
			metrics:  []models.Metrics{},
			expected: []string{},
		},
		{
			name: "metrics with empty names included",
			metrics: []models.Metrics{
				{ID: "metric1"},
				{ID: ""},
				{ID: "metric3"},
			},
			expected: []string{"metric1", "", "metric3"},
		},
		{
			name:     "nil slice returns nil",
			metrics:  nil,
			expected: []string{},
		},
		{
			name: "duplicate names all included",
			metrics: []models.Metrics{
				{ID: "same"},
				{ID: "same"},
				{ID: "different"},
			},
			expected: []string{"same", "same", "different"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMetricsNames(tt.metrics)
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestFileHandler_Handle(t *testing.T) {
	tests := []struct {
		name        string
		calls       int
		event       event
		expectLines int
	}{
		{
			name:  "single write",
			calls: 1,
			event: event{
				TS:        1,
				Metrics:   []string{"m1", "m2"},
				IPAddress: "127.0.0.1",
			},
			expectLines: 1,
		},
		{
			name:  "concurrent writes",
			calls: 10,
			event: event{
				TS:        2,
				Metrics:   []string{"m"},
				IPAddress: "ip",
			},
			expectLines: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := filepath.Join(t.TempDir(), "audit.log")

			f, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
			assert.NoError(t, err)
			defer f.Close()

			h, err := newFileHandler(tmp)
			assert.NoError(t, err)

			var wg sync.WaitGroup
			errCh := make(chan error, tt.calls)

			for i := 0; i < tt.calls; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					errCh <- h.handle(tt.event)
				}()
			}

			wg.Wait()
			close(errCh)

			for err := range errCh {
				assert.NoError(t, err)
			}

			data, err := os.ReadFile(tmp)
			assert.NoError(t, err)

			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			assert.Len(t, lines, tt.expectLines)

			for _, line := range lines {
				var got event
				err = json.Unmarshal([]byte(line), &got)
				assert.NoError(t, err)

				assert.Equal(t, tt.event.TS, got.TS)
				assert.Equal(t, tt.event.Metrics, got.Metrics)
				assert.Equal(t, tt.event.IPAddress, got.IPAddress)
			}
		})
	}
}

func TestURLHandler_Handle(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		expectError bool
	}{
		{
			name:        "success response",
			status:      http.StatusOK,
			expectError: false,
		},
		{
			name:        "not 200 response",
			status:      http.StatusInternalServerError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var received event

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				if len(body) > 0 {
					err = json.Unmarshal(body, &received)
					assert.NoError(t, err)
				}

				w.WriteHeader(tt.status)
				w.Write([]byte("ok"))
			}))
			defer srv.Close()

			client := httpclient.NewClient(
				httpclient.BaseURL(srv.URL),
				httpclient.MaxRetries(1),
			)

			h := &urlHandler{
				client: client,
			}

			e := event{
				TS:        123,
				Metrics:   []string{"m1", "m2"},
				IPAddress: "127.0.0.1",
			}

			err := h.handle(e)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, e.TS, received.TS)
			assert.Equal(t, e.Metrics, received.Metrics)
			assert.Equal(t, e.IPAddress, received.IPAddress)
		})
	}
}
