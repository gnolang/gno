package main

type S struct {
	i int
}

func main() {
	// Can't take the address of non-addressable member of
	// a non-addressable struct.
	_ = &S{i: 4}.i
}

// Error:
// main/files/addressable_3a_err.gno:10:6: cannot take address of S<VPBlock(3,0)>{i<VPField(0,0,i)>: (const (4 int))}.i
