package bzip2

import (
	"fmt"
	"sort"
	"sync"

	bit "github.com/fwip/bzip2w/bit"
	huffman "github.com/fwip/bzip2w/huffman"
)

const (
	bzip2BlockMagic = 0x314159265359 // BCD pi
)

type blockEncoder struct {
	input    []byte
	capacity int
	trees    []huffman.Book

	sync.WaitGroup
	output []byte
}

func (e *blockEncoder) Write(in []byte) (n int, err error) {
	//if e == nil {
	//return 0, errors.New("can't write to nil encoder")
	//}
	if len(in)+len(e.input) > e.capacity {
		n = e.capacity - len(e.input)
		e.input = append(e.input, in[:n]...)
		return n, nil
	}
	e.input = append(e.input, in...)
	return len(in), nil
}

func newBlockEncoder(capacity uint) *blockEncoder {
	return &blockEncoder{
		input:    make([]byte, 0, capacity),
		capacity: int(capacity),
	}
}

var _ = huffman.Book{}

// Transform input into encoded output
// This shouldn't ever error.
func (e *blockEncoder) encode() {
	//step1 := rle(e.input)
	step2 := bwt(e.input)
	used, step3 := mtf(step2)
	step4 := rleMTF(step3)

	fmt.Println(used)
	fmt.Println(step4)
	fmt.Println("Done encoding!")
	return
}

func (e *blockEncoder) writeTo(w *bit.Writer) {

	// compressed_magic:48            = 0x314159265359 (BCD (pi))
	w.WriteBits32(bzip2BlockMagic>>16, 32)
	w.WriteBits32(bzip2BlockMagic&((1<<17)-1), 16)
	// .crc:32                         = checksum for this block
	w.WriteBits32(1<<32-1, 32) //TODO
	// .randomised:1                   = 0=>normal, 1=>randomised (deprecated)
	w.WriteBit(0)
	// .origPtr:24                     = starting pointer into BWT for after untransform
	// .huffman_used_map:16            = bitmap, of ranges of 16 bytes, present/not present
	// .huffman_used_bitmaps:0..256    = bitmap, of symbols used, present/not present (multiples of 16)
	// .huffman_groups:3               = 2..6 number of different Huffman tables in use
	// .selectors_used:15              = number of times that the Huffman tables are swapped (each 50 bytes)
	// *.selector_list:1..6            = zero-terminated bit runs (0..62) of MTF'ed Huffman table (*selectors_used)
	// .start_huffman_length:5         = 0..20 starting bit length for Huffman deltas
	// *.delta_bit_length:1..40        = 0=>next symbol; 1=>alter length { 1=>decrement length; 0=>increment length } (*(symbols+2)*groups)
	// .contents:2..∞                  = Huffman encoded data stream until end of block (max. 7372800 bit)

	fmt.Println("Finished writing!")
}

// TODO: Doesn't handle runs of 256 or more
// TODO: Super inefficient (?)
// rle = run-length encoding
// This encoding limits the max repeat length to 4
// Sequences of 4 or more are represented by AAAAB, where A is the byte value,
// and B is one byte representing the number (0-255) of additional characters
// of that value
func rle_tmp(in []byte) (out []byte) {
	var count byte = 1
	for i := 1; i < len(in); i++ {
		if in[i] == in[i-1] {
			count++
		}
		if in[i] != in[i-1] {
			for j := 0; j < 4 && j < int(count); j++ {
				out = append(out, in[i-1])
			}
			if count >= 4 {
				out = append(out, count-4)
			}
			count = 1
		}
	}
	for j := 0; j < 4 && j < int(count); j++ {
		out = append(out, in[len(in)-1])
	}
	if count >= 4 {
		out = append(out, count-4)
	}
	return out
}

// TODO: This is super slow for large inputs
// bwt = burrows-wheeler transform
// This is the meat of the compression algorithm
func bwt(in []byte) (out []byte) {
	matrix := make([]string, len(in))
	for i := range in {
		matrix[i] = string(in[len(in)-i:]) + string(in[:len(in)-i])
	}
	sort.Strings(matrix)
	for i := range in {
		out = append(out, matrix[i][len(in)-1])
	}
	return out
}

// mtf = move-to-front transform
// TODO: Likely slow
func mtf(in []byte) (used [256]bool, out []byte) {

	for _, c := range in {
		used[c] = true
	}

	frontlist := [256]byte{}
	var count int
	for i := 0; i < 256; i++ {
		if used[i] {
			frontlist[count] = byte(i)
			count++
		}
	}

	// Walk the input string
	for _, c := range in {
		// Find the character in the list
		for i, d := range frontlist {
			if c == d {
				// Update the list
				copy(frontlist[1:i+1], frontlist[:i])
				frontlist[0] = d
				out = append(out, byte(i))
				break
			}
		}
	}

	return used, out
}

// This encodes runs of zeroes specially (RUNA=0, RUNB=1)
// And adds 1 to everything else
// This should probably be a part of mtf, to be honest
func rleMTF(in []byte) (out []uint16) {
	var count uint16
	var place uint16
	for _, c := range in {
		if c == 0 {
			count++
		} else {
			for place = 1; count > 0; place <<= 1 {
				if count&place != 0 {
					count -= place
					out = append(out, runA)
				} else {
					count -= place * 2
					out = append(out, runB)
				}
			}
			if count != 0 {
				panic("Count should definitely be zero")
			}
			out = append(out, uint16(c)+1)
		}
	}
	// In case we end with zeroes
	for place = 1; count > 0; place <<= 1 {
		if count&place != 0 {
			count -= place
			out = append(out, runA)
		} else {
			count -= place * 2
			out = append(out, runB)
		}
	}
	return out
}
