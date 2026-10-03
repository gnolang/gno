package main

// Interface == path: 8 comparisons of interface variables per iteration.
// Operands are statically interface-typed, so doOpEql takes the
// ATTR_IFACE_CMP branch (both hold comparable ints, so no panic).
func main() {
	var a interface{} = 0
	var b interface{} = 0
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
