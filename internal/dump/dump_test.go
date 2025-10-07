package dump

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewDumper(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		storage  *storage.MemStorage
		wantErr  error
		wantNil  bool
	}{
		{
			name:     "valid storage",
			filePath: "/tmp/test.json",
			storage:  storage.NewMemStorage(),
			wantErr:  nil,
			wantNil:  false,
		},
		{
			name:     "nil storage",
			filePath: "/tmp/test.json",
			storage:  nil,
			wantErr:  errors.New("create dumper error: storage can't be nil"),
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDumper(tt.filePath, tt.storage)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
				assert.Nil(t, d)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, d)
				assert.Equal(t, tt.filePath, d.filePath)
				assert.Equal(t, tt.storage, d.storage)
			}
		})
	}
}

func TestDumper_Dump(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		setup      func() *Dumper
		expectFile bool
		expectErr  bool
	}{
		{
			name: "success dump",
			setup: func() *Dumper {
				strg := storage.NewMemStorage()
				strg.Metrics["test"] = models.Metrics{
					ID:    "test",
					MType: "counter",
					Delta: int64Ptr(5),
				}
				file := filepath.Join(tmpDir, "metrics", "json")
				d, _ := NewDumper(file, strg)
				return d
			},
			expectFile: true,
			expectErr:  false,
		},
		{
			name: "invalid_path_error",
			setup: func() *Dumper {
				strg := storage.NewMemStorage()
				d, _ := NewDumper("/invalidpath", strg)
				return d
			},
			expectFile: false,
			expectErr:  true,
		},
		{
			name: "empty storage dump",
			setup: func() *Dumper {
				strg := storage.NewMemStorage()
				file := filepath.Join(tmpDir, "empty", "json")
				d, _ := NewDumper(file, strg)
				return d
			},
			expectFile: true,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()
			err := d.Dump()

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.FileExists(t, d.filePath)

			data, err := os.ReadFile(d.filePath)
			assert.NoError(t, err)

			var got []models.Metrics
			err = json.Unmarshal(data, &got)
			assert.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestDumper_Read(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setupFile func(path string)
		expectErr bool
		expectLen int
	}{
		{
			name: "valid dump file",
			setupFile: func(path string) {
				metrics := []models.Metrics{
					{ID: "test1", MType: "counter", Delta: int64Ptr(10)},
					{ID: "test2", MType: "gauge", Value: float64Ptr(3.14)},
				}
				data, _ := json.Marshal(metrics)
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, data, 0660)
			},
			expectErr: false,
			expectLen: 2,
		},
		{
			name: "empty file",
			setupFile: func(path string) {
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, []byte{}, 0660)
			},
			expectErr: false,
			expectLen: 0,
		},
		{
			name: "invalid JSON",
			setupFile: func(path string) {
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, []byte("{invalid json"), 0660)
			},
			expectErr: true,
			expectLen: 0,
		},
		{
			name: "file does not exist",
			setupFile: func(path string) {
			},
			expectErr: false,
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := filepath.Join(tmpDir, "dump", tt.name+".json")
			if tt.setupFile != nil {
				tt.setupFile(file)
			}

			strg := storage.NewMemStorage()
			d, _ := NewDumper(file, strg)

			err := d.Read()

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectLen, len(strg.Metrics))
		})
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}
