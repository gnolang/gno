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

// Error:
// main/files/addressable_3b_err.gno:8:6: cannot take address of getStruct<VPBlock(3,2)>().i
