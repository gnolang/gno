package main

func main() {
	a := 1
	r := a << "hello" // NOTE: go vet would fail, but still process
	println(r)
}

// Error:
// main/debug/10a11_filetest.gno:3: cannot convert StringKind to UintKind
