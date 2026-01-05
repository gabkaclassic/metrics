package compress

import (
	"compress/gzip"
	"io"
)

// reader is the internal interface for decompression readers.
// Combines io.Reader and io.Closer for resource cleanup.
type reader interface {
	io.Reader
	io.Closer
}

// CompressReader implements io.ReadCloser for decompressing gzipped data.
// Wraps an underlying gzip.Reader to provide transparent decompression.
type CompressReader struct {
	io.ReadCloser
	reader reader
}

// NewGzipReader creates a new gzip decompression reader.
//
// r: The original io.ReadCloser containing gzipped data.
//
// Returns:
//   - *CompressReader: Reader that transparently decompresses gzipped data
//   - error: If gzip reader initialization fails (invalid gzip data)
//
// The returned reader decompresses data on-the-fly as it's read.
// Both the gzip reader and original reader are closed when Close() is called.
func NewGzipReader(r io.ReadCloser) (*CompressReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &CompressReader{
		ReadCloser: r,
		reader:     gr,
	}, nil
}

// Read reads decompressed data from the gzipped stream.
// Implements io.Reader interface.
// Returns number of bytes read and any error encountered.
// When EOF is reached, both Read and subsequent Close return no error.
func (cr *CompressReader) Read(b []byte) (int, error) {
	return cr.reader.Read(b)
}

// Close closes both the gzip reader and the underlying reader.
// Implements io.Closer interface.
// Should always be called to prevent resource leaks.
func (cr *CompressReader) Close() error {
	return cr.reader.Close()
}
