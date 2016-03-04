package bzip2

import (
	"errors"
	"io"
)
import bit "github.com/fwip/bzip2w/bit"

const (
	bzip2FileMagic  = 0x425a         // "BZ"
	bzip2BlockMagic = 0x314159265359 // BCD pi
	bzip2FinalMagic = 0x177245385090 // BCD sqrt(pi)
)

const (
	runA = 0
	runB = 1
)

// Writer implememnts io.WriteCloser
type Writer struct {
	w             *bit.Writer
	blockSize     byte // 1 - 9
	headerWritten bool
	blocks        []blockEncoder
	currentBlock  int
}

var _ io.Writer = &Writer{}

// NewWriter creates a new Wrtier that bzip2 compresses the input
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:         bit.NewWriter(w),
		blockSize: 9,
	}
}

// Write compresses bytes with bzip2 and then sends them to the underlying io.Writer
func (w *Writer) Write(b []byte) (n int, err error) {

	// Allocate a new blockEncoder if we need to
	if w.currentBlock == len(w.blocks) {
		blockSize := 1e5 * uint(w.blockSize)
		w.blocks = append(w.blocks, *newBlockEncoder(blockSize))
	}

	n, _ = w.blocks[w.currentBlock].Write(b)

	// If encoder is full, Write again with the leftovers
	if n < len(b) {
		w.currentBlock++
		n2, err := w.Write(b[n:])
		return n + n2, err
	}

	return n, nil
}

// SetBlockSize takes an int from 1-9, and sets the block size used by bzip2 to
// 100KB-900KB, respectively. This should only be called before calling
// Write(), and will throw an error otherwise.
func (w *Writer) SetBlockSize(n int) error {
	if w.headerWritten {
		return errors.New("SetBlockSize() called after writing has begun")
	}
	if n < 1 || n > 9 {
		return errors.New("invalid block size")
	}
	w.blockSize = byte(n)
	return nil
}

// Close will finalize the writer and block until all data has been written
// out. Once Close has been called, further calls to Write will do nothing, and
// return an error
func (w *Writer) Close() error { return nil }

func (w *Writer) writeMagicNumber() {

}

func rle(in []byte) (out, leftovers []byte) {
	var count byte
	//var lastByte byte = 0
	for i := 1; i < len(in); i++ {
		count++
		if in[i] != in[i-1] || count == 255 {
			//for j := 0; j < 4 && j < int(count); j++ {
			//out = append(out, in[i-1])
			//}
			if count >= 4 {
				out = append(out, in[i-4:i]...)
				out = append(out, count-4)
			} else {
				out = append(out, in[i-int(count):i]...)
			}
			count = 0
		}
	}

	leftovers = append(leftovers, in[len(in)-int(count):]...)

	return out, leftovers
}
