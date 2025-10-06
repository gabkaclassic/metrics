package compress

import (
	"compress/gzip"
	"io"
)

type reader interface {
	io.Reader
	io.Closer
}

type CompressReader struct {
	io.ReadCloser
	reader reader
}

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

func (cr *CompressReader) Read(b []byte) (int, error) {
	return cr.reader.Read(b)
}

func (cr *CompressReader) Close() error {
	return cr.reader.Close()
}
