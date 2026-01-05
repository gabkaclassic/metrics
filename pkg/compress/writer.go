package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

// writer is the internal interface for compression writers.
// Combines io.Writer and io.Closer for flushing and cleanup.
type writer interface {
	io.Writer
	io.Closer
}

// CompressWriter implements http.ResponseWriter for compressing responses with gzip.
// Wraps an underlying gzip.Writer to provide transparent compression.
// Automatically sets Content-Encoding header to "gzip".
type CompressWriter struct {
	http.ResponseWriter
	writer writer
}

// NewGzipWriter creates a new gzip compression writer.
//
// w: The original http.ResponseWriter to wrap.
//
// Returns:
//   - *CompressWriter: Writer that transparently compresses data with gzip
//   - error: If gzip writer initialization fails
//
// Uses gzip.BestSpeed compression level for optimal performance.
// The writer buffers compressed data and flushes on Close().
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

// Write compresses data and writes it to the underlying response writer.
// Implements io.Writer interface.
// Returns number of uncompressed bytes written and any error encountered.
// Compression happens on-the-fly; data may be buffered until Close().
func (cw *CompressWriter) Write(b []byte) (int, error) {
	return cw.writer.Write(b)
}

// Close flushes any buffered compressed data and closes the gzip writer.
// Must be called to ensure all data is written to the response.
// Does not close the underlying http.ResponseWriter.
func (cw *CompressWriter) Close() error {
	return cw.writer.Close()
}
