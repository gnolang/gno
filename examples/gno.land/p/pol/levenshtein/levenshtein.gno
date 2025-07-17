// from www.youtube.com/watch?v=d-Eq6x1yssU
// Levenshtein distance is an algorithm to detect how close 2 words are.
// For example, "Float" and "Boats" have a levenstein distance of 3.
// Why ? Because you need 3 steps to go from Boats to Float
// - We start with the word (Boats)
// - Remove the s (Boat)
// - Insert the l (Bloat)
// - Substitute b with f (Float)
package levenshtein

import (
	"strconv"

	"gno.land/p/demo/avl"
)

func min3(a, b, c uint) uint {
	min := a
	if b < min {
		min = b
	}
	if c < min {
		min = c
	}
	return min
}

// Distance computes the Levenshtein distance between two strings.
//
// The Levenshtein distance is a measure of the similarity between two strings,
// defined as the minimum number of single-character edits (insertions,
// deletions, or substitutions) required to change one string into the other.
//
// For example, the Levenshtein distance between "Float" and "Boats" is 3:
//   - Remove 's' from "Boats" → "Boat"
//   - Insert 'l' → "Bloat"
//   - Substitute 'b' with 'f' → "Float"
//
// This implementation uses a memory-efficient dynamic programming approach
// that computes the distance in O(len(a) * len(b)) time and O(min(len(a), len(b))) space.
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
