package main

func main() {
	println(1 > "a")
}

// Error:
// main/debug/binary_1b.gno:4: cannot use untyped Bigint as StringKind
