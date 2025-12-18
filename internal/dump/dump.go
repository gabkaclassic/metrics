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
	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/repository"
)

type Dumper struct {
	file       *os.File
	repository repository.MetricsRepository
}

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
		go func() { errChan <- d.repository.AddAll(ctx, &counters) }()
	} else {
		go func() { errChan <- nil }()
	}

	if len(gauges) > 0 {
		go func() { errChan <- d.repository.ResetAll(ctx, &gauges) }()
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

func (d *Dumper) Close() {
	d.file.Close()
}
