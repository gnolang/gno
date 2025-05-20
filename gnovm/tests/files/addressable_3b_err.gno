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
// main/files/addressable_3b_err.gno:8:7: invalid operation: cannot take address of getStruct().i (value of type int)
