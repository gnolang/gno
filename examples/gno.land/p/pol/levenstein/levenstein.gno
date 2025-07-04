// from www.youtube.com/watch?v=d-Eq6x1yssU
// Levenstein distance is an algorithm to detect how close 2 words are.
// For example, "Float" and "Boats" have a levenstein distance of 3.
// Why ? Because you need 3 steps to go from Boats to Float
// - We start with the word (Boats)
// - Remove the s (Boat)
// - Insert the l (Bloat)
// - Substitute b with f (Float)
package levenstein

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
	prev := make([]uint, len(b)+1)
	curr := make([]uint, len(b)+1)

	// init the first row
	for j := range prev {
		prev[j] = uint(j)
	}

	// main loop
	for i := 1; i <= len(a); i++ {
		curr[0] = uint(i)
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1]
			} else {
				curr[j] = min3(
					curr[j-1]+1, // insertion
					prev[j]+1,   // deletion
					prev[j-1]+1, // substitution
				)
			}
		}
		// swap curr and prev
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

type Levensteinable interface {
	LString() string
}

// Pick the best object that is the closest using
// levenstein string distance with the original
// string.
// O(n) complexity.
func PickFromString(xStr string, xs []Levensteinable) (Levensteinable, uint) {
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
func Pick(x Levensteinable, xs []Levensteinable) (Levensteinable, uint) {
	return PickFromString(x.LString(), xs)
}

func padLeftNum(n uint, width int) string {
	s := strconv.FormatUint(uint64(n), 10)
	for len(s) < width {
		s = "0" + s
	}
	return s
}

// Pick the n best elements in a list of Levensteinable elements, using
// the levenstein distance between the original string.
// O(n) complexity.
func SeveralPickFromString(xStr string, xs []Levensteinable, n uint) []Levensteinable {
	// error casing
	if len(xs) == 0 {
		return nil
	}
	if n == 0 {
		return []Levensteinable{}
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

	// creqte the list
	out := make([]Levensteinable, 0, tree.Size())
	tree.Iterate("", "", func(k string, v any) bool {
		out = append(out, v.(Levensteinable))
		return false
	})
	return out
}

// Pick the n best elements in a list of Levensteinable elements, using
// the levenstein distance between the original object's string.
// O(n) complexity.
func SeveralPick(x Levensteinable, xs []Levensteinable, n uint) []Levensteinable {
	return SeveralPickFromString(x.LString(), xs, n)
}
