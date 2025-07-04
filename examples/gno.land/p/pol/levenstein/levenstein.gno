// from www.youtube.com/watch?v=d-Eq6x1yssU
// Levenstein distance is an algorithm to detect how close 2 words are.
// For example, "Float" and "Boats" have a levenstein distance of 3.
// Why ? Because you need 3 steps to go from Boats to Float
// - We start with the word (Boats)
// - Remove the s (Boat)
// - Insert the l (Bloat)
// - Substitute b with f (Float)
package levenstein

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
					curr[j-1]+1,   // insertion
					prev[j]+1,     // deletion
					prev[j-1]+1,   // substitution
				)
			}
		}
		// swap curr and prev
		prev, curr = curr, prev
	}
	return prev[len(b)]
}
