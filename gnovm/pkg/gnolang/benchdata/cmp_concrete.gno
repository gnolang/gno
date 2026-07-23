package main

// Concrete (non-interface) == hot path: 8 comparisons of int variables per
// iteration. Variant deltas isolate the dispatch cost of doOpEql; loop/if/
// counter overhead is identical across variants.
func main() {
	a := 0
	b := 0
	n := 0
	for i := 0; i < {{ .N }}; i++ {
		if a == b {
			n++
		}
		if a == b {
			n++
		}
		if a == b {
			n++
		}
		if a == b {
			n++
		}
		if a == b {
			n++
		}
		if a == b {
			n++
		}
		if a == b {
			n++
		}
		if a == b {
			n++
		}
		a = i
		b = i
	}
	_ = n
}
