package bzip2

import (
	"bytes"
	"fmt"
	"sort"
)

type Book struct {
	Codes []Code
}
type Code struct {
	val  uint32
	bits byte
}

func (b Book) String() string {
	stuff := bytes.Buffer{}
	for i, c := range b.Codes {
		stuff.WriteString(fmt.Sprintf("%3[1]d: %s\n", i, c))
	}
	return stuff.String()
}
func (c Code) String() string {
	return fmt.Sprintf("%0[2]*[1]b", c.val, int(c.bits))
}

func (nl nodeList) search(freq int) int {
	min := 0
	max := len(nl) - 1
	test := 0
	for min < max {
		test = max + min/2
		testn := nl[test].freq
		switch {
		case testn == freq:
			return test
		case testn < freq:
			min = test + 1
		case testn > freq:
			max = test - 1
		}
	}
	return test
}

func (n node) String() (out string) {
	if n.children == nil {
		return fmt.Sprintf("%d~%d", n.freq, n.val)
	}
	return fmt.Sprintf("%d(%s %s)", n.freq, n.children[0], n.children[1])
}

// Warning; Destructive! Modifies the input slice.
func buildTree(nodes []node) node {
	sort.Sort(nodeList(nodes))
	for len(nodes) > 1 {
		n := combineNodes(nodes[0], nodes[1])
		switch {
		case len(nodes) == 2:
			nodes[1] = n
		case n.freq <= nodes[2].freq:
			nodes[1] = n
		case n.freq >= nodes[len(nodes)-1].freq:
			copy(nodes[1:len(nodes)-1], nodes[2:])
			nodes[len(nodes)-1] = n
		default:
			idx := nodeList(nodes).search(n.freq)
			copy(nodes[:idx], nodes[1:idx+1])
			nodes[idx] = n

		}
		nodes = nodes[1:]
	}

	return nodes[0]
}

// Slower but simpler implementation
func buildTreeSlowly(nodes []node) node {
	for len(nodes) > 1 {
		sort.Sort(nodeList(nodes)) // TODO: This is dang slow.
		n := combineNodes(nodes[0], nodes[1])
		nodes = nodes[1:]
		nodes[0] = n
	}
	return nodes[0]

}

// Takes a node (tree) and a length (number of leaves in the tree) and returns
// a new Book with the canonical encodings
func genBookFromTree(n node, length int) Book {
	var book Book
	book.Codes = make([]Code, length)
	levels := make([][]node, 0, 20)
	levels = append(levels, []node{n})

	var count int32 = -1
	bitLength := 0
	for i := 0; i < len(levels); i++ {
		sort.Sort(byVal(levels[i]))
		for _, n := range levels[i] {
			if n.children != nil {
				if len(levels) == i+1 {
					//levels = append(levels, nil)
					levels = append(levels, make([]node, 0, 2*len(levels[len(levels)-1])))
				}
				levels[i+1] = append(levels[i+1], n.children[0])
				levels[i+1] = append(levels[i+1], n.children[1])
			} else {
				count = (count + 1) << uint(i-bitLength)
				bitLength = i
				book.Codes[n.val] = Code{bits: byte(i), val: uint32(count)}
			}
		}
	}

	return book
}

func NewBook(freq []int) Book {
	// Generate the nodes
	nodes := make([]node, 0, len(freq))
	for v, f := range freq {
		nodes = append(nodes, node{val: v, freq: f})
	}

	// Order the nodes
	n := buildTree(nodes)

	// TODO: Enforce maximum length of 17 (or 20, TBD)

	// Canonicalize the tree
	book := genBookFromTree(n, len(freq))

	return book
}

type node struct {
	val      int
	freq     int
	children []node
}

type nodeList []node

func (nl nodeList) Len() int      { return len(nl) }
func (nl nodeList) Swap(a, b int) { nl[a], nl[b] = nl[b], nl[a] }
func (nl nodeList) Less(a, b int) bool {
	return nl[a].freq < nl[b].freq
}

type byVal []node

func (nl byVal) Len() int      { return len(nl) }
func (nl byVal) Swap(a, b int) { nl[a], nl[b] = nl[b], nl[a] }
func (nl byVal) Less(a, b int) bool {
	return nl[a].val < nl[b].val
}

func combineNodes(a, b node) node {
	return node{freq: a.freq + b.freq, children: []node{a, b}}
}

/*
func (n node) canonicalize() (bits []byte) {
	bits = make([]byte, 258)
	levels := [][]node{[]node{n}}
	for i := 0; i < len(levels); i++ {
		for _, n := range levels[i] {
			if n.children != nil {
				levels[i+1] = append(levels[i+1], n.children[0])
				levels[i+1] = append(levels[i+1], n.children[1])
			} else {
				bits[n.val] = byte(i)
			}
		}
	}

	return bits
}

func (n node) getCodes() (out map[int]string) {
	out = make(map[int]string)
	if n.children == nil {
		out[n.val] = ""
		return out
	}
	for val, code := range n.children[0].getCodes() {
		out[val] = "0" + code
	}
	for val, code := range n.children[1].getCodes() {
		out[val] = "1" + code
	}
	return out
}


func freqs(nodes []node) (freqs []int) {
	for _, n := range nodes {
		freqs = append(freqs, n.freq)
	}
	return freqs
}

func buildTree(freq []int) node {
	nodes := make([]node, 0, len(freq))
	for v, f := range freq {
		nodes = append(nodes, node{val: v, freq: f})
	}
*/

//
//// Something wrong with this
//	sort.Sort(nodeList(nodes))
//	for len(nodes) > 1 {
//		n := combineNodes(nodes[0], nodes[1])
//
//		min := 2
//		max := len(nodes) - 1
//		var test int
//		switch {
//		case len(nodes) == 2 || n.freq <= nodes[2].freq:
//			nodes[1] = n
//		case n.freq >= nodes[len(nodes)-1].freq:
//			copy(nodes[1:len(nodes)-1], nodes[2:])
//			nodes[len(nodes)-1] = n
//
//		default:
//			for min < max {
//				test = (max + min) / 2
//				tf := nodes[test].freq
//				if tf == n.freq {
//					break
//				}
//				if tf < n.freq {
//					min = test + 1
//					continue
//				}
//				max = test - 1
//			}
//		}
//		nodes = nodes[1:]
//
//		//if !sort.IsSorted(nodeList(nodes)) {
//		//panic("Not sorted")
//		//}
//
//	}
//
/*
	for len(nodes) > 1 {
		sort.Sort(nodeList(nodes)) // TODO: This is slow.
		n := combineNodes(nodes[0], nodes[1])
		nodes = nodes[1:]
		nodes[0] = n
	}

	return nodes[0]
}
*/
