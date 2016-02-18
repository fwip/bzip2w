package bzip2

import (
	"math/rand"
	"testing"
)

var bookTests = []struct {
	freq     []int
	expected Book
}{
	{[]int{10, 5, 2, 1}, Book{[]Code{{0, 1}, {2, 2}, {6, 3}, {7, 3}}}},
	{[]int{1000, 6, 5, 10, 1}, Book{[]Code{{0, 1}, {6, 3}, {14, 4}, {2, 2}, {15, 4}}}},
}

func TestNewBook(t *testing.T) {
	for _, test := range bookTests {
		book := NewBook(test.freq)

		t.Log(book)
		if len(book.Codes) != len(test.freq) {
			t.Errorf("Expected %d codes, got %d", len(test.freq), len(book.Codes))
			t.FailNow()
		}
		for i := range book.Codes {
			if test.expected.Codes[i].bits != book.Codes[i].bits {
				t.Errorf("NewBook(%v).Codes[%d] => %s, want %s", test.freq, i, book.Codes[i], test.expected.Codes[i])
			}
		}
	}
	/*
		in := make([]int, 258)
		for i := 0; i < len(in); i++ {
			in[i] = rand.Intn(3000*len(in) - i)
		}
		in[0] = 100000000
		in[len(in)-1] = 1

		book := NewBook(in)
		t.Log("\n" + book.String())
		if len(book.Codes) != len(in) {
			t.Errorf("Expected %d codes, got %d :(", len(in), len(book.Codes))
		}
	*/
}

func BenchmarkNewBook(b *testing.B) {
	in := make([]int, 258)
	for i := 0; i < len(in); i++ {
		//in[i] = rand.Intn(3000*len(in) - i)
		in[i] = rand.Intn(3000)
	}
	//in[0] = 100000000
	//in[len(in)-1] = 1

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		NewBook(in)
	}
}

/*
func TestBuildTree(t *testing.T) {
	input := []int{50, 42, 31, 10}
	tree := buildTree(input)
	//t.Log(tree)
	codes := tree.getCodes()
	//t.Log(codes)
	//for val, code := range codes {
	//t.Log(val, code)
	//}
}

func TestGetCodes(t *testing.T) {
	freq := make([]int, 258)
	for i := range freq {
		freq[i] = rand.Intn(300)
	}
	//freq[0] = 10000000
	freq[257] = 1
	freq[64] = 0

	codes := buildTree(freq).getCodes()
	if len(freq) != len(codes) {
		t.Errorf("Expected %d codes, got %d??", len(freq), len(codes))
	}
	maxLen := 0
	for v, code := range codes {
		t.Log(v, code)
		if maxLen < len(code) {
			maxLen = len(code)
		}

	}
	t.Log("Max:", maxLen)
}

func BenchmarkBuildTree(b *testing.B) {
	freq := make([]int, 258)
	for i := range freq {
		freq[i] = rand.Intn(30000)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildTree(freq)
	}
	b.ReportAllocs()
}
*/
