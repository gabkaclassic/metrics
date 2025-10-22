package dump

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestNewDumper(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		repository repository.MetricsRepository
		wantErr    error
		wantNil    bool
	}{
		{
			name:       "valid repository",
			filePath:   "/tmp/test.json",
			repository: repository.NewMockMetricsRepository(t),
			wantErr:    nil,
			wantNil:    false,
		},
		{
			name:       "nil repository",
			filePath:   "/tmp/test.json",
			repository: nil,
			wantErr:    errors.New("create dumper error: repository can't be nil"),
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := tt.repository
			d, err := NewDumper(tt.filePath, repository)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
				assert.Nil(t, d)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, d)
				assert.Equal(t, tt.repository, d.repository)
			}
		})
	}
}

func TestDumper_Dump(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		filename   string
		mockSetup  func(*repository.MockMetricsRepository)
		expectFile bool
		expectErr  bool
	}{
		{
			name:     "success dump",
			filename: "success_dump.json",
			mockSetup: func(m *repository.MockMetricsRepository) {
				metrics := []models.Metrics{
					{ID: "test", MType: "counter", Delta: int64Ptr(5)},
				}
				m.EXPECT().GetAllMetrics().Return(&metrics, nil)
			},
			expectFile: true,
			expectErr:  false,
		},
		{
			name:     "empty storage dump",
			filename: "empty_storage_dump.json",
			mockSetup: func(m *repository.MockMetricsRepository) {
				metrics := []models.Metrics{}
				m.EXPECT().GetAllMetrics().Return(&metrics, nil)
			},
			expectFile: true,
			expectErr:  false,
		},
		{
			name:     "repository error",
			filename: "error_dump.json",
			mockSetup: func(m *repository.MockMetricsRepository) {
				m.EXPECT().GetAllMetrics().Return(nil, errors.New("get metrics error"))
			},
			expectFile: false,
			expectErr:  true,
		},
		{
			name:     "nil metrics",
			filename: "nil_metrics.json",
			mockSetup: func(m *repository.MockMetricsRepository) {
				m.EXPECT().GetAllMetrics().Return(nil, nil)
			},
			expectFile: true,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockMetricsRepository(t)
			tt.mockSetup(mockRepo)

			filePath := filepath.Join(tmpDir, tt.filename)
			d, err := NewDumper(filePath, mockRepo)
			assert.NoError(t, err)

			err = d.Dump()

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.expectFile {
				assert.FileExists(t, filePath)

				data, err := os.ReadFile(filePath)
				assert.NoError(t, err)

				if tt.name == "nil metrics" {
					assert.Equal(t, "null", string(data))
				} else {
					var got []models.Metrics
					err = json.Unmarshal(data, &got)
					assert.NoError(t, err)
					assert.NotNil(t, got)
				}
			}
		})
	}
}

func TestDumper_Read(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		filename  string
		setupFile func(path string)
		mockSetup func(*repository.MockMetricsRepository, []models.Metrics, []models.Metrics)
		expectErr bool
	}{
		{
			name:     "valid dump file with mixed metrics",
			filename: "valid_dump_file.json",
			setupFile: func(path string) {
				metrics := []models.Metrics{
					{ID: "test1", MType: models.Counter, Delta: int64Ptr(10)},
					{ID: "test2", MType: models.Gauge, Value: float64Ptr(3.14)},
				}
				data, _ := json.Marshal(metrics)
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, data, 0660)
			},
			mockSetup: func(m *repository.MockMetricsRepository, counters []models.Metrics, gauges []models.Metrics) {
				m.EXPECT().AddAll(&counters).Return(nil)
				m.EXPECT().ResetAll(&gauges).Return(nil)
			},
			expectErr: false,
		},
		{
			name:     "only counters",
			filename: "only_counters.json",
			setupFile: func(path string) {
				metrics := []models.Metrics{
					{ID: "c1", MType: models.Counter, Delta: int64Ptr(5)},
					{ID: "c2", MType: models.Counter, Delta: int64Ptr(3)},
				}
				data, _ := json.Marshal(metrics)
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, data, 0660)
			},
			mockSetup: func(m *repository.MockMetricsRepository, counters []models.Metrics, gauges []models.Metrics) {
				m.EXPECT().AddAll(&counters).Return(nil)
			},
			expectErr: false,
		},
		{
			name:     "only gauges",
			filename: "only_gauges.json",
			setupFile: func(path string) {
				metrics := []models.Metrics{
					{ID: "g1", MType: models.Gauge, Value: float64Ptr(1.1)},
					{ID: "g2", MType: models.Gauge, Value: float64Ptr(2.2)},
				}
				data, _ := json.Marshal(metrics)
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, data, 0660)
			},
			mockSetup: func(m *repository.MockMetricsRepository, counters []models.Metrics, gauges []models.Metrics) {
				m.EXPECT().ResetAll(&gauges).Return(nil)
			},
			expectErr: false,
		},
		{
			name:     "empty file",
			filename: "empty_file.json",
			setupFile: func(path string) {
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, []byte{}, 0660)
			},
			mockSetup: func(m *repository.MockMetricsRepository, counters []models.Metrics, gauges []models.Metrics) {
			},
			expectErr: false,
		},
		{
			name:     "invalid JSON",
			filename: "invalid_json.json",
			setupFile: func(path string) {
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, []byte("{invalid json"), 0660)
			},
			mockSetup: func(m *repository.MockMetricsRepository, counters []models.Metrics, gauges []models.Metrics) {
			},
			expectErr: true,
		},
		{
			name:     "counters repository error",
			filename: "counters_error.json",
			setupFile: func(path string) {
				metrics := []models.Metrics{
					{ID: "c1", MType: models.Counter, Delta: int64Ptr(10)},
					{ID: "g1", MType: models.Gauge, Value: float64Ptr(3.14)},
				}
				data, _ := json.Marshal(metrics)
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, data, 0660)
			},
			mockSetup: func(m *repository.MockMetricsRepository, counters []models.Metrics, gauges []models.Metrics) {
				m.EXPECT().AddAll(&counters).Return(errors.New("counter error"))
				m.EXPECT().ResetAll(&gauges).Return(nil)
			},
			expectErr: true,
		},
		{
			name:     "gauges repository error",
			filename: "gauges_error.json",
			setupFile: func(path string) {
				metrics := []models.Metrics{
					{ID: "c1", MType: models.Counter, Delta: int64Ptr(10)},
					{ID: "g1", MType: models.Gauge, Value: float64Ptr(3.14)},
				}
				data, _ := json.Marshal(metrics)
				os.MkdirAll(filepath.Dir(path), 0755)
				os.WriteFile(path, data, 0660)
			},
			mockSetup: func(m *repository.MockMetricsRepository, counters []models.Metrics, gauges []models.Metrics) {
				m.EXPECT().AddAll(&counters).Return(nil)
				m.EXPECT().ResetAll(&gauges).Return(errors.New("gauge error"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.filename)
			if tt.setupFile != nil {
				tt.setupFile(filePath)
			}

			mockRepo := repository.NewMockMetricsRepository(t)

			var expectedCounters, expectedGauges []models.Metrics
			if data, err := os.ReadFile(filePath); err == nil && len(data) > 0 {
				var metrics []models.Metrics
				if err := json.Unmarshal(data, &metrics); err == nil {
					for _, metric := range metrics {
						switch metric.MType {
						case models.Counter:
							expectedCounters = append(expectedCounters, metric)
						case models.Gauge:
							expectedGauges = append(expectedGauges, metric)
						}
					}
				}
			}

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, expectedCounters, expectedGauges)
			}

			d, err := NewDumper(filePath, mockRepo)
			assert.NoError(t, err)

			err = d.Read()

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}
