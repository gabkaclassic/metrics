package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

type writer interface {
	io.Writer
	io.Closer
}

type CompressWriter struct {
	http.ResponseWriter
	writer writer
}

func NewGzipWriter(w http.ResponseWriter) (*CompressWriter, error) {
	gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	return &CompressWriter{
		ResponseWriter: w,
		writer:         gz,
	}, nil
}

func (cw *CompressWriter) Write(b []byte) (int, error) {
	return cw.writer.Write(b)
}

func (cw *CompressWriter) Close() error {
	return cw.writer.Close()
}
