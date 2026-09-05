package main

type S struct {
	i int
}

func main() {
	_ = &getStruct().i
}

func getStruct() S {
	return S{i: 9}
}

// TypeCheckError:
// main/addressable_3b_err.gno:8:7: invalid operation: cannot take address of getStruct().i (value of type int)

// Error:
// main/addressable_3b_err.gno:8:6-20: invalid operation: cannot take address of getStruct<VPBlock(3,2)>().i (value of type int)
