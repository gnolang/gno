package main

// An oversized ellipsis array [...]T{idx: v} is sized by its largest keyed
// index, known only after the elements are measured. It must be rejected at
// compile time too (Go: "larger than address space"), not leak to a runtime
// allocation panic like the explicit [N]T forms in make20/make21.

func main() {
	x := [...]int{9223372036854775806: 1}
	_ = x
}

// Error:
// main/make22.gno:9:7-39: type [9223372036854775807]int larger than address space
