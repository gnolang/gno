package main

func main() {
	// Iterator syntax in latest go versions. Currently unsupported in Gno.
	p := func(yield func(v int) bool) {
		yield(123)
	}
	for j := range p {
		println(j)
	}
}

// Error:
// main/range12.gno:8:2-10:3: range iteration requires map, string, array, slice, or pointer to array; got FuncKind
