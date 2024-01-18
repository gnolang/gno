package main

func gen() interface{} {
	return false
}

func main() {
	gen()++
}

// Error:
// main/files/types/26a3_filetest.gno:8: operator ++ not defined on: InterfaceKind
