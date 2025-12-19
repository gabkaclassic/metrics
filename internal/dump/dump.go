// Package dump provides persistence layer for metrics data.
//
// The dumper implements periodic backup and restore functionality for metrics,
// ensuring data persistence across service restarts. Supports both synchronous
// (on-demand) and asynchronous (periodic) dumping strategies.
package dump

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gabkaclassic/metrics/internal/config"
	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/repository"
)

// Dumper handles metrics persistence to and from filesystem.
// Provides atomic write operations and concurrent-safe file access.
// Implements both backup (dump) and restore (read) functionality.
type Dumper struct {
	// file is the open file handle for dump operations.
	// All operations use the same file handle for consistency.
	file *os.File

	// repository provides access to current metrics data.
	// Used to retrieve metrics for dumping and restore dumped data.
	repository repository.MetricsRepository
}

// NewDumper creates a new dumper instance with file and repository access.
//
// filePath: Absolute or relative path to the dump file.
//
//	Parent directories will be created if they don't exist.
//
// repository: Metrics repository for data access operations.
//
// Returns:
//   - *Dumper: Initialized dumper ready for operations
//   - error: If repository is nil or file cannot be opened
//
// File permissions:
//   - Directories: 0755 (rwxr-xr-x)
//   - File: 0660 (rw-rw----)
//
// The file is opened in read-write mode with create flag, ensuring it exists.
func NewDumper(filePath string, repository repository.MetricsRepository) (*Dumper, error) {

	if repository == nil {
		return nil, errors.New("create dumper error: repository can't be nil")
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("Failed to create directories", slog.String("error", err.Error()), slog.String("path", dir))
		return nil, err
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0660)

	if err != nil {
		slog.Error("Open file error", slog.String("error", err.Error()))
		return nil, err
	}

	return &Dumper{
		file:       file,
		repository: repository,
	}, nil
}

// Dump saves all current metrics to the dump file in JSON format.
// Performs atomic write: truncates file, writes new data, ensures consistency.
//
// ctx: Context for cancellation and timeout of repository operations.
//
// Returns:
//   - error: If any step fails (retrieval, marshaling, or file operations)
//
// Operation sequence:
//  1. Retrieve all metrics from repository
//  2. Marshal to JSON
//  3. Seek to file beginning
//  4. Truncate file to 0 bytes
//  5. Write JSON data
//
// The operation ensures the file either contains complete new data or preserves
// previous data on failure (atomic write pattern).
func (d *Dumper) Dump(ctx context.Context) error {

	data, err := d.repository.GetAllMetrics(ctx)

	if err != nil {
		slog.Error("Get data error", slog.String("error", err.Error()))
		return err
	}

	marshalledData, err := json.Marshal(data)

	if err != nil {
		slog.Error("Marshall data error", slog.String("error", err.Error()))
		return err
	}

	_, err = d.file.Seek(0, 0)
	if err != nil {
		slog.Error("Seek file error", slog.String("error", err.Error()))
		return err
	}

	err = d.file.Truncate(0)
	if err != nil {
		slog.Error("Truncate file error", slog.String("error", err.Error()))
		return err
	}

	_, err = d.file.Write(marshalledData)

	if err != nil {
		slog.Error("Write data error", slog.String("error", err.Error()))
		return err
	}

	return nil
}

// Read restores metrics from dump file to the repository.
// Reads JSON data, separates counters and gauges, restores concurrently.
//
// Returns:
//   - error: If file read, unmarshal, or repository operations fail
//
// Restoration process:
//  1. Read entire file contents
//  2. Skip if file is empty (no previous dump)
//  3. Unmarshal JSON to metrics slice
//  4. Separate counters and gauges
//  5. Restore counters (AddAll) and gauges (ResetAll) concurrently
//  6. Log success or combined error
//
// Note: Uses background context since this is typically called at startup.
func (d *Dumper) Read() error {

	data, err := io.ReadAll(d.file)

	if err != nil {
		slog.Info("Read data error", slog.String("error", err.Error()))
		return err
	}

	if len(data) == 0 {
		slog.Info("Dump file is empty, nothing to restore")
		return nil
	}

	var metrics []models.Metrics
	var counters []models.Metrics
	var gauges []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		slog.Error("Unmarshal data error", slog.String("error", err.Error()))
		return err
	}

	for _, metric := range metrics {
		switch metric.MType {
		case models.Counter:
			counters = append(counters, metric)
		case models.Gauge:
			gauges = append(gauges, metric)
		}
	}

	errChan := make(chan error, 2)
	ctx := context.Background()
	if len(counters) > 0 {
		go func() { errChan <- d.repository.AddAll(ctx, counters) }()
	} else {
		go func() { errChan <- nil }()
	}

	if len(gauges) > 0 {
		go func() { errChan <- d.repository.ResetAll(ctx, gauges) }()
	} else {
		go func() { errChan <- nil }()
	}

	err1 := <-errChan
	err2 := <-errChan

	if err1 != nil || err2 != nil {
		return fmt.Errorf("save metrics error: counters: %v, gauges: %v", err1, err2)
	}

	slog.Info("Dump restored successfully")
	return nil
}

// StartDumper initiates periodic dumping of metrics based on configuration.
// Runs asynchronously until context cancellation.
//
// ctx: Context for graceful shutdown (cancellation stops the dumper)
// cfg: Dump configuration containing store interval
//
// The dumper:
//   - Runs immediately on first ticker interval
//   - Continues dumping at configured interval
//   - Logs errors but continues on dump failures
//   - Stops gracefully on context cancellation
//
// Typical store intervals: 30s, 1m, 5m depending on data volatility.
func (d *Dumper) StartDumper(ctx context.Context, cfg config.Dump) {
	ticker := time.NewTicker(cfg.StoreInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := d.Dump(ctx); err != nil {
				slog.Error("Dump error", slog.String("error", err.Error()))
			} else {
				slog.Info("Dump completed")
			}
		case <-ctx.Done():
			slog.Info("Dumper stopped")
			return
		}
	}
}

// Close releases the dump file handle.
// Should be called during application shutdown to ensure proper resource cleanup.
func (d *Dumper) Close() {
	d.file.Close()
}
