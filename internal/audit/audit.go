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
		// Implementations should be thread-safe and handle errors gracefully.
		handle(event)
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
		Ts int64 `json:"ts"`

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

// AuditMany logs multiple metrics operations to all configured handlers.
// Operates asynchronously - returns immediately without waiting for handlers.
// If no handlers are configured, the operation is a no-op.
//
// The method:
//  1. Creates an audit event from the provided data
//  2. Distributes the event to all handlers in parallel goroutines
//  3. Returns immediately (fire-and-forget)
//  4. Handlers process events asynchronously with their own error handling
func (a *auditor) AuditMany(metrics []models.Metrics, timestamp int64, ip string) {

	if len(a.handlers) == 0 {
		return
	}

	e := event{
		Ts:        timestamp,
		Metrics:   getMetricsNames(metrics),
		IPAddress: ip,
	}

	var wg sync.WaitGroup

	for _, hndlr := range a.handlers {
		wg.Add(1)
		go func(h handler) {
			defer wg.Done()
			h.handle(e)
		}(hndlr)
	}

	go func() {
		wg.Wait()
	}()
}

// handle processes an audit event by writing it to a file.
// Implements thread-safe file writing with immediate sync to disk.
// Each event is written as a JSON line followed by newline.
// Errors are logged but not propagated to maintain operation continuity.
func (h fileHandler) handle(e event) {
	marshalledData, err := json.Marshal(e)

	if err != nil {
		slog.Error("Marshall data error", slog.String("error", err.Error()))
		return
	}

	marshalledData = append(marshalledData, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err = h.file.Write(marshalledData)
	if err != nil {
		slog.Error("Write to file error", slog.String("error", err.Error()))
		return
	}

	err = h.file.Sync()
	if err != nil {
		slog.Error("File sync error", slog.String("error", err.Error()))
		return
	}
}

// handle processes an audit event by sending it to a remote HTTP endpoint.
// Sends event as JSON in POST request body.
// Validates HTTP response status and logs errors.
// Errors are logged but not propagated to maintain operation continuity.
func (h urlHandler) handle(e event) {
	marshalledData, err := json.Marshal(e)

	if err != nil {
		slog.Error("Marshall data error", slog.String("error", err.Error()))
		return
	}

	body := bytes.NewReader(marshalledData)

	resp, err := h.client.Post("", &httpclient.RequestOptions{
		Body: body,
	})

	if err != nil {
		slog.Error("Audit URL handle error", slog.Any("error", err))
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Audit URL handle HTTP error", slog.Int("status", resp.StatusCode))
		return
	}

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		slog.Error("Audit URL handle error: read response body error", slog.Any("error", err))
		return
	}

	slog.Debug("URL audit completed successfully",
		slog.String("response", string(responseBody)),
	)
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
