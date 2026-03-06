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

	"gno.land/p/nt/avl/v0"
)

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

// Levenshteinable is an interface for types that can be compared using
// Levenshtein distance. It provides the LString method, which returns
// the string representation used for comparison.
//
// Note: This differs from the standard String() method, which may be used
// for display or debugging.
type Levenshteinable interface {
	LString() string
}

// PickFromString returns the Levenshteinable object from the slice xs
// that is closest in Levenshtein distance to the input string xStr.
//
// It returns the best match along with its distance. If xs is empty,
// it returns (nil, 0).
//
// This function runs in O(n) time, where n is the length of the slice.
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

// Pick returns the Levenshteinable object from the slice xs that is
// closest in Levenshtein distance to the input object x.
//
// This is a convenience wrapper around PickFromString(x.LString(), xs).
// It runs in O(n) time.
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

// SeveralPickFromString returns the n closest Levenshteinable objects
// to the input string xStr from the slice xs, using Levenshtein distance.
//
// If xs is empty, it returns nil. If n is 0, it returns an empty slice.
//
// The result is not guaranteed to be sorted by distance.
// This function runs in O(n) time using an AVL tree for efficient selection.
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

// SeveralPick returns the n closest Levenshteinable objects to the
// input object x from the slice xs, using Levenshtein distance.
//
// This is a convenience wrapper around SeveralPickFromString(x.LString(), xs, n).
// It runs in O(n) time.
func SeveralPick(x Levenshteinable, xs []Levenshteinable, n uint) []Levenshteinable {
	return SeveralPickFromString(x.LString(), xs, n)
}

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
