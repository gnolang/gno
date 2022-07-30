package main

type T int

func (t T) Error() string { return "T: error" }

var invalidT T

func main() {
	var err error
	if err > invalidT {
		println("ok")
	}
}

// Error:
// main/files/op7.gno:11: incompatible types in binary expression: err<VPBlock(2,0)> GTR invalidT<VPBlock(4,2)>