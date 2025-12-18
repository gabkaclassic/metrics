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
	handler interface {
		handle(*event)
	}
	fileHandler struct {
		file *os.File
		mu   sync.Mutex
	}
	urlHandler struct {
		client httpclient.HTTPClient
	}

	Auditor interface {
		AuditOne(*models.Metrics, int64, string)
		AuditMany([]models.Metrics, int64, string)
	}
	auditor struct {
		handlers []handler
	}

	event struct {
		Ts        int64    `json:"ts"`
		Metrics   []string `json:"metrics"`
		IPAddress string   `json:"ip_address"`
	}
)

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

func newURLHandler(url string) (*urlHandler, error) {

	client := httpclient.NewClient(
		httpclient.BaseURL(url),
	)

	return &urlHandler{
		client: client,
	}, nil
}

func newFileHandler(filePath string) (*fileHandler, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0660)

	if err != nil {
		slog.Error("Open file error", slog.String("error", err.Error()))
		return nil, err
	}

	return &fileHandler{
		file: file,
	}, nil
}

func (a *auditor) AuditOne(metric *models.Metrics, timestamp int64, ip string) {
	metrics := []models.Metrics{*metric}
	a.AuditMany(metrics, timestamp, ip)
}

func (a *auditor) AuditMany(metrics []models.Metrics, timestamp int64, ip string) {

	if len(a.handlers) == 0 {
		return
	}

	e := &event{
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

func (h fileHandler) handle(e *event) {
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

func (h urlHandler) handle(e *event) {
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

func getMetricsNames(metrics []models.Metrics) []string {
	result := make([]string, len(metrics))

	for ind, metric := range metrics {
		result[ind] = metric.ID
	}

	return result
}
