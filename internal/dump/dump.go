package dump

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gabkaclassic/metrics/internal/config"
	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/storage"
)

type Dumper struct {
	filePath string
	storage  *storage.MemStorage
}

func NewDumper(filePath string, storage *storage.MemStorage) (*Dumper, error) {

	if storage == nil {
		return nil, errors.New("create dumper error: storage can't be nil")
	}

	return &Dumper{
		filePath: filePath,
		storage:  storage,
	}, nil
}

func (d *Dumper) Dump() error {

	dir := filepath.Dir(d.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("Failed to create directories", slog.String("error", err.Error()), slog.String("path", dir))
		return err
	}

	data := make([]models.Metrics, 0)
	for _, metric := range d.storage.Metrics {
		data = append(data, metric)
	}

	marshalledData, err := json.Marshal(data)

	if err != nil {
		slog.Error("Marshall data error", slog.String("error", err.Error()))
		return err
	}

	file, err := os.OpenFile(d.filePath, os.O_WRONLY|os.O_CREATE, 0660)

	if err != nil {
		slog.Error("Open file error", slog.String("error", err.Error()), slog.String("path", d.filePath))
		return err
	}
	defer file.Close()

	_, err = file.Write(marshalledData)

	if err != nil {
		slog.Error("Write data error", slog.String("error", err.Error()), slog.String("path", d.filePath))
		return err
	}

	return nil
}

func (d *Dumper) Read() error {

	file, err := os.OpenFile(d.filePath, os.O_RDONLY|os.O_CREATE, 0660)

	if err != nil {
		slog.Info("Open file error", slog.String("error", err.Error()), slog.String("path", d.filePath))
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)

	if err != nil {
		slog.Info("Read data error", slog.String("error", err.Error()), slog.String("path", d.filePath))
		return err
	}

	if len(data) == 0 {
		slog.Info("Dump file is empty, nothing to restore", slog.String("path", d.filePath))
		return nil
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		slog.Error("Unmarshal data error", slog.String("error", err.Error()), slog.String("path", d.filePath))
		return err
	}

	for _, metric := range metrics {
		d.storage.Metrics[metric.ID] = metric
	}

	slog.Info("Dump restored successfully", slog.String("path", d.filePath))
	return nil
}

func (dumper *Dumper) StartDumper(ctx context.Context, cfg config.Dump) {
	ticker := time.NewTicker(cfg.StoreInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := dumper.Dump(); err != nil {
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
