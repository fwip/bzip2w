package bzip2

import (
	"fmt"
	"sort"
	"sync"
)
import huffman "github.com/fwip/bzip2w/huffman"

type blockEncoder struct {
	input    []byte
	capacity int
	trees    []huffman.Book

	sync.Mutex
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
	e.Lock()
	defer e.Unlock()
	//step1 := rle(e.input)
	step2 := bwt(e.input)
	used, step3 := mtf(step2)
	step4 := rleMTF(step3)

	fmt.Println(used)
	fmt.Println(step4)
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
