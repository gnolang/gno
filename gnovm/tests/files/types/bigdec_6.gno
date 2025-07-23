package main

func main() {
	println(1 % 1e9)
}

// Error:
// main/bigdec_6.gno:4:10-17: operator % not defined on: BigdecKind

// TypeCheckError:
// main/bigdec_6.gno:4:10: invalid operation: operator % not defined on 1 (untyped float constant)
