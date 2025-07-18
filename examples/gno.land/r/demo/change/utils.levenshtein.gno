package change

// from this pr: gnolang/gno/pull/4472
// from p/pol/levenshtein
// since its not merged, i copy pasted here

import (
	"strconv"

	"gno.land/p/demo/avl"
)

func min3(a, b, c uint) uint {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// Return the levenstein distance between 2 strings
func Distance(a, b string) uint {
	// make sure a is the longest
	if len(a) < len(b) {
		a, b = b, a
	}

	// init first row
	line := make([]uint, len(b)+1)
	for j := range line {
		line[j] = uint(j)
	}

	// main loop
	for i := 1; i <= len(a); i++ {
		prev := line[0]
		line[0] = uint(i)
		for j := 1; j <= len(b); j++ {
			old := line[j]
			if a[i-1] == b[j-1] {
				line[j] = prev
			} else {
				line[j] = min3(
					line[j-1]+1, // insertion
					line[j]+1,   // deletion
					prev+1,      // substitution
				)
			}
			prev = old
		}
	}
	return line[len(b)]
}

// The general interface for levenshteinable string.
// You need to use LString, because String is used to convert your object into a string.
// LString is only when you want to compare the results.
type Levenshteinable interface {
	LString() string
}

// Pick the best object that is the closest using
// levenstein string distance with the original
// string.
// O(n) complexity.
func PickFromString(xStr string, xs []Levenshteinable) (Levenshteinable, uint) {
	// check for empty list
	if len(xs) == 0 {
		return nil, 0
	}

	// initialize
	bestDistance := Distance(xStr, xs[0].LString())
	best := xs[0]

	// loop through objects
	for n := 1; n < len(xs); n++ {
		dist := Distance(xStr, xs[n].LString())
		if dist < bestDistance {
			bestDistance = dist
			best = xs[n]
		}
	}
	return best, bestDistance
}

// Pick the best object that is the closest using levenstein string
// distance with the original string object.
// O(n) complexity.
func Pick(x Levenshteinable, xs []Levenshteinable) (Levenshteinable, uint) {
	return PickFromString(x.LString(), xs)
}

func padLeftNum(n uint, width int) string {
	s := strconv.FormatUint(uint64(n), 10)
	for len(s) < width {
		s = "0" + s
	}
	return s
}

// Pick the n best elements in a list of Levenshteinable elements, using
// the levenstein distance between the original string.
// O(n) complexity.
func SeveralPickFromString(xStr string, xs []Levenshteinable, n uint) []Levenshteinable {
	// error casing
	if len(xs) == 0 {
		return nil
	}
	if n == 0 {
		return []Levenshteinable{}
	}

	tree := avl.NewTree()
	maxKey := ""

	for k, v := range xs {
		// calculate distance and create unique key
		d := Distance(xStr, v.LString())
		// ufmt.Sprintf("%016u:%d", d, k)
		key := padLeftNum(d, 6) + ":" + strconv.Itoa(k) // no sprintf ( ;-;)
		tree.Set(key, v)

		// remove the element with the greater key
		if uint(tree.Size()) > n {
			tree.ReverseIterate("", "", func(k string, _ any) bool {
				maxKey = k
				return true
			})
			tree.Remove(maxKey)
		}
	}

	// create the list
	out := make([]Levenshteinable, 0, tree.Size())
	tree.Iterate("", "", func(k string, v any) bool {
		out = append(out, v.(Levenshteinable))
		return false
	})
	return out
}

// Pick the n best elements in a list of Levenshteinable elements, using
// the levenstein distance between the original object's string.
// O(n) complexity.
func SeveralPick(x Levenshteinable, xs []Levenshteinable, n uint) []Levenshteinable {
	return SeveralPickFromString(x.LString(), xs, n)
}