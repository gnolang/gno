package main

func main() {
	println(&getPtr())
}

type S struct{}

func getPtr() *S {
	return &S{}
}

// Error:
// main/addressable_10a_err.gno:4:10-19: invalid operation: cannot take address of getPtr<VPBlock(3,2)>() (value of type *main.S)

// TypeCheckError:
// main/addressable_10a_err.gno:4:11: invalid operation: cannot take address of getPtr() (value of type *S)
