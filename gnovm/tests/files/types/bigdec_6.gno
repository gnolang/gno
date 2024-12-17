package main

func main() {
	println(1 % 1e9)
}

// Error:
// main/files/types/bigdec_6.gno:4:10: operator % not defined on: BigdecKind
