// Package compress provides transparent compression/decompression wrappers for HTTP.
//
// The package implements gzip compression and decompression that integrates
// seamlessly with Go's standard HTTP and I/O interfaces. It provides:
//   - CompressReader: Decompresses gzipped HTTP request bodies
//   - CompressWriter: Compresses HTTP responses with gzip
//
// Both types implement standard io.Reader/io.Writer interfaces and can be used
// as drop-in replacements in HTTP middleware chains.
package compress
