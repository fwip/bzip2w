package bzip2

import "testing"

func TestRle(t *testing.T) {
	input := []byte("AAAAAAABBBBCCCDEE")
	expected := []byte{'A', 'A', 'A', 'A', 3, 'B', 'B', 'B', 'B', 0, 'C', 'C', 'C', 'D'}
	expectedLeftover := []byte{'E', 'E'}
	output, leftover := rle(input)
	if string(output) != string(expected) {
		t.Errorf("rle: Gave %s, expected:\n%v\nGot:\n%v\n", input, expected, output)
	}
	if string(output) != string(expected) {
		t.Errorf("rle_leftover: Gave %s, expected leftover:\n%v\nGot:\n%v\n", input, expectedLeftover, leftover)
	}

}

func TestBwt(t *testing.T) {
	input := []byte("^BANANA|")
	expected := []byte("BNN^AA|A")
	output := bwt(input)
	if string(output) != string(expected) {
		t.Errorf("\nbwt: Gave %s, expected:\n%v\nGot:\n%v (%s)\n", input, expected, output, output)
	}
}

func TestMtf(t *testing.T) {

	input := []byte("bananaaa")
	expected := []byte{1, 1, 2, 1, 1, 1, 0, 0}
	_, output := mtf(input)
	if string(output) != string(expected) {
		t.Errorf("\nmtf: Gave %v, expected:\n%v\nGot:\n%v\n", input, expected, output)
	}
}

func TestRleMTF(t *testing.T) {
	input := []byte{0, 0, 0, 0, 0, 1, 0}
	expected := []uint16{runA, runB, 2, runA}
	output := rleMTF(input)
	if len(output) != len(expected) {
		t.Errorf("\nrle_mtf: Gave %v, expected:\n%v\nGot:\n%v\n", input, expected, output)
	}
	for i := range expected {
		if output[i] != expected[i] {
			t.Errorf("\nrle_mtf: Gave %v, expected:\n%v\nGot:\n%v\n", input, expected, output)
		}
	}

}

/*
func TestMagicNumber(t *testing.T) {
	if bytes.Equal(bzipMagicNumber, []byte{'B', 'Z'}) {
		t.Errorf("Magic number should be BZ, not %v\n", bzipMagicNumber)
	}
}
*/
