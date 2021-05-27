package grpc

import (
	"io"

	"github.com/andybalholm/brotli"
	"google.golang.org/grpc/encoding"
)

func init() {
	registerBrotliCompressor()
}

func registerBrotliCompressor() {
	encoding.RegisterCompressor(&BrotliCompressor{})
}

const (
	brCompressionLevel = 3
)

//BrotliCompressor compressor that uses the Brotli algorithm
type BrotliCompressor struct{}

//Name is the name of the compression codec and is used to set the content
// coding header.  The result must be static; the result cannot change
// between calls.
func (bc *BrotliCompressor) Name() string {
	return "brotli"
}

//Compress writes the data written to wc to w after compressing it.  If an
// error occurs while initializing the compressor, that error is returned
// instead.
func (bc *BrotliCompressor) Compress(w io.Writer) (io.WriteCloser, error) {
	return brotli.NewWriterLevel(w, brCompressionLevel), nil
}

//Decompress reads data from r, decompresses it, and provides the
// uncompressed data via the returned io.Reader.  If an error occurs while
// initializing the decompressor, that error is returned instead.
func (bc *BrotliCompressor) Decompress(r io.Reader) (io.Reader, error) {
	return brotli.NewReader(r), nil
}
