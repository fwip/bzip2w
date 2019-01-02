package bitwriter

import (
	"errors"
	"fmt"
	"io"
)

const capacity = 1024

//type Writer interface {
//WriteBit(b byte) (err error)
//WriteBits(b byte, count int) (n int, err error)
//Close() (err error)
//}

// Writer writes to an io.Writer, one bit at a time.
type Writer struct {
	w             io.Writer
	cache         []byte
	bitsUnflushed uint
	closed        bool
}

// NewWriter creates a new bitwriter
func NewWriter(w io.Writer) *Writer {

	return &Writer{w: w, cache: make([]byte, capacity)}
}

// WriteBit writes a single bit
func (w *Writer) WriteBit(b byte) (err error) {
	if w.closed {
		return errors.New("Can't write to a closed file")
	}
	if w.bitsUnflushed == capacity {
		err := w.flush()
		if err != nil {
			return err
		}
	}
	idx := w.bitsUnflushed >> 3         // Divide by 8 to get byte index
	bitPos := w.bitsUnflushed & 7       // Only retain lower three bits
	w.cache[idx] |= (b << (7 - bitPos)) // TODO: Is this right? Little-endian vs big-endian
	w.bitsUnflushed++

	return nil
}

// WriteBits32 writes up to 32 bits at once
func (w *Writer) WriteBits32(b uint32, count uint) (n int, err error) {
	if count > 32 {
		return 0, fmt.Errorf("You can't stuff %d bits in an int32", count)
	}
	//TODO: Faster implementation
	//TODO: Should this be reversed?
	for ; count > 0; count-- {
		w.WriteBit(byte((b >> (count - 1)) & 1))
	}
	return 0, nil
}

// This is probably overly-complex since flush() is only called when the cache is full
func (w *Writer) flush() (err error) {
	if w.bitsUnflushed == 0 {
		return nil
	}

	w.bitsUnflushed = w.bitsUnflushed & 7
	var tmp byte
	if w.bitsUnflushed != 0 { // If not an even number of bytes...?
		tmp = w.cache[len(w.cache)-1] // pop off the last bit
		w.cache = w.cache[:len(w.cache)-1]
	}

	n, err := w.w.Write(w.cache[:])

	w.cache = w.cache[:0]
	if w.bitsUnflushed != 0 {
		w.cache[0] = tmp
	}

	if n < len(w.cache) || err != nil {
		return err
	}

	return nil
}

// Close adds padding and prevents any more bits from being written
func (w *Writer) Close() (err error) {
	_, err = w.w.Write(w.cache)
	w.closed = true
	fmt.Println("Wrote")
	return err
}
