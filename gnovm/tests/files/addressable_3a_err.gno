package main

type S struct {
	i int
}

func main() {
	// Can't take the address of non-addressable member of
	// a non-addressable struct.
	_ = &S{i: 4}.i
}

// TypeCheckError:
// main/files/addressable_3a_err.gno:10:7: invalid operation: cannot take address of S{…}.i (value of type int)
