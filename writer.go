package bzip2

import (
	"errors"
	"fmt"
	"io"
)
import bit "github.com/fwip/bzip2w/bit"

const (
	bzip2FileMagic = 0x425a // "BZ"
	//bzip2BlockMagic = 0x314159265359 // BCD pi
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
	sendTo        chan []byte
	blocks        []blockEncoder
	currentBlock  int
	closed        chan struct{}
}

var _ io.Writer = &Writer{}

// NewWriter creates a new Wrtier that bzip2 compresses the input
func NewWriter(w io.Writer) *Writer {
	writer := Writer{
		w:         bit.NewWriter(w),
		blockSize: 9,
	}

	writer.setUp()

	return &writer
}

// Write compresses bytes with bzip2 and then sends them to the underlying io.Writer
func (w *Writer) Write(b []byte) (n int, err error) {

	w.sendTo <- b
	return len(b), nil
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

func (w *Writer) setUp() {
	w.closed = make(chan struct{})
	size := int(w.blockSize) * 1e5
	w.sendTo = make(chan []byte)
	postRLE := make(chan []byte)
	go rlePipeline(w.sendTo, postRLE)
	outputChan := make(chan *blockEncoder)
	go chunker(size, postRLE, outputChan)
	go writePipeline(outputChan, w.w, w.closed)
}

// TODO: This probably holds onto all the memory and prevents GC
func rlePipeline(input <-chan []byte, output chan<- []byte) {
	var prev []byte
	for in := range input {
		fmt.Println("RLEing", in)
		send, leftovers := rle(append(prev, in...))
		output <- send
		prev = leftovers
	}
	output <- prev
	close(output)

}

func chunker(size int, input <-chan []byte, results chan *blockEncoder) {
	cache := make([]byte, 0, size)
	for in := range input {
		if len(in) == 0 {
			continue
		}
		fmt.Println("Chunking...")
		room := cap(cache) - len(cache)
		length := len(in)

		if length > room {
			cache = append(cache, in[:room]...)
			block := encodeAsync(cache)
			results <- block
			cache = cache[:0]
			copy(cache, in[room:])

		} else {
			cache = append(cache, in...)
		}
	}
	block := encodeAsync(cache)
	results <- block
	close(results)
}

// Returns a locked blockEncoder that asynchronously encodes
// Will unlock once it's finished.
func encodeAsync(in []byte) *blockEncoder {
	fmt.Println("Async encoding", in)
	b := blockEncoder{}
	b.input = in

	b.Add(1)
	go func() {
		b.encode()
		b.Done()
	}()
	return &b
}

func writePipeline(blocks chan *blockEncoder, w *bit.Writer, done chan struct{}) {

	// Write header

	w.WriteBits32('B', 8)
	w.WriteBits32('Z', 8)
	w.WriteBits32('h', 8)
	w.WriteBits32('9', 8) // FIXME: Should be '1'-'9'

	// Write blocks
	for block := range blocks {
		fmt.Println("Waiting for block...")
		block.Wait() // Wait for the block to be ready
		fmt.Println("Block obtained")
		block.writeTo(w)
		fmt.Println("Wrote block")
	}

	// Write finalizer ?

	done <- struct{}{}
	fmt.Println("Signalled doneness")
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
func (w *Writer) Close() error {
	close(w.sendTo)
	fmt.Println("Closing")
	<-w.closed
	fmt.Println("Closed")
	w.w.Close()
	return nil
}

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
