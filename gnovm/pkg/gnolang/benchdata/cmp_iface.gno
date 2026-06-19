package main

// Interface == path: 8 comparisons of interface variables per iteration.
// Operands are statically interface-typed, so this exercises the
// isInterfaceCmp/OpEqlIface path (uncomparable-dynamic-type check applies;
// both hold comparable ints, so no panic). Variant deltas isolate the
// dispatch + interface-check cost end-to-end.
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
