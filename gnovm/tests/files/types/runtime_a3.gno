package main

func gen() interface{} {
	return false
}

func main() {
	gen()++
}

// Error:
// main/files/types/runtime_a3.gno:8: operator ++ not defined on: InterfaceKind
