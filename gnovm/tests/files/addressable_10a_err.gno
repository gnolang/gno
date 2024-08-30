package main

func main() {
	println(&getPtr())
}

type S struct{}

func getPtr() *S {
	return &S{}
}

// Error:
// main/files/addressable_10a_err.gno:4:10: cannot take address of getPtr<VPBlock(3,2)>()
