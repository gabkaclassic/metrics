// Package audit provides auditing capabilities for metrics operations.
//
// The audit package implements a flexible auditing system that can log metric
// operations to multiple destinations simultaneously (file, HTTP endpoint).
// It provides non-blocking, asynchronous audit logging to minimize performance
// impact on the main application flow.
//
// Audit events include:
//   - Timestamp of the operation
//   - List of metric IDs involved
//   - Source IP address of the request
//
// The system supports multiple concurrent handlers and ensures thread-safe
// operations where required (e.g., file writing).
package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/gabkaclassic/metrics/internal/config"
	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/pkg/httpclient"
)

type (
	// handler is the internal interface for audit event handlers.
	// Implementations define how audit events are processed and stored
	handler interface {
		// handle processes a single audit event.
		// Implementations should be thread-safe and return error.
		handle(event) error
	}
	// fileHandler implements handler interface for file-based audit logging.
	// Writes audit events as JSON lines to a specified file.
	fileHandler struct {
		file *os.File
		mu   *sync.Mutex
	}
	// urlHandler implements handler interface for HTTP-based audit logging.
	// Sends audit events as JSON to a remote HTTP endpoint.
	urlHandler struct {
		client httpclient.HTTPClient
	}

	// Auditor defines the public interface for audit operations.
	// Provides methods for auditing single and multiple metrics operations.
	Auditor interface {
		// AuditOne logs a single metric operation.
		// metric: The metric that was operated on
		// timestamp: Unix timestamp of the operation
		// ip: Source IP address of the request
		AuditOne(models.Metrics, int64, string)

		// AuditMany logs multiple metrics operations in a single event.
		// metrics: List of metrics that were operated on
		// timestamp: Unix timestamp of the operation
		// ip: Source IP address of the request
		AuditMany([]models.Metrics, int64, string)
	}
	// auditor implements the Auditor interface with multiple handler support.
	// Distributes audit events to all configured handlers concurrently.
	auditor struct {
		handlers []handler
	}

	// event represents a single audit event to be logged.
	// Serialized as JSON for both file and HTTP handlers.
	event struct {
		// Ts is the Unix timestamp of the audited operation.
		TS int64 `json:"ts"`

		// Metrics contains the IDs of all metrics involved in the operation.
		Metrics []string `json:"metrics"`

		// IPAddress is the source IP address of the request.
		IPAddress string `json:"ip_address"`
	}
)

// NewAudior creates a new Auditor instance based on configuration.
//
// cfg: Audit configuration specifying file path and/or URL endpoints.
//
// Returns:
//   - Auditor: Configured audit system ready for use
//   - error: If handler initialization fails (file cannot be opened, etc.)
//
// The auditor supports multiple simultaneous destinations:
//   - File logging: For local audit trails (JSON lines format)
//   - URL endpoint: For centralized audit collection
//
// If no handlers are configured, audit operations become no-ops.
func NewAudior(cfg config.Audit) (Auditor, error) {

	a := &auditor{}

	a.handlers = make([]handler, 0)

	if len(cfg.File) > 0 {
		fh, err := newFileHandler(cfg.File)

		if err != nil {
			return nil, fmt.Errorf("auditor creation error: file handler creation error: %w", err)
		}

		a.handlers = append(a.handlers, fh)
	}

	if len(cfg.URL) > 0 {
		uh, err := newURLHandler(cfg.URL)

		if err != nil {
			return nil, fmt.Errorf("auditor creation error: url handler creation error: %w", err)
		}

		a.handlers = append(a.handlers, uh)
	}

	return a, nil
}

// newURLHandler creates a URL handler for HTTP audit logging.
//
// url: HTTP endpoint URL for audit events (e.g., "https://audit.example.com/log")
//
// Returns:
//   - urlHandler: Configured HTTP audit handler
//   - error: Always nil in current implementation, reserved for future validation
func newURLHandler(url string) (urlHandler, error) {

	client := httpclient.NewClient(
		httpclient.BaseURL(url),
	)

	return urlHandler{
		client: client,
	}, nil
}

// newFileHandler creates a file handler for local audit logging.
//
// filePath: Path to the audit log file (created if doesn't exist)
//
// Returns:
//   - fileHandler: Configured file audit handler with open file handle
//   - error: If file cannot be opened or created
//
// File permissions: 0660 (rw-rw----)
// File is opened in read-write mode with create flag.
func newFileHandler(filePath string) (fileHandler, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0660)

	if err != nil {
		slog.Error("Open file error", slog.String("error", err.Error()))
		return fileHandler{}, err
	}

	return fileHandler{
		file: file,
		mu:   &sync.Mutex{},
	}, nil
}

// AuditOne logs a single metric operation to all configured handlers.
// Wraps the single metric in a slice and calls AuditMany.
// This is a convenience method for single metric operations.
func (a *auditor) AuditOne(metric models.Metrics, timestamp int64, ip string) {
	metrics := []models.Metrics{metric}
	a.AuditMany(metrics, timestamp, ip)
}

// AuditMany logs multiple metric operations using all configured audit handlers.
//
// The method is asynchronous (fire-and-forget):
//   - returns immediately
//   - handler execution happens in background goroutines
//
// Behavior:
//  1. Builds a single audit event from input data
//  2. Dispatches the event to all handlers concurrently
//  3. Collects and logs handler errors internally
//  4. Does not propagate errors to the caller
//
// If no handlers are configured, the method is a no-op.
func (a *auditor) AuditMany(metrics []models.Metrics, timestamp int64, ip string) {

	if len(a.handlers) == 0 {
		return
	}

	e := event{
		TS:        timestamp,
		Metrics:   getMetricsNames(metrics),
		IPAddress: ip,
	}

	go func() {
		var wg sync.WaitGroup

		for _, hndlr := range a.handlers {
			wg.Add(1)

			go func(h handler) {
				defer wg.Done()

				if err := h.handle(e); err != nil {
					slog.Error(
						"audit handler error",
						slog.Any("error", err),
						slog.String("handler", fmt.Sprintf("%T", h)),
					)
				}
			}(hndlr)
		}

		wg.Wait()
	}()
}

// handle writes an audit event to a file as a single JSON line.
//
// Guarantees:
//   - thread-safe write
//   - immediate fsync
//
// Returns:
//   - error if marshalling, writing or syncing fails
func (h fileHandler) handle(e event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	data = append(data, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, err = h.file.Write(data); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	if err = h.file.Sync(); err != nil {
		return fmt.Errorf("sync audit file: %w", err)
	}

	return nil
}

// handle sends an audit event to a remote HTTP endpoint.
//
// The event is sent as JSON via POST request.
// A non-200 HTTP response is treated as an error.
//
// Returns:
//   - error if request creation, sending, or response validation fails
func (h urlHandler) handle(e event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	resp, err := h.client.Post("", &httpclient.RequestOptions{
		Body: bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("send audit request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected audit response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read audit response body: %w", err)
	}

	slog.Debug(
		"URL audit completed successfully",
		slog.String("response", string(body)),
	)

	return nil
}

// getMetricsNames extracts metric IDs from a slice of metrics.
// Used to populate the Metrics field in audit events.
// Returns a slice of metric ID strings in the same order as input.
func getMetricsNames(metrics []models.Metrics) []string {
	result := make([]string, len(metrics))

	for ind, metric := range metrics {
		result[ind] = metric.ID
	}

	return result
}
